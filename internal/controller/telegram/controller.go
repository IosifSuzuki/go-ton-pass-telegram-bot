package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/pkg/logger"
)

type BotController interface {
	Serve(update *telegram.Update) error
}

type botController struct {
	container          container.Container
	telegramBotService service.TelegramBotService
	sessionService     service.SessionService
	smsService         service.SMSService
	profileRepository  repository.ProfileRepository
}

func NewBotController(
	container container.Container,
	sessionService service.SessionService,
	smsService service.SMSService,
	profileRepository repository.ProfileRepository,
) BotController {
	return &botController{
		container:          container,
		telegramBotService: service.NewTelegramBot(container),
		sessionService:     sessionService,
		smsService:         smsService,
		profileRepository:  profileRepository,
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

	userBotState := b.sessionService.GetBotStateForUser(ctx, *telegramID)
	switch userBotState {
	case app.SelectLanguageBotState:
		return b.userSelectedLanguageBotStageHandler(ctx, update)
	case app.SelectCurrencyBotState:
		return b.userSelectedCurrencyBotStageHandler(ctx, update)
	}

	callbackQueryCommand := b.telegramBotService.ParseCallbackQueryCommand(update)
	switch callbackQueryCommand {
	case app.BalanceCallbackQueryCommand:
		return b.balanceCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.MainMenuCallbackQueryCommand:
		return b.mainMenuCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.BuyNumberCallbackQueryCommand:
		return b.servicesCallbackQueryCommandHandle(ctx, update.CallbackQuery)
	case app.HelpCallbackQueryCommand, app.HistoryCallbackQueryCommand, app.LanguageCallbackQueryCommand:
		return b.unsupportedCallbackQueryCommandHandle(ctx, update.CallbackQuery)
	}

	return b.helpTelegramCommandHandler(ctx, update)
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
			TelegramID: *telegramID,
			Username:   username,
		}
		_, err := b.profileRepository.Create(ctx, profile)
		if err != nil {
			log.Debug("fail to record profile to db", logger.F("telegramID", telegramID))
			return err
		}
	}
	return nil
}

func (b *botController) getLanguageCode(ctx context.Context, user telegram.User) (*string, error) {
	profile, err := b.profileRepository.FetchByTelegramID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if preferredLanguage := profile.PreferredLanguage; preferredLanguage != nil {
		return preferredLanguage, nil
	}
	return user.LanguageCode, nil
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
