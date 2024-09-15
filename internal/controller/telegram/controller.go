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
	profileRepository  repository.ProfileRepository
}

func NewBotController(container container.Container, sessionService service.SessionService, profileRepository repository.ProfileRepository) BotController {
	return &botController{
		container:          container,
		telegramBotService: service.NewTelegramBot(container),
		sessionService:     sessionService,
		profileRepository:  profileRepository,
	}
}

func (b *botController) Serve(update *telegram.Update) error {
	if isEmpty(update) {
		return app.EmptyUpdateError
	}

	log := b.container.GetLogger()
	ctx := context.Background()
	telegramID := getTelegramID(update)
	log.Debug("receive telegram message", logger.F("update", update))

	if err := b.recordTelegramProfile(ctx, update); err != nil {
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
		if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
			return err
		}
		return b.unknownTelegramCommandHandler(ctx, update)
	}

	userBotState := b.sessionService.GetBotStateForUser(ctx, telegramID)
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
	case app.HelpCallbackQueryCommand, app.HistoryCallbackQueryCommand, app.BuyNumberCallbackQueryCommand, app.LanguageCallbackQueryCommand:
		return b.unsupportedCallbackQueryCommandHandle(ctx, update.CallbackQuery)
	}

	return b.helpTelegramCommandHandler(ctx, update)
}

func (b *botController) recordTelegramProfile(ctx context.Context, update *telegram.Update) error {
	log := b.container.GetLogger()
	telegramID := getTelegramID(update)
	username := getTelegramUsername(update)
	profileExist, err := b.profileRepository.ExistsWithTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("existsWithTelegramID has failed", logger.FError(err))
		return err
	} else if !profileExist {
		log.Debug("record profile to db", logger.F("telegramID", telegramID))
		profile := &domain.Profile{
			TelegramID: telegramID,
			Username:   &username,
		}
		_, err := b.profileRepository.Create(ctx, profile)
		if err != nil {
			log.Debug("fail to record profile to db", logger.F("telegramID", telegramID))
			return err
		}
	}
	return nil
}

func (b *botController) getLanguageCode(ctx context.Context, user telegram.User) string {
	profile, err := b.profileRepository.FetchByTelegramID(ctx, user.ID)
	if err != nil {
		return user.LanguageCode
	}
	if preferredLanguage := profile.PreferredLanguage; preferredLanguage != nil {
		return *preferredLanguage
	}
	return user.LanguageCode
}

func getTelegramID(update *telegram.Update) int64 {
	telegramID := update.Message.From.ID
	if telegramID == 0 && update.CallbackQuery != nil {
		telegramID = update.CallbackQuery.From.ID
	}
	return telegramID
}

func getTelegramUsername(update *telegram.Update) string {
	username := update.Message.From.Username
	if len(username) == 0 && update.CallbackQuery != nil {
		username = update.CallbackQuery.From.Username
	}
	return username
}

func isEmpty(update *telegram.Update) bool {
	return update.Message.ID == 0 && update.CallbackQuery == nil
}
