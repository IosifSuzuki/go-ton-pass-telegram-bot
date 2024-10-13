package telegram

import (
	"context"
	"fmt"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
)

const (
	MaxInlineKeyboardRows    = 8
	MaxInlineKeyboardColumns = 3
)

func (b *botController) getLanguagesInlineKeyboardMarkup(ctx context.Context, user telegram.User) (*telegram.InlineKeyboardMarkup, error) {
	langTag, err := b.getLanguageCode(ctx, user)
	if err != nil {
		return nil, err
	}
	languages := b.container.GetConfig().AvailableLanguages()
	keyboardButtons := make([]telegram.InlineKeyboardButton, 0, len(languages))
	for _, language := range languages {
		parameters := []any{language.Code}
		selectLanguageTelegramCallbackData := app.TelegramCallbackData{
			Name:       app.SelectLanguageCallbackQueryCmdText,
			Parameters: &parameters,
		}
		data, err := utils.EncodeTelegramCallbackData(selectLanguageTelegramCallbackData)
		if err != nil {
			continue
		}
		presentableLanguageText := utils.LanguageTextFormat(language)
		languageKeyboardButton := telegram.InlineKeyboardButton{
			Text: presentableLanguageText,
			Data: data,
		}
		keyboardButtons = append(keyboardButtons, languageKeyboardButton)
	}
	backToMainMenuKeyboardButton, err := b.getMenuInlineKeyboardButton(*langTag)
	if err != nil {
		return nil, err
	}
	keyboardButtons = append(keyboardButtons, *backToMainMenuKeyboardButton)
	gridKeyboardButtons := b.prepareGridInlineKeyboardButton(keyboardButtons, 2)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridKeyboardButtons,
	}, nil
}

func (b *botController) getServicesInlineKeyboardMarkup(ctx context.Context, callbackQuery *telegram.CallbackQuery, pagination *app.Pagination[sms.Service]) (*telegram.InlineKeyboardMarkup, error) {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to getLanguageCode", logger.FError(err))
		return nil, err
	}
	inlineKeyboardButtons := make([]telegram.InlineKeyboardButton, 0, pagination.ItemsPerPage)
	startIndex := pagination.CurrentPage * pagination.ItemsPerPage
	if startIndex > pagination.Len() {
		return nil, app.IndexOutOfRangeError
	}
	endIndex := startIndex + pagination.ItemsPerPage
	if endIndex > pagination.Len() {
		endIndex = pagination.Len()
	}
	dataSourceSlice := pagination.DataSource[startIndex:endIndex]
	for _, smsService := range dataSourceSlice {
		parameters := []any{smsService.Code, 0}
		selectSMSServiceTelegramCallbackData := app.TelegramCallbackData{
			Name:       app.SelectSMSServiceCallbackQueryCmdText,
			Parameters: &parameters,
		}
		data, err := utils.EncodeTelegramCallbackData(selectSMSServiceTelegramCallbackData)
		if err != nil {
			log.Error("fail to encode telegram callback data", logger.FError(err))
			continue
		}
		inlineKeyboardButtons = append(inlineKeyboardButtons, telegram.InlineKeyboardButton{
			Text: utils.ButtonTitle(smsService.Name, "🌐"),
			Data: data,
		})
	}
	mainMenuInlineKeyboardButton, err := b.getMenuInlineKeyboardButton(*langTag)
	if err != nil {
		return nil, err
	}
	pageControlsButtons, err := getPageControlInlineKeyboardButtons(pagination, app.BuyNumberCallbackQueryCmdText, []any{})
	if err != nil {
		return nil, err
	}
	gridInlineKeyboardButtons := b.prepareGridInlineKeyboardButton(inlineKeyboardButtons, MaxInlineKeyboardColumns)
	gridInlineKeyboardButtons = append(gridInlineKeyboardButtons, pageControlsButtons)
	gridInlineKeyboardButtons = append(gridInlineKeyboardButtons, []telegram.InlineKeyboardButton{
		*mainMenuInlineKeyboardButton,
	})
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridInlineKeyboardButtons,
	}, nil
}

