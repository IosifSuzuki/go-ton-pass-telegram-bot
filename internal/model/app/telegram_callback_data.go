package app

const (
	SelectInitialLanguageCallbackQueryCmdText          = "s_i_lang"
	SelectInitialPreferredCurrencyCallbackQueryCmdText = "s_i_pref_curr"
	BalanceCallbackQueryCmdText                        = "bal"
	BuyNumberCallbackQueryCmdText                      = "buy_num"
	HistoryCallbackQueryCmdText                        = "hist"
	HelpCallbackQueryCmdText                           = "help"
	LanguageCallbackQueryCmdText                       = "lang"
	MainMenuCallbackQueryCmdText                       = "menu"
	SelectSMSServiceCallbackQueryCmdText               = "s_serv"
	PayServiceCallbackQueryCmdText                     = "s_serv_count"
	SelectLanguageCallbackQueryCmdText                 = "s_lang"
	ListPayCurrenciesCallbackQueryCmdText              = "l_pay_curr"
	SelectPayCurrencyCallbackQueryCmdText              = "s_pay_curr"
	PreferredCurrenciesCallbackQueryCmdText            = "pre_curr"
	SelectPreferredCurrencyCallbackQueryCmdText        = "s_pref_curr"
	EmptyCallbackQueryCmdText                          = "empty"
	CancelInvoiceQueryCmdText                          = "c_inv"
)

type TelegramCallbackData struct {
	Name       string `msgpack:"n"`
	Parameters *[]any `msgpack:"p"`
}

func (t *TelegramCallbackData) CallbackQueryCommand() CallbackQueryCommand {
	switch t.Name {
	case SelectInitialLanguageCallbackQueryCmdText:
		return SelectInitialLanguageCallbackQueryCommand
	case SelectInitialPreferredCurrencyCallbackQueryCmdText:
		return SelectInitialPreferredCurrencyCallbackQueryCommand
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
	case PayServiceCallbackQueryCmdText:
		return PayServiceCallbackQueryCommand
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
	case EmptyCallbackQueryCmdText:
		return EmptyCallbackQueryCommand
	default:
		return NotCallbackQueryCommand
	}
}
