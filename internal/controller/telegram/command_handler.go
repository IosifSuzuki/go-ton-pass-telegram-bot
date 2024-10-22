package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/telegram"
)

func (b *botController) startTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	telegramID := update.Message.From.ID

	exist, _ := b.profileRepository.ExistsWithTelegramID(ctx, telegramID)
	if !exist {
		return b.messageToSelectLanguage(ctx, update)
	}
	profile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}
	if profile.PreferredLanguage == nil {
		return b.messageToSelectLanguage(ctx, update)
	} else if profile.PreferredCurrency == nil {
		return b.messageToSelectPreferredCurrency(ctx, update)
	}
	if err := b.messageWelcome(ctx, update); err != nil {
		return err
	}
	return b.messageMainMenu(ctx, update)
}

func (b *botController) helpTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	telegramID := update.Message.From.ID
	langTag, err := b.getLanguageCode(ctx, *update.Message.From)
	if err != nil {
		return err
	}

	if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
		return err
	}
	return b.messageWithPlainText(ctx, b.container.GetLocalizer(langTag).LocalizedString("help_cmd_text"), update)
}

func (b *botController) unknownTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	telegramID := update.Message.From.ID
	langTag, err := b.getLanguageCode(ctx, *update.Message.From)
	if err != nil {
		return err
	}

	if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
		return err
	}
	return b.messageWithPlainText(ctx, b.container.GetLocalizer(langTag).LocalizedString("unknown_cmd_text"), update)
}
