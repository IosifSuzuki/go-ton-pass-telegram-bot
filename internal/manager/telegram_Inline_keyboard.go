package manager

import (
	"fmt"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/internal/worker"
	"go-ton-pass-telegram-bot/pkg/localizer"
	"go-ton-pass-telegram-bot/pkg/logger"
)

type TelegramInlineKeyboardManager interface {
	Set(languageTag string)
	MainMenuKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	InitialLanguagesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	LanguagesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	InitialPreferredCurrenciesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	PreferredCurrenciesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	PayCurrenciesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	MainMenuKeyboardButton() *telegram.InlineKeyboardButton
	LinkKeyboardButton(text, link string) *telegram.InlineKeyboardButton
	TopUpBalanceKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	CryptoPayBotKeyboardMarkup(url string) (*telegram.InlineKeyboardMarkup, error)
	PageControlKeyboardButtons(commandName string, pagination app.Pagination, leftButtonParameters []any, rightButtonParameters []any) ([]telegram.InlineKeyboardButton, error)
	ServicesInlineKeyboardMarkup(services []sms.Service, pagination app.Pagination) (*telegram.InlineKeyboardMarkup, error)
	ServiceCountriesInlineKeyboardMarkup(serviceCode string, preferredCurrency string, pagination app.Pagination, servicePrices []sms.PriceForService, countries []sms.Country) (*telegram.InlineKeyboardMarkup, error)
}

type telegramInlineKeyboardManager struct {
	container          container.Container
	localizer          localizer.Localizer
	formatterWorker    worker.Formatter
	exchangeRateWorker worker.ExchangeRate
}

func NewTelegramInlineKeyboardManager(container container.Container, exchangeRateWorker worker.ExchangeRate) TelegramInlineKeyboardManager {
	return &telegramInlineKeyboardManager{
		container:          container,
		localizer:          container.GetLocalizer("en"),
		formatterWorker:    worker.NewFormatter(container),
		exchangeRateWorker: exchangeRateWorker,
	}
}

func (t *telegramInlineKeyboardManager) Set(languageTag string) {
	t.localizer = t.container.GetLocalizer(languageTag)
}

func (t *telegramInlineKeyboardManager) MainMenuKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
	balanceInlineKeyboardButton, err := NewTelegramInlineButtonBuilder().
		SetText(utils.ButtonTitle(t.localizer.LocalizedString("balance"), "💰")).
		SetCommandName(app.BalanceCallbackQueryCmdText).
		Build()
	if err != nil {
		return nil, err
	}
	buyNumberInlineKeyboardButton, err := NewTelegramInlineButtonBuilder().
		SetText(utils.ButtonTitle(t.localizer.LocalizedString("buy_number"), "🛒")).
		SetCommandName(app.BuyNumberCallbackQueryCmdText).
		SetParameters([]any{0}).
		Build()
	if err != nil {
		return nil, err
	}
	helpInlineKeyboardButton, err := NewTelegramInlineButtonBuilder().
		SetText(utils.ButtonTitle(t.localizer.LocalizedString("help"), "❓")).
		SetCommandName(app.HelpCallbackQueryCmdText).
		Build()
	if err != nil {
		return nil, err
	}
	historyInlineKeyboardButton, err := NewTelegramInlineButtonBuilder().
		SetText(utils.ButtonTitle(t.localizer.LocalizedString("history"), "📖")).
		SetCommandName(app.HistoryCallbackQueryCmdText).
		SetParameters([]any{0, 3}).
		Build()
	if err != nil {
		return nil, err
	}
	languageInlineKeyboardButton, err := NewTelegramInlineButtonBuilder().
		SetText(utils.ButtonTitle(t.localizer.LocalizedString("language"), "🗣️")).
		SetCommandName(app.LanguageCallbackQueryCmdText).
		Build()
	if err != nil {
		return nil, err
	}
	preferredCurrenciesInlineKeyboardButton, err := NewTelegramInlineButtonBuilder().
		SetText(utils.ButtonTitle(t.localizer.LocalizedString("currency"), "💵")).
		SetCommandName(app.PreferredCurrenciesCallbackQueryCmdText).
		Build()
	if err != nil {
		return nil, err
	}
	inlineKeyboardButtons := t.getGridInlineKeyboardButton([]telegram.InlineKeyboardButton{
		*balanceInlineKeyboardButton, *buyNumberInlineKeyboardButton,
		*helpInlineKeyboardButton, *historyInlineKeyboardButton,
		*languageInlineKeyboardButton, *preferredCurrenciesInlineKeyboardButton,
	}, 2)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboardButtons,
	}, nil
}

