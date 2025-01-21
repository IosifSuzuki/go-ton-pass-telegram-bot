package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/manager"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/service/postpone"
	"go-ton-pass-telegram-bot/internal/worker"
	"go-ton-pass-telegram-bot/pkg/logger"
)

type BotController interface {
	Serve(*ContextOptions) error
}

type ContextOptions struct {
	TelegramInlineKeyboardManager manager.TelegramInlineKeyboardManager
	Update                        *telegram.Update
	Profile                       *domain.Profile
	IsMemberSubscription          bool
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
	container                  container.Container
	telegramBotService         service.TelegramBotService
	cryptoPayBot               service.CryptoPayBot
	sessionService             service.SessionService
	cacheService               service.Cache
	smsService                 service.SMSService
	postponeService            postpone.Postpone
	profileRepository          repository.ProfileRepository
	smsHistoryRepository       repository.SMSHistoryRepository
	temporalWorkflowRepository repository.TemporalWorkflowRepository
	exchangeRateWorker         worker.ExchangeRate
	smsActivateWorker          worker.SMSActivate
	formatterWorker            worker.Formatter
	callbackDataStack          service.CallbackDataStack
}

func NewBotController(
	container container.Container,
	sessionService service.SessionService,
	cacheService service.Cache,
	smsService service.SMSService,
	postponeService postpone.Postpone,
	profileRepository repository.ProfileRepository,
	smsHistoryRepository repository.SMSHistoryRepository,
	cryptoPayBot service.CryptoPayBot,
	exchangeRateWorker worker.ExchangeRate,
	temporalWorkflowRepository repository.TemporalWorkflowRepository,
) BotController {
	smsActivateWorker := worker.NewSMSActivate(container, smsService, cacheService)
	formatterWorker := worker.NewFormatter(container)
	callbackDataStack := service.NewCallbackDataStack(container, cacheService)
	return &botController{
		container:                  container,
		telegramBotService:         service.NewTelegramBot(container),
		cryptoPayBot:               cryptoPayBot,
		sessionService:             sessionService,
		cacheService:               cacheService,
		smsService:                 smsService,
		postponeService:            postponeService,
		profileRepository:          profileRepository,
		smsHistoryRepository:       smsHistoryRepository,
		temporalWorkflowRepository: temporalWorkflowRepository,
		exchangeRateWorker:         exchangeRateWorker,
		smsActivateWorker:          smsActivateWorker,
		formatterWorker:            formatterWorker,
		callbackDataStack:          callbackDataStack,
	}
}

