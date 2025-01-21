package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) sendMessageToSelectInitialLanguage(_ context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	replyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.InitialLanguagesKeyboardMarkup()
	if err != nil {
		log.Error("fail to get initial languages keyboard markup", logger.FError(err))
		return err
	}
	sendPhoto := telegram.SendPhoto{
		ChatID:      ctxOptions.Update.GetChatID(),
		Photo:       selectPreferredLanguageImageURL,
		Caption:     localizer.LocalizedString("select_preferred_language_markdown"),
		ReplyMarkup: replyMarkup,
	}
	return b.telegramBotService.SendResponse(sendPhoto, app.SendPhotoTelegramMethod)
}

func (b *botController) sendMessageToSelectInitialPreferredCurrency(_ context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	replyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.InitialPreferredCurrenciesKeyboardMarkup()
	if err != nil {
		log.Error("fail to get initial preferred currencies keyboard markup", logger.FError(err))
		return err
	}
	resp := telegram.SendPhoto{
		ChatID:      ctxOptions.Update.GetChatID(),
		Photo:       selectPreferredCurrencyImageURL,
		Caption:     localizer.LocalizedString("select_preferred_currency_markdown"),
		ReplyMarkup: replyMarkup,
	}
	return b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod)
}

func (b *botController) sendMessageWelcome(_ context.Context, ctxOptions *ContextOptions) error {
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	sendPhotoResp := telegram.SendPhoto{
		ChatID:    ctxOptions.Update.GetChatID(),
		Photo:     welcomeImageURL,
		Caption:   b.container.GetLocalizer(preferredLanguage).LocalizedString("bot_markdown_description"),
		ParseMode: utils.NewString("MarkdownV2"),
	}
	return b.telegramBotService.SendResponse(sendPhotoResp, app.SendPhotoTelegramMethod)
}

func (b *botController) sendMessageMainMenu(_ context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	mainMenuInlineKeyboardMarkup, err := ctxOptions.TelegramInlineKeyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error(
			"fail to get a main menu keyboard markup",
			logger.FError(err),
		)
		return err
	}
	resp := telegram.SendPhoto{
		ChatID:      ctxOptions.Update.GetChatID(),
		Caption:     b.container.GetLocalizer(preferredLanguage).LocalizedString("short_description_markdown"),
		Photo:       avatarImageURL,
		ParseMode:   utils.NewString("MarkdownV2"),
		ReplyMarkup: mainMenuInlineKeyboardMarkup,
	}
	return b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod)
}

func (b *botController) sendMessageEnterAmountCurrency(
	ctx context.Context,
	ctxOptions *ContextOptions,
) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	profileCurrency := b.container.GetConfig().CurrencyByAbbr(*ctxOptions.Profile.PreferredCurrency)
	if profileCurrency == nil {
		log.Error("fail to get profile's currency")
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	enteringAmountInlineKeyboardMarkup, err := ctxOptions.TelegramInlineKeyboardManager.EnteringAmountInlineKeyboardMarkup()
	if err != nil {
		log.Error("fail to entering amount inline keyboard markup", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	resp := telegram.SendPhoto{
		ChatID: ctxOptions.Update.GetChatID(),
		Caption: localizer.LocalizedStringWithTemplateData("enter_amount_for_payment_in_currency_markdown", map[string]any{
			"Currency": utils.EscapeMarkdownText(utils.ShortCurrencyTextFormat(*profileCurrency)),
		}),
		Photo:       enterAmountImageURL,
		ParseMode:   utils.NewString("MarkdownV2"),
		ReplyMarkup: enteringAmountInlineKeyboardMarkup,
	}
	return b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod)
}

func (b *botController) sendMessagePlainText(_ context.Context, text string, options *ContextOptions) error {
	resp := telegram.SendResponse{
		ChatID: options.Update.GetChatID(),
		Text:   text,
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) sendHelpText(_ context.Context, ctxOptions *ContextOptions) error {
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	text := b.container.GetLocalizer(preferredLanguage).LocalizedString("help_cmd_text_markdown")
	resp := telegram.SendResponse{
		ChatID:      ctxOptions.Update.GetChatID(),
		Text:        text,
		ParseMode:   utils.NewString("MarkdownV2"),
		ReplyMarkup: nil,
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) sendMessageInternalServerError(_ context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	replyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error(
			"fail to get main menu keyboard markup",
			logger.FError(err),
		)
	}
	return b.SendTextWithPhotoMedia(
		ctxOptions.Update.GetChatID(),
		localizer.LocalizedString("internal_error_markdown"),
		avatarImageURL,
		replyMarkup,
	)
}

func (b *botController) sendMessageStartSMSActivation(
	ctx context.Context,
	ctxOptions *ContextOptions,
	smsHistory *domain.SMSHistory,
	smsHistoryID int64,
) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	text := b.formatterWorker.StartSMSActivation(preferredLanguage, smsHistory)
	refundReplyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.RefundInlineKeyboardMarkup(smsHistoryID)
	if err != nil {
		log.Error("fail to get refund inline keyboard", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	return b.SendTextWithPhotoMedia(
		ctxOptions.Update.GetChatID(),
		text,
		avatarImageURL,
		refundReplyMarkup,
	)
}

func (b *botController) sendMessageSuccessfullyDeletedInvoice(
	_ context.Context,
	ctxOptions *ContextOptions,
) error {
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	text := localizer.LocalizedString("success_deleted_invoice_markdown")
	return b.SendTextWithPhotoMedia(
		ctxOptions.Update.GetChatID(),
		text,
		avatarImageURL,
		nil,
	)
}

func (b *botController) sendMessageSubscription(_ context.Context, ctxOptions *ContextOptions) error {
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	text := localizer.LocalizedStringWithTemplateData("subscribe_to_channel_markdown", map[string]any{
		"Channel": "@tonpassnews",
	})
	return b.SendTextWithPhotoMedia(
		ctxOptions.Update.GetChatID(),
		text,
		avatarImageURL,
		nil,
	)
}

func (b *botController) sendMessageConfirmTouchUpBalance(
	ctx context.Context,
	ctxOptions *ContextOptions,
	invoice *bot.Invoice,
) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	text := localizer.LocalizedString("crypto_bot_pay_title_markdown")
	cryptoPayReplyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.CryptoPayBotKeyboardMarkup(
		invoice.BotInvoiceURL,
		invoice.ID,
	)
	if err != nil {
		log.Error("fail to get crypto pay bot keyboard markup", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	return b.SendTextWithPhotoMedia(
		ctxOptions.Update.GetChatID(),
		text,
		avatarImageURL,
		cryptoPayReplyMarkup,
	)
}