func (t *telegramInlineKeyboardManager) InitialPreferredCurrenciesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
	return t.preparePreferredCurrenciesKeyboardMarkup(app.SelectInitialPreferredCurrencyCallbackQueryCmdText, false)
}

func (t *telegramInlineKeyboardManager) PreferredCurrenciesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
	return t.preparePreferredCurrenciesKeyboardMarkup(app.SelectPreferredCurrencyCallbackQueryCmdText, true)
}

func (t *telegramInlineKeyboardManager) PayCurrenciesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
	payCurrencies := t.container.GetConfig().AvailablePayCurrencies()
	payCurrenciesInlineKeyboardButtons := make([]telegram.InlineKeyboardButton, 0, len(payCurrencies))
	for _, payCurrency := range payCurrencies {
		payCurrencyInlineKeyboardButton, err := NewTelegramInlineButtonBuilder().
			SetCommandName(app.SelectPayCurrencyCallbackQueryCmdText).
			SetParameters([]any{payCurrency.ABBR}).
			SetText(utils.ShortCurrencyTextFormat(payCurrency)).
			Build()
		if err != nil {
			continue
		}
		payCurrenciesInlineKeyboardButtons = append(payCurrenciesInlineKeyboardButtons, *payCurrencyInlineKeyboardButton)
	}
	payCurrenciesInlineKeyboardButtons = append(payCurrenciesInlineKeyboardButtons, *t.MainMenuKeyboardButton())
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: t.getGridInlineKeyboardButton(payCurrenciesInlineKeyboardButtons, 2),
	}, nil
}

func (t *telegramInlineKeyboardManager) ServicesInlineKeyboardMarkup(services []sms.Service, pagination app.Pagination) (*telegram.InlineKeyboardMarkup, error) {
	log := t.container.GetLogger()
	columns := 2
	startIndex := pagination.CurrentPage * pagination.ItemsPerPage
	endIndex := (pagination.CurrentPage + 1) * pagination.ItemsPerPage
	if endIndex > len(services) {
		endIndex = len(services)
	}
	if startIndex > len(services) {
		return nil, app.IndexOutOfRangeError
	}
	servicesSlice := services[startIndex:endIndex]
	buttons := make([]telegram.InlineKeyboardButton, 0, len(servicesSlice))
	for _, service := range servicesSlice {
		button, err := NewTelegramInlineButtonBuilder().
			SetCommandName(app.SelectSMSServiceCallbackQueryCmdText).
			SetText(t.formatterWorker.Service(&service, worker.DefaultFormatterType)).
			SetParameters([]any{service.Code, 0}).
			Build()
		if err != nil {
			log.Debug("fail to create inline button for telegram", logger.FError(err))
			continue
		}
		buttons = append(buttons, *button)
	}
	gridButtons := t.getGridInlineKeyboardButton(buttons, columns)
	pageControlButtons, err := t.PageControlKeyboardButtons(app.SelectSMSServiceCallbackQueryCmdText, pagination, []any{pagination.PrevPage()}, []any{pagination.PrevPage()})
	if err != nil {
		return nil, err
	}
	gridButtons = append(gridButtons, pageControlButtons)
	goToMainMenuButton := t.MainMenuKeyboardButton()
	gridButtons = append(gridButtons, []telegram.InlineKeyboardButton{*goToMainMenuButton})
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridButtons,
	}, nil
}

func (t *telegramInlineKeyboardManager) InitialLanguagesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
	return t.prepareLanguageKeyboardMarkup(app.SelectInitialLanguageCallbackQueryCmdText, false)
}

func (t *telegramInlineKeyboardManager) LanguagesKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
	return t.prepareLanguageKeyboardMarkup(app.SelectLanguageCallbackQueryCmdText, true)
}