func (b *botController) Serve(ctxOptions *ContextOptions) error {
	ctx := context.Background()
	log := b.container.GetLogger()
	log.Debug(
		"serve data from bot",
		logger.F("update", ctxOptions.Update),
	)

	hasSubscription, err := b.ServeSubscription(ctx, ctxOptions)
	if err != nil {
		log.Error("fail to serve subscription", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if !hasSubscription {
		return nil
	}

	telegramCmd, err := b.telegramBotService.ParseTelegramCommand(ctxOptions.Update)
	switch telegramCmd {
	case app.StartTelegramCommand:
		return b.startTelegramCommandHandler(ctx, ctxOptions)
	case app.HelpTelegramCommand:
		return b.helpTelegramCommandHandler(ctx, ctxOptions)
	}
	if err != nil && telegramCmd == app.UnknownTelegramCommand {
		err := b.sessionService.ClearBotStateForUser(ctx, ctxOptions.Update.GetChatID())
		if err != nil {
			return err
		}
		return b.unknownTelegramCommandHandler(ctx, ctxOptions)
	}

	userBotState := b.sessionService.GetBotStateForUser(ctx, ctxOptions.Update.GetTelegramID())
	log.Debug(
		"got bot state from session service",
		logger.F("userBotState", userBotState),
	)
	switch userBotState {
	case app.EnteringAmountCurrencyBotState:
		return b.enteringAmountCurrencyBotStageHandler(ctx, ctxOptions)
	}
	callbackQuery := ctxOptions.Update.CallbackQuery
	if callbackQuery == nil {
		log.Error(
			"fail to retrieve callback query from update",
			logger.FError(err),
		)
		// fallback
		return b.helpTelegramCommandHandler(ctx, ctxOptions)
	}
	telegramCallbackData, err := b.telegramBotService.ParseTelegramCallbackData(callbackQuery)
	if err != nil {
		log.Error(
			"fail to parse telegram callback data",
			logger.F("telegramCallbackData", telegramCallbackData),
			logger.FError(err),
		)
		return err
	}
	if telegramCallbackData == nil {
		log.Error("telegramCallbackData has nil value")
		return b.helpTelegramCommandHandler(ctx, ctxOptions)
	}
	transformedTelegramCallbackData, err := b.ServeCallbackQueryStack(ctx, callbackQuery, telegramCallbackData)
	if err != nil {
		log.Error("fail to serve callback query stack", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	callbackQueryCommand := transformedTelegramCallbackData.CallbackQueryCommand()
	switch callbackQueryCommand {
	case app.SelectInitialLanguageCallbackQueryCommand:
		return b.selectedInitialLanguageCallbackQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.SelectInitialPreferredCurrencyCallbackQueryCommand:
		return b.selectedInitialPreferredCurrencyCallbackQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.BalanceCallbackQueryCommand:
		return b.balanceCallbackQueryCommandHandler(ctx, ctxOptions)
	case app.MainMenuCallbackQueryCommand:
		return b.mainMenuCallbackQueryCommandHandler(ctx, ctxOptions)
	case app.BuyNumberCallbackQueryCommand:
		return b.servicesCallbackQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.SelectSMSServiceCallbackQueryCommand:
		return b.selectServiceCallbackQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.HelpCallbackQueryCommand:
		return b.helpCallbackQueryCommandHandler(ctx, ctxOptions)
	case app.LanguageCallbackQueryCommand:
		return b.languagesCallbackQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.SelectLanguageCallbackQueryCommand:
		return b.selectedLanguageCallbackQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.HistoryCallbackQueryCommand:
		return b.historyCallbackQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.PayServiceCallbackQueryCommand:
		return b.payServiceQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.ListPayCurrenciesCallbackQueryCommand:
		return b.listPayCurrenciesCallbackQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.SelectPayCurrencyCallbackQueryCommand:
		return b.selectedPayCurrenciesCallbackQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.CancelEnterAmountCallbackQueryCommand:
		return b.cancelEnteringAmountCallbackQueryCommandHandler(ctx, ctxOptions)
	case app.PreferredCurrenciesCallbackQueryCommand:
		return b.preferredCurrenciesQueryCommandHandler(ctx, ctxOptions)
	case app.SelectPreferredCurrencyCallbackQueryCommand:
		return b.selectPreferredCurrencyQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.EmptyCallbackQueryCommand:
		return b.emptyQueryCommandHandler(ctx, callbackQuery)
	case app.DeleteCryptoBotInvoiceCallbackQueryCommand:
		return b.deleteCryptoBotQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.ConfirmationPayServiceCallbackQueryCommand:
		return b.confirmServiceQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	case app.RefundAmountFromSMSActivationCallbackQueryCommand:
		return b.refundAmountFromSMSActivationQueryCommandHandler(ctx, ctxOptions, transformedTelegramCallbackData)
	default:
		return b.developingCallbackQueryCommandHandler(ctx, ctxOptions)
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

func (b *botController) ServeCallbackQueryStack(
	ctx context.Context,
	callbackQuery *telegram.CallbackQuery,
	telegramCallbackData *app.TelegramCallbackData,
) (*app.TelegramCallbackData, error) {
	log := b.container.GetLogger()
	requestedCallbackQueryCommand := telegramCallbackData.CallbackQueryCommand()
	telegramMessagingInfo := service.TelegramMessagingInfo{
		ChatID:    callbackQuery.Message.Chat.ID,
		MessageID: callbackQuery.Message.ID,
	}
	previousCallbackQueryCommandPointer, _ := b.cacheService.GetLastCallbackQueryCommand(ctx, telegramMessagingInfo)
	previousCallbackQueryCommand := app.NotCallbackQueryCommand
	if previousCallbackQueryCommandPointer != nil {
		previousCallbackQueryCommand = *previousCallbackQueryCommandPointer
	}
	var newTelegramCallbackData = telegramCallbackData
	if requestedCallbackQueryCommand == app.BackCallbackQueryCommand {
		var (
			poppedCallbackData *app.TelegramCallbackData
			err                error
		)
		if requestedCallbackQueryCommand == app.BackCallbackQueryCommand {
			_, _ = b.callbackDataStack.Pop(ctx, callbackQuery)
			poppedCallbackData, err = b.callbackDataStack.Top(ctx, callbackQuery)
		} else {
			poppedCallbackData, err = b.callbackDataStack.Pop(ctx, callbackQuery)
		}
		if err != nil {
			log.Error("fail to perform pop operation in callbackDataStack", logger.FError(err))
		}
		if poppedCallbackData != nil {
			newTelegramCallbackData = poppedCallbackData
		} else {
			newTelegramCallbackData = &app.TelegramCallbackData{
				Name:       app.MainMenuCallbackQueryCmdText,
				Parameters: nil,
			}
		}
	}
	switch requestedCallbackQueryCommand {
	case
		app.EmptyCallbackQueryCommand,
		app.NotCallbackQueryCommand,
		app.SelectInitialLanguageCallbackQueryCommand,
		app.SelectInitialPreferredCurrencyCallbackQueryCommand,
		app.BackCallbackQueryCommand,
		app.CancelEnterAmountCallbackQueryCommand,
		app.SelectPayCurrencyCallbackQueryCommand:
		// skip serving these commands
		break
	default:
		if previousCallbackQueryCommand == requestedCallbackQueryCommand {
			// we should omit pushing commands with pagination
			break
		}
		err := b.callbackDataStack.Push(ctx, callbackQuery)
		if err != nil {
			log.Error("fail to record callback query to cache", logger.FError(err))
			return nil, err
		}
	}
	if err := b.cacheService.SetLastCallbackQueryCommand(ctx, requestedCallbackQueryCommand, telegramMessagingInfo); err != nil {
		log.Error("fail to set last callbackQueryCommand", logger.FError(err))
		return nil, err
	}
	commandNames, err := b.callbackDataStack.DebugListCommands(ctx, callbackQuery)
	if err != nil {
		log.Debug("fail to get list commands", logger.FError(err))
	}
	log.Debug("callback query commands in cache", logger.F("commands", commandNames))
	return newTelegramCallbackData, nil
}

func (b *botController) ServeSubscription(ctx context.Context, ctxOption *ContextOptions) (bool, error) {
	shouldSendMessageAboutSubscription := !ctxOption.IsMemberSubscription
	if ctxOption.Profile.PreferredLanguage == nil || ctxOption.Profile.PreferredCurrency == nil {
		shouldSendMessageAboutSubscription = true
	}
	if shouldSendMessageAboutSubscription {
		return false, b.sendMessageSubscription(ctx, ctxOption)
	}
	return true, nil
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