func (b *botController) getMainMenuInlineKeyboardMarkup(ctx context.Context, user telegram.User) (*telegram.InlineKeyboardMarkup, error) {
	langTag, err := b.getLanguageCode(ctx, user)
	if err != nil {
		return nil, err
	}
	buyNumberParameters := []any{0}
	localizer := b.container.GetLocalizer(*langTag)
	balanceTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.BalanceCallbackQueryCmdText,
		Parameters: nil,
	}
	buyNumberTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.BuyNumberCallbackQueryCmdText,
		Parameters: &buyNumberParameters,
	}
	helpTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.HelpCallbackQueryCmdText,
		Parameters: nil,
	}
	historyTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.HistoryCallbackQueryCmdText,
		Parameters: nil,
	}
	languageTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.LanguageCallbackQueryCmdText,
		Parameters: nil,
	}
	preferredCurrenciesCallbackData := app.TelegramCallbackData{
		Name:       app.PreferredCurrenciesCallbackQueryCmdText,
		Parameters: nil,
	}
	balanceData, err := utils.EncodeTelegramCallbackData(balanceTelegramCallbackData)
	if err != nil {
		return nil, err
	}
	buyNumberData, err := utils.EncodeTelegramCallbackData(buyNumberTelegramCallbackData)
	if err != nil {
		return nil, err
	}
	helpData, err := utils.EncodeTelegramCallbackData(helpTelegramCallbackData)
	if err != nil {
		return nil, err
	}
	historyData, err := utils.EncodeTelegramCallbackData(historyTelegramCallbackData)
	if err != nil {
		return nil, err
	}
	languageData, err := utils.EncodeTelegramCallbackData(languageTelegramCallbackData)
	if err != nil {
		return nil, err
	}
	preferredCurrenciesData, err := utils.EncodeTelegramCallbackData(preferredCurrenciesCallbackData)
	if err != nil {
		return nil, err
	}
	inlineKeyboardButtons := [][]telegram.InlineKeyboardButton{
		{
			telegram.InlineKeyboardButton{
				Text: utils.ButtonTitle(localizer.LocalizedString("balance"), "💰"),
				Data: balanceData,
			},
			telegram.InlineKeyboardButton{
				Text: utils.ButtonTitle(localizer.LocalizedString("buy_number"), "🛒"),
				Data: buyNumberData,
			},
		},
		{
			telegram.InlineKeyboardButton{
				Text: utils.ButtonTitle(localizer.LocalizedString("help"), "❓"),
				Data: helpData,
			},
			telegram.InlineKeyboardButton{
				Text: utils.ButtonTitle(localizer.LocalizedString("history"), "📖"),
				Data: historyData,
			},
		},
		{
			telegram.InlineKeyboardButton{
				Text: utils.ButtonTitle(localizer.LocalizedString("language"), "🗣️"),
				Data: languageData,
			},
			telegram.InlineKeyboardButton{
				Text: utils.ButtonTitle(localizer.LocalizedString("currency"), "💵"),
				Data: preferredCurrenciesData,
			},
		},
	}
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboardButtons,
	}, nil
}

func (b *botController) getMenuInlineKeyboardMarkup(langTag string) (*telegram.InlineKeyboardMarkup, error) {
	backToMainMenuInlineKeyboardButton, err := b.getMenuInlineKeyboardButton(langTag)
	if err != nil {
		return nil, err
	}
	inlineKeyboardButtons := [][]telegram.InlineKeyboardButton{
		{
			*backToMainMenuInlineKeyboardButton,
		},
	}
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboardButtons,
	}, nil
}

func (b *botController) getInlineKeyboardMarkupWithMainMenuButton(
	langTag string,
	extraKeyboardButtons []telegram.InlineKeyboardButton,
	columns int,
) (*telegram.InlineKeyboardMarkup, error) {
	mainMenuKeyboardButton, err := b.getMenuInlineKeyboardButton(langTag)
	if err != nil {
		return nil, err
	}
	allKeyboardButtons := extraKeyboardButtons
	allKeyboardButtons = append(allKeyboardButtons, *mainMenuKeyboardButton)
	gridKeyboardButtons := b.prepareGridInlineKeyboardButton(allKeyboardButtons, columns)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridKeyboardButtons,
	}, nil
}

func (b *botController) getMenuInlineKeyboardButton(langTag string) (*telegram.InlineKeyboardButton, error) {
	localizer := b.container.GetLocalizer(langTag)
	backToMenuTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.MainMenuCallbackQueryCmdText,
		Parameters: nil,
	}
	backToMainData, err := utils.EncodeTelegramCallbackData(backToMenuTelegramCallbackData)
	if err != nil {
		return nil, err
	}
	return &telegram.InlineKeyboardButton{
		Text: localizer.LocalizedString("back_to_main_menu"),
		Data: backToMainData,
	}, nil
}

