package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) editMessageMainMenu(ctx context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	callbackQuery := ctxOptions.Update.CallbackQuery
	localizer := b.container.GetLocalizer(preferredLanguage)
	mainMenuInlineKeyboardMarkup, err := ctxOptions.TelegramInlineKeyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get a main menu keyboard markup", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		localizer.LocalizedString("short_description_markdown"),
		avatarImageURL,
		mainMenuInlineKeyboardMarkup,
	)
}

func (b *botController) editMessageDevelopingMode(_ context.Context, ctxOptions *ContextOptions) error {
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	text := b.container.GetLocalizer(preferredLanguage).LocalizedString("development_process_markdown")
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        ctxOptions.Update.CallbackQuery.ID,
		Text:      &text,
		ShowAlert: true,
	}
	return b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod)
}

func (b *botController) editMessageListPayCurrencies(ctx context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	payCurrenciesInlineKeyboardMarkup, err := ctxOptions.TelegramInlineKeyboardManager.PayCurrenciesKeyboardMarkup()
	if err != nil {
		log.Error("fail to get a pay currencies inline keyboard", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		ctxOptions.Update.CallbackQuery,
		localizer.LocalizedString("select_currency_to_pay_markdown"),
		topUpImageURL,
		payCurrenciesInlineKeyboardMarkup,
	)
}

func (b *botController) editMessageHelp(
	_ context.Context,
	ctxOptions *ContextOptions,
) error {
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	return b.AnswerCallbackQueryWithEditMessageMedia(
		ctxOptions.Update.CallbackQuery,
		localizer.LocalizedString("help_cmd_text_markdown"),
		helpImageURL,
		ctxOptions.TelegramInlineKeyboardManager.BackKeyboardMarkup(),
	)
}

func (b *botController) editMessageInternalServerError(_ context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	callbackQuery := ctxOptions.Update.CallbackQuery
	localizer := b.container.GetLocalizer(preferredLanguage)
	mainMenuInlineKeyboardMarkup, err := ctxOptions.TelegramInlineKeyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get a main menu keyboard markup", logger.FError(err))
		return err
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		localizer.LocalizedString("internal_error_markdown"),
		avatarImageURL,
		mainMenuInlineKeyboardMarkup,
	)
}

