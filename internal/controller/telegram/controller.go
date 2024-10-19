package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/service/postpone"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/internal/worker"
	"go-ton-pass-telegram-bot/pkg/logger"
)

type BotController interface {
	Serve(update *telegram.Update) error
}

const (
	selectPreferredLanguageImageURL = "https://i.ibb.co/4g08xxg/language.png"
	selectPreferredCurrencyImageURL = "https://i.ibb.co/NSfjN7Y/currency.png"
	avatarImageURL                  = "https://i.ibb.co/rmqsKty/avatar.png"
	welcomeImageURL                 = "https://i.imghippo.com/files/vi44s1726518102.png"
	enterAmountImageURL             = "https://i.ibb.co/g97tpZj/amount.png"
	topUpImageURL                   = "https://i.ibb.co/F0WKPfR/topup.png"
	helpImageURL                    = "https://i.ibb.co/gyMDN5t/help.png"
	historyImageURL                 = "https://i.ibb.co/Gf6QMCG/history.png"
	chooseCountryImageURL           = "https://i.ibb.co/VSBbV14/country.png"
	chooseServiceImageURL           = "https://i.ibb.co/m4KYq4n/service.png"
	failReceivedCodeImageURL        = "https://www.imghippo.com/i/D8eey1728514326.png"
)

type botController struct {
	container            container.Container
	telegramBotService   service.TelegramBotService
	cryptoPayBot         service.CryptoPayBot
	sessionService       service.SessionService
	cacheService         service.Cache
	smsService           service.SMSService
	postponeService      postpone.Postpone
	profileRepository    repository.ProfileRepository
	smsHistoryRepository repository.SMSHistoryRepository
	exchangeRateWorker   worker.ExchangeRate
	smsActivateWorker    worker.SMSActivate
	formatterWorker      worker.Formatter
}

func NewBotController(
	container container.Container,
	sessionService service.SessionService,
	cacheService service.Cache,
	smsService service.SMSService,
	postponeService postpone.Postpone,
	profileRepository repository.ProfileRepository,
	smsHistoryRepository repository.SMSHistoryRepository,
) BotController {
	cryptoPayBot := service.NewCryptoPayBot(container)
	exchangeRateWorker := worker.NewExchangeRate(container, cacheService, cryptoPayBot)
	smsActivateWorker := worker.NewSMSActivate(container, smsService, cacheService)
	formatterWorker := worker.NewFormatter(container)
	return &botController{
		container:            container,
		telegramBotService:   service.NewTelegramBot(container),
		cryptoPayBot:         cryptoPayBot,
		sessionService:       sessionService,
		cacheService:         cacheService,
		smsService:           smsService,
		postponeService:      postponeService,
		profileRepository:    profileRepository,
		smsHistoryRepository: smsHistoryRepository,
		exchangeRateWorker:   exchangeRateWorker,
		smsActivateWorker:    smsActivateWorker,
		formatterWorker:      formatterWorker,
	}
}

func (b *botController) Serve(update *telegram.Update) error {
	log := b.container.GetLogger()
	log.Debug("receive telegram message", logger.F("update", update))
	if isEmpty(update) {
		return app.EmptyUpdateError
	}

	ctx := context.Background()
	telegramID, err := getTelegramID(update)
	if err != nil {
		log.Error("can't get telegram id", logger.FError(err))
		return err
	}
	if err := b.recordTelegramProfile(ctx, update); err != nil {
		log.Error("recordTelegramProfile has failed", logger.FError(err))
		return err
	}

	telegramCmd, err := b.telegramBotService.ParseTelegramCommand(update)
	switch telegramCmd {
	case app.StartTelegramCommand:
		return b.startTelegramCommandHandler(ctx, update)
	case app.HelpTelegramCommand:
		return b.helpTelegramCommandHandler(ctx, update)
	}
	if err != nil && telegramCmd == app.UnknownTelegramCommand {
		if err := b.sessionService.ClearBotStateForUser(ctx, *telegramID); err != nil {
			return err
		}
		return b.unknownTelegramCommandHandler(ctx, update)
	}

	var telegramCallbackData *app.TelegramCallbackData
	if callbackQuery := update.CallbackQuery; callbackQuery != nil {
		telegramCallbackData, err = b.telegramBotService.ParseTelegramCallbackData(update.CallbackQuery)
		if err != nil {
			log.Debug("fail to parse telegram callback data", logger.F("telegramCallbackData", telegramCallbackData), logger.FError(err))
			return err
		}
	}
	userBotState := b.sessionService.GetBotStateForUser(ctx, *telegramID)
	log.Debug("got bot state from session service", logger.F("userBotState", userBotState))
	switch userBotState {
	case app.SelectLanguageBotState:
		return b.userSelectedLanguageBotStageHandler(ctx, update)
	case app.SelectCurrencyBotState:
		return b.userSelectedPreferredCurrencyBotStageHandler(ctx, update)
	case app.EnteringAmountCurrencyBotState:
		if callbackQuery := update.CallbackQuery; callbackQuery == nil {
			return b.enteringAmountCurrencyBotStageHandler(ctx, update)
		}
		break
	}

	if telegramCallbackData == nil {
		return b.helpTelegramCommandHandler(ctx, update)
	}
	switch telegramCallbackData.CallbackQueryCommand() {
	case app.BalanceCallbackQueryCommand:
		return b.balanceCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.MainMenuCallbackQueryCommand:
		return b.mainMenuCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.BuyNumberCallbackQueryCommand:
		return b.servicesCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.SelectSMSServiceCallbackQueryCommand:
		return b.selectServiceCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.HelpCallbackQueryCommand:
		return b.helpCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.LanguageCallbackQueryCommand:
		return b.languagesCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.SelectLanguageCallbackQueryCommand:
		return b.selectLanguageCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.HistoryCallbackQueryCommand:
		return b.historyCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.PayServiceCallbackQueryCommand:
		return b.payServiceQueryCommandHandler(ctx, update.CallbackQuery)
	case app.ListPayCurrenciesCallbackQueryCommand:
		return b.listPayCurrenciesCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.SelectPayCurrencyCallbackQueryCommand:
		return b.selectedPayCurrenciesCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.PreferredCurrenciesCallbackQueryCommand:
		return b.preferredCurrenciesQueryCommandHandler(ctx, update.CallbackQuery)
	case app.SelectPreferredCurrencyCallbackQueryCommand:
		return b.selectPreferredCurrencyQueryCommandHandler(ctx, update.CallbackQuery)
	case app.EmptyCallbackQueryCommand:
		return b.emptyQueryCommandHandler(ctx, update.CallbackQuery)
	default:
		return b.developingCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	}
}