func (t *telegramInlineKeyboardManager) ServiceCountriesInlineKeyboardMarkup(
	serviceCode string,
	preferredCurrency string,
	pagination app.Pagination,
	servicePrices []sms.PriceForService,
	countries []sms.Country,
) (*telegram.InlineKeyboardMarkup, error) {
	log := t.container.GetLogger()
	columns := 2
	startIndex := pagination.CurrentPage * pagination.ItemsPerPage
	endIndex := (pagination.CurrentPage + 1) * pagination.ItemsPerPage
	if endIndex > len(servicePrices) {
		endIndex = len(servicePrices)
	}
	if startIndex > len(servicePrices) {
		return nil, app.IndexOutOfRangeError
	}
	servicePricesSlice := servicePrices[startIndex:endIndex]
	buttons := make([]telegram.InlineKeyboardButton, 0, len(servicePricesSlice))
	for _, servicePrice := range servicePricesSlice {
		filteredCountries := utils.Filter(countries, func(country sms.Country) bool {
			return country.ID == servicePrice.CountryCode
		})
		if len(filteredCountries) == 0 {
			log.Debug("can't find country by country id", logger.F("country_code", servicePrice.CountryCode))
			continue
		}
		country := filteredCountries[0]
		priceInRUB := servicePrice.RetailPrice
		serviceCountry := t.formatterWorker.Country(&country, worker.DefaultFormatterType)
		priceInPreferredCurrency, err := t.exchangeRateWorker.ConvertFromRUB(priceInRUB, preferredCurrency)
		if err != nil {
			log.Debug("can't convert amount from rub", logger.F("to_currency", preferredCurrency), logger.FError(err))
			continue
		}
		priceWithFee := t.exchangeRateWorker.PriceWithFee(*priceInPreferredCurrency)
		currency := t.container.GetConfig().CurrencyByAbbr(preferredCurrency)
		representableText := fmt.Sprintf("%s | %s",
			serviceCountry,
			utils.CurrencyAmountTextFormat(priceWithFee, *currency),
		)
		button, err := NewTelegramInlineButtonBuilder().
			SetText(representableText).
			SetCommandName(app.PayServiceCallbackQueryCmdText).
			SetParameters([]any{serviceCode, country.ID, *priceInPreferredCurrency}).
			Build()
		if err != nil {
			log.Debug("can't crete button with price service", logger.FError(err))
			continue
		}
		buttons = append(buttons, *button)
	}
	gridButtons := t.getGridInlineKeyboardButton(buttons, columns)
	pageControlButtons, err := t.PageControlKeyboardButtons(
		app.SelectSMSServiceCallbackQueryCmdText,
		pagination,
		[]any{serviceCode, pagination.PrevPage()},
		[]any{serviceCode, pagination.PrevPage()},
	)
	if err != nil {
		log.Debug("fail to create control keyboard buttons", logger.FError(err))
	}
	gridButtons = append(gridButtons, pageControlButtons)
	goToMainMenuButton := t.MainMenuKeyboardButton()
	gridButtons = append(gridButtons, []telegram.InlineKeyboardButton{*goToMainMenuButton})
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridButtons,
	}, nil
}

func (t *telegramInlineKeyboardManager) CryptoPayBotKeyboardMarkup(url string) (*telegram.InlineKeyboardMarkup, error) {
	log := t.container.GetLogger()
	columns := 1
	linkButton := t.LinkKeyboardButton(t.localizer.LocalizedString("pay"), url)
	cancelInvoiceButton, err := NewTelegramInlineButtonBuilder().
		SetCommandName(app.CancelInvoiceQueryCmdText).
		SetText(utils.ButtonTitle(t.localizer.LocalizedString("cancel_invoice"), "❌")).
		SetParameters([]any{}).
		Build()
	if err != nil {
		log.Debug("fail to create cancel invoice button", logger.FError(err))
		return nil, err
	}
	gridButtons := t.getGridInlineKeyboardButton([]telegram.InlineKeyboardButton{*linkButton, *cancelInvoiceButton}, columns)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridButtons,
	}, nil
}

