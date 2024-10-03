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
		presentableLanguageText := fmt.Sprintf("%s %s", language.FlagEmoji, language.NativeName)
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

func (b *botController) getServicesInlineKeyboardMarkup(ctx context.Context, callbackQuery *telegram.CallbackQuery, smsServices []sms.Service) (*telegram.InlineKeyboardMarkup, error) {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to getLanguageCode", logger.FError(err))
		return nil, err
	}
	localizer := b.container.GetLocalizer(*langTag)
	inlineKeyboardButtons := make([]telegram.InlineKeyboardButton, 0, len(smsServices))
	lengthSmsServices := len(smsServices)
	if lengthSmsServices > 10 {
		lengthSmsServices = 10
	}
	for _, smsService := range smsServices[:lengthSmsServices] {
		parameters := []any{smsService.Code}
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
			Text: smsService.Name,
			Data: data,
		})
	}
	backToMenuTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.MainMenuCallbackQueryCmdText,
		Parameters: nil,
	}
	backToMainData, err := utils.EncodeTelegramCallbackData(backToMenuTelegramCallbackData)
	if err != nil {
		log.Error("fail to encode telegram callback data", logger.FError(err))
		return nil, err
	}
	inlineKeyboardButtons = append(inlineKeyboardButtons, telegram.InlineKeyboardButton{
		Text: localizer.LocalizedString("back_to_main_menu"),
		Data: backToMainData,
	})
	gridInlineKeyboardButtons := b.prepareGridInlineKeyboardButton(inlineKeyboardButtons, 2)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridInlineKeyboardButtons,
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

func (b *botController) getMainMenuInlineKeyboardMarkup(ctx context.Context, user telegram.User) (*telegram.InlineKeyboardMarkup, error) {
	langTag, err := b.getLanguageCode(ctx, user)
	if err != nil {
		return nil, err
	}
	localizer := b.container.GetLocalizer(*langTag)
	balanceTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.BalanceCallbackQueryCmdText,
		Parameters: nil,
	}
	buyNumberTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.BuyNumberCallbackQueryCmdText,
		Parameters: nil,
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
	inlineKeyboardButtons := [][]telegram.InlineKeyboardButton{
		{
			telegram.InlineKeyboardButton{
				Text: localizer.LocalizedString("balance"),
				Data: balanceData,
			},
			telegram.InlineKeyboardButton{
				Text: localizer.LocalizedString("buy_number"),
				Data: buyNumberData,
			},
		},
		{
			telegram.InlineKeyboardButton{
				Text: localizer.LocalizedString("help"),
				Data: helpData,
			},
			telegram.InlineKeyboardButton{
				Text: localizer.LocalizedString("history"),
				Data: historyData,
			},
			telegram.InlineKeyboardButton{
				Text: localizer.LocalizedString("language"),
				Data: languageData,
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

func (b *botController) getServicePricesInlineKeyboardMarkup(langTag string, servicePrices []sms.ServicePrice, countries []sms.Country) (*telegram.InlineKeyboardMarkup, error) {
	keyboardButtons := make([]telegram.InlineKeyboardButton, 0, len(servicePrices))
	lengthServicePrices := len(servicePrices)
	if lengthServicePrices > 10 {
		lengthServicePrices = 10
	}
	for _, servicePrice := range servicePrices[:lengthServicePrices] {
		filteredCountries := utils.Filter(countries, func(country sms.Country) bool {
			return country.Id == servicePrice.CountryCode
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
		parameters := []any{servicePrice.Code}
		telegramCallbackData := app.TelegramCallbackData{
			Name:       app.SelectSMSServiceWithPriceCallbackQueryCmdText,
			Parameters: &parameters,
		}
		data, err := utils.EncodeTelegramCallbackData(telegramCallbackData)
		if err != nil {
			continue
		}
		keyboardButtons = append(keyboardButtons, telegram.InlineKeyboardButton{
			Text: representableText,
			Data: data,
		})
	}
	menuKeyboardButton, err := b.getMenuInlineKeyboardButton(langTag)
	if err != nil {
		return nil, err
	}
	keyboardButtons = append(keyboardButtons, *menuKeyboardButton)
	gridKeyboardButtons := b.prepareGridInlineKeyboardButton(keyboardButtons, 1)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridKeyboardButtons,
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
		representableText := fmt.Sprintf("%s %s", currency.ABBR, currency.Symbol)
		keyboardButton := telegram.InlineKeyboardButton{
			Text: representableText,
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