func (b *botController) recordTelegramProfile(ctx context.Context, update *telegram.Update) error {
	log := b.container.GetLogger()
	telegramID, err := getTelegramID(update)
	if err != nil {
		return err
	}
	username, err := getTelegramUsername(update)
	if err != nil {
		log.Error("can't get telegram username", logger.FError(err))
		return err
	}
	profileExist, err := b.profileRepository.ExistsWithTelegramID(ctx, *telegramID)
	if err != nil {
		log.Error("existsWithTelegramID has failed", logger.FError(err))
		return err
	} else if !profileExist {
		log.Debug("record profile to db", logger.F("telegramID", telegramID))
		profile := &domain.Profile{
			TelegramID:     *telegramID,
			TelegramChatID: update.Message.Chat.ID,
			Username:       username,
		}
		_, err := b.profileRepository.Create(ctx, profile)
		if err != nil {
			log.Debug("fail to record profile to db", logger.F("telegramID", telegramID))
			return err
		}
	}
	return nil
}

func (b *botController) getLanguageCode(ctx context.Context, user telegram.User) (string, error) {
	profile, err := b.profileRepository.FetchByTelegramID(ctx, user.ID)
	defaultLang := "en"
	if err != nil {
		return defaultLang, err
	}
	if preferredLanguage := profile.PreferredLanguage; preferredLanguage != nil {
		return *preferredLanguage, nil
	}
	if user.LanguageCode != nil {
		return *user.LanguageCode, nil
	}
	return defaultLang, nil
}

func (b *botController) EditMessageMedia(callbackQuery *telegram.CallbackQuery, text string, photoURL string, replyMarkup any) error {
	log := b.container.GetLogger()
	photoMedia := telegram.InputPhotoMedia{
		Type:      "photo",
		Media:     photoURL,
		Caption:   utils.NewString(text),
		ParseMode: utils.NewString("MarkdownV2"),
	}
	editMessageMedia := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: replyMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessageMedia, app.EditMessageMediaTelegramMethod); err != nil {
		log.Error("fail to edit message media", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) AnswerCallbackQuery(callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Debug("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) AnswerCallbackQueryWithEditMessageMedia(
	callbackQuery *telegram.CallbackQuery,
	text string,
	photoURL string,
	replyMarkup any,
) error {
	log := b.container.GetLogger()
	if err := b.AnswerCallbackQuery(callbackQuery); err != nil {
		log.Debug("fail to answer callback query", logger.FError(err))
		return err
	}
	if err := b.EditMessageMedia(callbackQuery, text, photoURL, replyMarkup); err != nil {
		log.Debug("fail to perform EditMessageMedia", logger.FError(err))
		return err
	}
	return nil
}

func getTelegramID(update *telegram.Update) (*int64, error) {
	var from *telegram.User
	if update.Message != nil {
		from = update.Message.From
	} else if update.CallbackQuery != nil {
		from = &update.CallbackQuery.From
	}
	if from == nil {
		return nil, app.NilError
	}
	return &from.ID, nil
}

func getTelegramUsername(update *telegram.Update) (*string, error) {
	var from *telegram.User
	if update.Message != nil {
		from = update.Message.From
	} else if update.CallbackQuery != nil {
		from = &update.CallbackQuery.From
	}
	if from == nil {
		return nil, app.NilError
	}
	username := from.Username
	if username == nil {
		return nil, app.NilError
	}
	return username, nil
}

func isEmpty(update *telegram.Update) bool {
	return update.Message == nil && update.CallbackQuery == nil
}