func (t *telegramInlineKeyboardManager) TopUpBalanceKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
	log := t.container.GetLogger()
	columns := 1
	topUpBalanceButton, err := NewTelegramInlineButtonBuilder().
		SetText(t.localizer.LocalizedString("top_up_balance")).
		SetCommandName(app.ListPayCurrenciesCallbackQueryCmdText).
		Build()
	if err != nil {
		log.Debug("fail to create top up balance button", logger.FError(err))
		return nil, err
	}
	goToMainMenuButton := t.MainMenuKeyboardButton()
	gridButtons := t.getGridInlineKeyboardButton([]telegram.InlineKeyboardButton{*topUpBalanceButton, *goToMainMenuButton}, columns)
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: gridButtons,
	}, nil
}

func (t *telegramInlineKeyboardManager) MainMenuKeyboardButton() *telegram.InlineKeyboardButton {
	mainMenuInlineKeyboardButton, _ := NewTelegramInlineButtonBuilder().
		SetText(t.localizer.LocalizedString("back_to_main_menu")).
		SetCommandName(app.MainMenuCallbackQueryCmdText).
		Build()
	return mainMenuInlineKeyboardButton
}

func (t *telegramInlineKeyboardManager) PageControlKeyboardButtons(commandName string, pagination app.Pagination, leftButtonParameters []any, rightButtonParameters []any) ([]telegram.InlineKeyboardButton, error) {
	prevButton, err := NewTelegramInlineButtonBuilder().
		SetText(pagination.PreviousTitle()).
		SetCommandName(commandName).
		SetParameters(leftButtonParameters).
		Build()
	if err != nil {
		return nil, err
	}
	currentPageButton, err := NewTelegramInlineButtonBuilder().
		SetText(pagination.MidTitle()).
		SetCommandName(app.EmptyCallbackQueryCmdText).
		Build()
	if err != nil {
		return nil, err
	}
	nextButton, err := NewTelegramInlineButtonBuilder().
		SetText(pagination.NextTitle()).
		SetCommandName(commandName).
		SetParameters(rightButtonParameters).
		Build()
	if err != nil {
		return nil, err
	}
	inlineKeyboardButtons := []telegram.InlineKeyboardButton{
		*prevButton,
		*currentPageButton,
		*nextButton,
	}
	return inlineKeyboardButtons, nil
}

func (t *telegramInlineKeyboardManager) LinkKeyboardButton(text, link string) *telegram.InlineKeyboardButton {
	linkInlineKeyboardButton, _ := NewTelegramInlineButtonBuilder().
		SetLink(link).
		SetText(text).
		Build()
	return linkInlineKeyboardButton
}

func (t *telegramInlineKeyboardManager) getGridInlineKeyboardButton(keyboardButtons []telegram.InlineKeyboardButton, columns int) [][]telegram.InlineKeyboardButton {
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

func (t *telegramInlineKeyboardManager) prepareLanguageKeyboardMarkup(commandName string, shouldContainsMainMenu bool) (*telegram.InlineKeyboardMarkup, error) {
	languages := t.container.GetConfig().AvailableLanguages()
	buttons := make([]telegram.InlineKeyboardButton, 0, len(languages))
	for _, language := range languages {
		button, err := NewTelegramInlineButtonBuilder().
			SetCommandName(commandName).
			SetParameters([]any{language.Code}).
			SetText(utils.LanguageTextFormat(language)).
			Build()
		if err != nil {
			continue
		}
		buttons = append(buttons, *button)
	}
	if shouldContainsMainMenu {
		buttons = append(buttons, *t.MainMenuKeyboardButton())
	}
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: t.getGridInlineKeyboardButton(buttons, 2),
	}, nil
}

func (t *telegramInlineKeyboardManager) preparePreferredCurrenciesKeyboardMarkup(commandName string, shouldContainsMainMenu bool) (*telegram.InlineKeyboardMarkup, error) {
	preferredCurrencies := t.container.GetConfig().AvailablePreferredCurrencies()
	buttons := make([]telegram.InlineKeyboardButton, 0, len(preferredCurrencies))
	for _, currency := range preferredCurrencies {
		button, err := NewTelegramInlineButtonBuilder().
			SetCommandName(commandName).
			SetParameters([]any{currency.ABBR}).
			SetText(utils.ShortCurrencyTextFormat(currency)).
			Build()
		if err != nil {
			continue
		}
		buttons = append(buttons, *button)
	}
	if shouldContainsMainMenu {
		buttons = append(buttons, *t.MainMenuKeyboardButton())
	}
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: t.getGridInlineKeyboardButton(buttons, 2),
	}, nil
}