func (b *botController) getServiceWithCountryInlineKeyboardMarkup(langTag string, pagination *app.Pagination[sms.ServicePrice], countries []sms.Country) (*telegram.InlineKeyboardMarkup, error) {
	inlineKeyboardButtons := make([]telegram.InlineKeyboardButton, 0, pagination.Len())
	startIndex := pagination.CurrentPage * pagination.ItemsPerPage
	if startIndex > pagination.Len() {
		return nil, app.IndexOutOfRangeError
	}
	endIndex := startIndex + pagination.ItemsPerPage
	if endIndex > pagination.Len() {
		endIndex = pagination.Len()
	}
	dataSourceSlice := pagination.DataSource[startIndex:endIndex]
	var serviceCode string
	if len(dataSourceSlice) > 0 {
		serviceCode = dataSourceSlice[0].Code
	} else {
		return nil, app.NilError
	}
	for _, servicePrice := range dataSourceSlice {
		filteredCountries := utils.Filter(countries, func(country sms.Country) bool {
			return country.Id == int64(servicePrice.CountryCode)
		})
		if len(filteredCountries) == 0 {
			continue
		}
		country := filteredCountries[0]
		language := b.container.GetConfig().LanguageByName(country.Title)
		serviceCountry := country.Title
		if language != nil {
			serviceCountry = fmt.Sprintf("%s %s", language.FlagEmoji, country.Title)
		}
		representableText := fmt.Sprintf("%s | %.2f ₽ | %d", serviceCountry, servicePrice.Cost, servicePrice.Count)
		parameters := []any{servicePrice.Code, country.Id, servicePrice.Cost}
		telegramCallbackData := app.TelegramCallbackData{
			Name:       app.SelectSMSServiceWithCountryCallbackQueryCmdText,
			Parameters: &parameters,
		}
		data, err := utils.EncodeTelegramCallbackData(telegramCallbackData)
		if err != nil {
			continue
		}
		inlineKeyboardButtons = append(inlineKeyboardButtons, telegram.InlineKeyboardButton{
			Text: representableText,
			Data: data,
		})
	}
	mainMenuInlineKeyboardButton, err := b.getMenuInlineKeyboardButton(langTag)
	if err != nil {
		return nil, err
	}
	gridInlineKeyboardButtons := b.prepareGridInlineKeyboardButton(inlineKeyboardButtons, 1)
	pageControlsButtons, err := getPageControlInlineKeyboardButtons(
		pagination,
		app.SelectSMSServiceCallbackQueryCmdText,
		[]any{serviceCode},
	)
	if err != nil {
		return nil, err
	}
	gridInlineKeyboardButtons = append(gridInlineKeyboardButtons, pageControlsButtons)
	gridInlineKeyboardButtons = append(gridInlineKeyboardButtons, []telegram.InlineKeyboardButton{
		*mainMenuInlineKeyboardButton,
	})
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridInlineKeyboardButtons,
	}, nil
}

func (b *botController) getPayCurrenciesInlineKeyboardMarkup(langTag string) (*telegram.InlineKeyboardMarkup, error) {
	payCurrencies := b.container.GetConfig().AvailablePayCurrencies()
	keyboardButtons := make([]telegram.InlineKeyboardButton, 0, len(payCurrencies))
	for _, currency := range payCurrencies {
		parameters := []any{currency.ABBR}
		currencyCallbackData := app.TelegramCallbackData{
			Name:       app.SelectPayCurrencyCallbackQueryCmdText,
			Parameters: &parameters,
		}
		data, err := utils.EncodeTelegramCallbackData(currencyCallbackData)
		if err != nil {
			continue
		}
		keyboardButton := telegram.InlineKeyboardButton{
			Text: utils.ShortCurrencyTextFormat(currency),
			Data: data,
		}
		keyboardButtons = append(keyboardButtons, keyboardButton)
	}
	mainMenuKeyboardButton, err := b.getMenuInlineKeyboardButton(langTag)
	if err != nil {
		return nil, err
	}
	keyboardButtons = append(keyboardButtons, *mainMenuKeyboardButton)
	gridKeyboardButtons := b.prepareGridInlineKeyboardButton(keyboardButtons, 2)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridKeyboardButtons,
	}, nil
}

