package manager

import (
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/localizer"
)

type TelegramKeyboardManager interface {
	Set(languageTag string)
	MainMenuInlineKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	LanguagesInlineKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	PayCurrenciesInlineKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error)
	MainMenuInlineKeyboardButton() *telegram.InlineKeyboardButton
	LinkInlineKeyboardButton(text, link string) *telegram.InlineKeyboardButton
}

type telegramKeyboardManager struct {
	container container.Container
	localizer localizer.Localizer
}

func NewTelegramKeyboardManager(container container.Container) TelegramKeyboardManager {
	return &telegramKeyboardManager{
		container: container,
		localizer: container.GetLocalizer("en"),
	}
}

func (t *telegramKeyboardManager) Set(languageTag string) {
	t.localizer = t.container.GetLocalizer(languageTag)
}

func (t *telegramKeyboardManager) MainMenuInlineKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
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

func (t *telegramKeyboardManager) PayCurrenciesInlineKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
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
	payCurrenciesInlineKeyboardButtons = append(payCurrenciesInlineKeyboardButtons, *t.MainMenuInlineKeyboardButton())
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: t.getGridInlineKeyboardButton(payCurrenciesInlineKeyboardButtons, 2),
	}, nil
}

func (t *telegramKeyboardManager) LanguagesInlineKeyboardMarkup() (*telegram.InlineKeyboardMarkup, error) {
	languages := t.container.GetConfig().AvailableLanguages()
	languageInlineKeyboardButtons := make([]telegram.InlineKeyboardButton, 0, len(languages))
	for _, language := range languages {
		languageInlineKeyboardButton, err := NewTelegramInlineButtonBuilder().
			SetCommandName(app.SelectLanguageCallbackQueryCmdText).
			SetParameters([]any{language.Code}).
			SetText(utils.LanguageTextFormat(language)).
			Build()
		if err != nil {
			continue
		}
		languageInlineKeyboardButtons = append(languageInlineKeyboardButtons, *languageInlineKeyboardButton)
	}
	languageInlineKeyboardButtons = append(languageInlineKeyboardButtons, *t.MainMenuInlineKeyboardButton())
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: t.getGridInlineKeyboardButton(languageInlineKeyboardButtons, 2),
	}, nil
}

func (t *telegramKeyboardManager) MainMenuInlineKeyboardButton() *telegram.InlineKeyboardButton {
	mainMenuInlineKeyboardButton, _ := NewTelegramInlineButtonBuilder().
		SetText(t.localizer.LocalizedString("back_to_main_menu")).
		SetCommandName(app.MainMenuCallbackQueryCmdText).
		Build()
	return mainMenuInlineKeyboardButton
}

func (t *telegramKeyboardManager) LinkInlineKeyboardButton(text, link string) *telegram.InlineKeyboardButton {
	linkInlineKeyboardButton, _ := NewTelegramInlineButtonBuilder().
		SetLink(link).
		SetText(text).
		Build()
	return linkInlineKeyboardButton
}

func (t *telegramKeyboardManager) getGridInlineKeyboardButton(keyboardButtons []telegram.InlineKeyboardButton, columns int) [][]telegram.InlineKeyboardButton {
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
