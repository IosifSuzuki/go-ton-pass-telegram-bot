package app

const (
	BalanceCallbackQueryCmdText                   = "balance"
	BuyNumberCallbackQueryCmdText                 = "buy_number"
	HistoryCallbackQueryCmdText                   = "history"
	HelpCallbackQueryCmdText                      = "help"
	LanguageCallbackQueryCmdText                  = "language"
	MainMenuCallbackQueryCmdText                  = "main_menu"
	SelectSMSServiceCallbackQueryCmdText          = "select_service"
	SelectSMSServiceWithPriceCallbackQueryCmdText = "select_service_with_price"
	SelectLanguageCallbackQueryCmdText            = "select_language"
	ListPayCurrenciesCallbackQueryCmdText         = "list_pay_currencies"
	SelectPayCurrencyCallbackQueryCmdText         = "select_pay_currency"
	PreferredCurrenciesCallbackQueryCmdText       = "preferred_currencies"
	SelectPreferredCurrencyCallbackQueryCmdText   = "select_preferred_currency"
)

type TelegramCallbackData struct {
	Name       string
	Parameters *[]any
}

func (t *TelegramCallbackData) CallbackQueryCommand() CallbackQueryCommand {
	switch t.Name {
	case BalanceCallbackQueryCmdText:
		return BalanceCallbackQueryCommand
	case BuyNumberCallbackQueryCmdText:
		return BuyNumberCallbackQueryCommand
	case HistoryCallbackQueryCmdText:
		return HistoryCallbackQueryCommand
	case HelpCallbackQueryCmdText:
		return HelpCallbackQueryCommand
	case LanguageCallbackQueryCmdText:
		return LanguageCallbackQueryCommand
	case MainMenuCallbackQueryCmdText:
		return MainMenuCallbackQueryCommand
	case SelectSMSServiceCallbackQueryCmdText:
		return SelectSMSServiceCallbackQueryCommand
	case SelectSMSServiceWithPriceCallbackQueryCmdText:
		return SelectSMSServiceWithPriceCallbackQueryCommand
	case SelectLanguageCallbackQueryCmdText:
		return SelectLanguageCallbackQueryCommand
	case ListPayCurrenciesCallbackQueryCmdText:
		return ListPayCurrenciesCallbackQueryCommand
	case SelectPayCurrencyCallbackQueryCmdText:
		return SelectPayCurrencyCallbackQueryCommand
	case PreferredCurrenciesCallbackQueryCmdText:
		return PreferredCurrenciesCallbackQueryCommand
	case SelectPreferredCurrencyCallbackQueryCmdText:
		return SelectPreferredCurrencyCallbackQueryCommand
	default:
		return NotCallbackQueryCommand
	}
}