func (b *botController) getCryptoPayBotKeyboardMarkup(langTag string, url string) (*telegram.InlineKeyboardMarkup, error) {
	localizer := b.container.GetLocalizer(langTag)
	payKeyboardButton := telegram.InlineKeyboardButton{
		Text: localizer.LocalizedString("pay"),
		URL:  &url,
	}
	gridKeyboardButtons := b.prepareGridInlineKeyboardButton([]telegram.InlineKeyboardButton{payKeyboardButton}, 1)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridKeyboardButtons,
	}, nil
}

func (b *botController) getPreferredCurrenciesKeyboardMarkup(langTag string) (*telegram.InlineKeyboardMarkup, error) {
	mainMenuKeyboardButton, err := b.getMenuInlineKeyboardButton(langTag)
	if err != nil {
		return nil, err
	}
	preferredCurrencies := b.container.GetConfig().AvailablePreferredCurrencies()
	keyboardButtons := make([]telegram.InlineKeyboardButton, 0, len(preferredCurrencies)+1)
	for _, preferredCurrency := range preferredCurrencies {
		parameters := []any{preferredCurrency.ABBR}
		preferredCurrencyCallbackData := app.TelegramCallbackData{
			Name:       app.SelectPreferredCurrencyCallbackQueryCmdText,
			Parameters: &parameters,
		}
		data, err := utils.EncodeTelegramCallbackData(preferredCurrencyCallbackData)
		if err != nil {
			continue
		}
		keyboardButton := telegram.InlineKeyboardButton{
			Text: utils.ShortCurrencyTextFormat(preferredCurrency),
			Data: data,
		}
		keyboardButtons = append(keyboardButtons, keyboardButton)
	}
	keyboardButtons = append(keyboardButtons, *mainMenuKeyboardButton)
	gridKeyboardButtons := b.prepareGridInlineKeyboardButton(keyboardButtons, 2)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridKeyboardButtons,
	}, nil
}

func (b *botController) prepareGridInlineKeyboardButton(keyboardButtons []telegram.InlineKeyboardButton, columns int) [][]telegram.InlineKeyboardButton {
	rows := len(keyboardButtons) / columns
	if len(keyboardButtons)%columns > 0 {
		rows += 1
	}
	gridInlineKeyboardButtons := make([][]telegram.InlineKeyboardButton, 0, rows)
	for i := 0; i < rows; i++ {
		start := i * columns
		end := start + columns
		endLimit := len(keyboardButtons)
		if end > endLimit {
			end = endLimit
		}
		gridInlineKeyboardButtons = append(gridInlineKeyboardButtons, keyboardButtons[start:end])
	}
	return gridInlineKeyboardButtons
}

func getPageControlInlineKeyboardButtons[T any](pagination *app.Pagination[T], queryCmdText string, parameters []any) ([]telegram.InlineKeyboardButton, error) {
	previousPage := pagination.CurrentPage - 1
	if previousPage < 0 {
		previousPage = pagination.Pages() - 1
	}
	nextPage := pagination.CurrentPage + 1
	if nextPage > pagination.Pages()-1 {
		nextPage = 0
	}
	previousPageParameters := make([]any, len(parameters))
	copy(previousPageParameters, parameters)
	previousPageParameters = append(previousPageParameters, previousPage)
	previousTelegramCallbackData := app.TelegramCallbackData{
		Name:       queryCmdText,
		Parameters: &previousPageParameters,
	}
	infoPageCallbackData := app.TelegramCallbackData{
		Name: app.EmptyCallbackQueryCmdText,
	}
	nextPageParameters := make([]any, len(parameters))
	copy(nextPageParameters, parameters)
	nextPageParameters = append(nextPageParameters, nextPage)
	nextTelegramCallbackData := app.TelegramCallbackData{
		Name:       queryCmdText,
		Parameters: &nextPageParameters,
	}

	previousData, err := utils.EncodeTelegramCallbackData(previousTelegramCallbackData)
	if err != nil {
		return nil, err
	}
	infoData, err := utils.EncodeTelegramCallbackData(infoPageCallbackData)
	if err != nil {
		return nil, err
	}
	nextData, err := utils.EncodeTelegramCallbackData(nextTelegramCallbackData)
	if err != nil {
		return nil, err
	}

	inlineKeyboardButtons := []telegram.InlineKeyboardButton{
		{
			Text: pagination.Previous(),
			Data: previousData,
		},
		{
			Text: pagination.MidTitle(),
			Data: infoData,
		},
		{
			Text: pagination.NextTitle(),
			Data: nextData,
		},
	}
	return inlineKeyboardButtons, nil
}
