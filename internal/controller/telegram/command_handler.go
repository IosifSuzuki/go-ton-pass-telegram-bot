package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) startTelegramCommandHandler(ctx context.Context, ctxOptions *ContextOptions) error {
	if ctxOptions.Profile.PreferredLanguage == nil {
		return b.sendMessageToSelectInitialLanguage(ctx, ctxOptions)
	} else if ctxOptions.Profile.PreferredCurrency == nil {
		return b.sendMessageToSelectInitialPreferredCurrency(ctx, ctxOptions)
	}
	if err := b.sendMessageWelcome(ctx, ctxOptions); err != nil {
		return err
	}
	return b.sendMessageMainMenu(ctx, ctxOptions)
}

func (b *botController) helpTelegramCommandHandler(ctx context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	telegramID := ctxOptions.Update.GetTelegramID()
	if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
		log.Error("fail to clear bot state for telegram profile", logger.FError(err))
		return err
	}
	return b.sendHelpText(ctx, ctxOptions)
}

func (b *botController) unknownTelegramCommandHandler(ctx context.Context, ctxOptions *ContextOptions) error {
	telegramID := ctxOptions.Profile.TelegramID
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
		return err
	}
	text := b.container.GetLocalizer(preferredLanguage).LocalizedString("unknown_cmd_text")
	return b.sendMessagePlainText(ctx, text, ctxOptions)
}