func (b *botController) editMessageProfileBalance(ctx context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	callbackQuery := ctxOptions.Update.CallbackQuery
	localizer := b.container.GetLocalizer(preferredLanguage)
	preferredCurrency := ctxOptions.Profile.PreferredCurrency
	topUpBalanceKeyboardMarkup, err := ctxOptions.TelegramInlineKeyboardManager.TopUpBalanceKeyboardMarkup()
	if err != nil {
		log.Error("fail to get a main menu keyboard markup", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if preferredCurrency == nil {
		log.Error("preferred currency has nil value")
		return app.NilError
	}
	currency := b.container.GetConfig().CurrencyByAbbr(*preferredCurrency)
	convertedBalance, err := b.exchangeRateWorker.ConvertFromUSD(ctxOptions.Profile.Balance, *preferredCurrency)
	if err != nil {
		log.Error(
			"fail to convert balance from usd balance",
			logger.FError(err),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	text := localizer.LocalizedStringWithTemplateData("your_balance_is_markdown", map[string]any{
		"Balance": utils.EscapeMarkdownText(utils.CurrencyAmountTextFormat(*convertedBalance, *currency)),
	})
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		text,
		topUpImageURL,
		topUpBalanceKeyboardMarkup,
	)
}

func (b *botController) editMessageProfileLanguages(ctx context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	replyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.LanguagesKeyboardMarkup()
	if err != nil {
		log.Error("fail to get languages inline keyboard", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	language := b.container.GetConfig().LanguageByCode(preferredLanguage)
	text := localizer.LocalizedStringWithTemplateData("choose_language_markdown", map[string]any{
		"Language": utils.EscapeMarkdownText(utils.LanguageTextFormat(*language)),
	})
	return b.AnswerCallbackQueryWithEditMessageMedia(
		ctxOptions.Update.CallbackQuery,
		text,
		selectPreferredLanguageImageURL,
		replyMarkup,
	)
}

func (b *botController) editMessageHistories(
	ctx context.Context,
	ctxOptions *ContextOptions,
	pagination app.Pagination,
) error {
	log := b.container.GetLogger()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	profileID := ctxOptions.Profile.ID
	offset := pagination.CurrentPage * pagination.ItemsPerPage
	if pagination.LenItems == 0 {
		text := localizer.LocalizedString("empty_history_markdown")
		return b.EditMessageMedia(
			ctxOptions.Update.CallbackQuery,
			text,
			historyImageURL,
			ctxOptions.TelegramInlineKeyboardManager.BackKeyboardMarkup(),
		)
	}
	smsHistories, err := b.smsHistoryRepository.FetchList(ctx, profileID, offset, pagination.ItemsPerPage)
	if err != nil {
		log.Error("fail to get list of sms history",
			logger.F("profile_id", profileID),
			logger.F("offset", offset),
			logger.F("items_per_page", pagination.ItemsPerPage),
			logger.FError(err),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	prevPageParameters := []any{pagination.PrevPage(), pagination.ItemsPerPage}
	nextPageParameters := []any{pagination.NextPage(), pagination.ItemsPerPage}
	pageControlButtons, err := ctxOptions.TelegramInlineKeyboardManager.PageControlKeyboardButtons(
		app.HistoryCallbackQueryCmdText,
		pagination,
		prevPageParameters,
		nextPageParameters,
	)
	if err != nil {
		log.Error("fail to get page control keyboard buttons", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	backButton := ctxOptions.TelegramInlineKeyboardManager.BackKeyboardButton()
	buttons := [][]telegram.InlineKeyboardButton{
		pageControlButtons,
		{*backButton},
	}
	replyMarkup := &telegram.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}
	text := b.formatterWorker.SHSHistories(preferredLanguage, smsHistories)
	return b.AnswerCallbackQueryWithEditMessageMedia(
		ctxOptions.Update.CallbackQuery,
		text,
		historyImageURL,
		replyMarkup,
	)
}

func (b *botController) editMessageServices(
	ctx context.Context,
	ctxOptions *ContextOptions,
	pagination app.Pagination,
	smsServices []sms.Service,
) error {
	log := b.container.GetLogger()
	replyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.ServicesInlineKeyboardMarkup(smsServices, pagination)
	if err != nil {
		log.Error("fail to get services inline keyboard markup", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	text := localizer.LocalizedString("select_sms_service_markdown")
	return b.AnswerCallbackQueryWithEditMessageMedia(
		ctxOptions.Update.CallbackQuery,
		text,
		chooseServiceImageURL,
		replyMarkup,
	)
}

func (b *botController) editMessageServiceCountries(
	ctx context.Context,
	ctxOptions *ContextOptions,
	pagination app.Pagination,
	selectedServiceCode string,
	servicePrices []sms.PriceForService,
	countries []sms.Country,
) error {
	log := b.container.GetLogger()
	if ctxOptions.Profile.PreferredCurrency == nil {
		log.Error("profile must have preferred currency")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	preferredCurrency := *ctxOptions.Profile.PreferredCurrency
	replyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.ServiceCountriesInlineKeyboardMarkup(
		selectedServiceCode,
		preferredCurrency,
		pagination,
		servicePrices,
		countries,
	)
	if err != nil {
		log.Error("fail to get service countries inline keyboard markup", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	text := localizer.LocalizedString("select_sms_service_with_country_markdown")
	return b.AnswerCallbackQueryWithEditMessageMedia(
		ctxOptions.Update.CallbackQuery,
		text,
		chooseCountryImageURL,
		replyMarkup,
	)
}

func (b *botController) editMessagePreferredCurrencies(
	ctx context.Context,
	ctxOptions *ContextOptions,
) error {
	log := b.container.GetLogger()
	if ctxOptions.Profile.PreferredCurrency == nil {
		log.Error("profile must have preferred currency")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	preferredCurrency := *ctxOptions.Profile.PreferredCurrency
	currency := b.container.GetConfig().CurrencyByAbbr(preferredCurrency)
	replyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.PreferredCurrenciesKeyboardMarkup()
	if err != nil {
		log.Error("fail to get currencies keyboard markup", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localizer := b.container.GetLocalizer(preferredLanguage)
	myCurrencyText := utils.ShortCurrencyTextFormat(*currency)
	text := localizer.LocalizedStringWithTemplateData("choose_preferred_currency_markdown", map[string]any{
		"Currency": utils.EscapeMarkdownText(myCurrencyText),
	})
	return b.AnswerCallbackQueryWithEditMessageMedia(
		ctxOptions.Update.CallbackQuery,
		text,
		selectPreferredCurrencyImageURL,
		replyMarkup,
	)
}

func (b *botController) editMessageConfirmService(
	ctx context.Context,
	ctxOptions *ContextOptions,
	service *sms.Service,
	country *sms.Country,
	priceInRub float64,
	priceWithFeeInPreferredCurrency float64,
) error {
	log := b.container.GetLogger()
	profile := ctxOptions.Profile
	if profile.PreferredLanguage == nil {
		log.Error("profile must have preferred language")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if profile.PreferredCurrency == nil {
		log.Error("profile must have preferred currency")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	preferredLanguage := *profile.PreferredLanguage
	preferredCurrencyAbbr := *profile.PreferredCurrency
	preferredCurrency := b.container.GetConfig().CurrencyByAbbr(preferredCurrencyAbbr)
	if preferredCurrency == nil {
		log.Error(
			"can't find currency",
			logger.F("preferred_currency_abbr", *profile.PreferredCurrency),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	text := b.formatterWorker.ConfirmationPay(
		preferredLanguage,
		service,
		country,
		priceWithFeeInPreferredCurrency,
		*preferredCurrency,
	)
	replyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.ConfirmationPayInlineKeyboardMarkup(
		service.Code,
		country.ID,
		priceInRub,
	)
	if err != nil {
		log.Error("fail to get confirmation inline keyboard", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		ctxOptions.Update.CallbackQuery,
		text,
		avatarImageURL,
		replyMarkup,
	)
}

func (b *botController) editMessageEnterAmountPayError(
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
	text := localizer.LocalizedStringWithTemplateData("enter_amount_for_payment_in_currency_error_markdown",
		map[string]any{
			"Currency": utils.EscapeMarkdownText(utils.ShortCurrencyTextFormat(*profileCurrency)),
		},
	)
	enteringAmountInlineKeyboardMarkup, err := ctxOptions.TelegramInlineKeyboardManager.EnteringAmountInlineKeyboardMarkup()
	if err != nil {
		log.Error("fail to get entering amount inline keyboard markup", logger.FError(err))
		return err
	}
	return b.SendTextWithPhotoMedia(
		ctxOptions.Update.GetChatID(),
		text,
		avatarImageURL,
		enteringAmountInlineKeyboardMarkup,
	)
}
