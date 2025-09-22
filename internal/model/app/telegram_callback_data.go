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
	CancelEnterAmountCallbackQueryCmdText              = "c_ent_am"
	SelectSMSServiceCallbackQueryCmdText               = "s_serv"
	PayServiceCallbackQueryCmdText                     = "s_serv_count"
	SelectLanguageCallbackQueryCmdText                 = "s_lang"
	CryptoBotListPayCurrenciesCallbackQueryCmdText     = "l_pay_curr"
	SelectTelegramStarsCallbackQueryCmdText            = "s_telegram_stars"
	SelectPayCurrencyCallbackQueryCmdText              = "s_pay_curr"
	PreferredCurrenciesCallbackQueryCmdText            = "pre_curr"
	SelectPreferredCurrencyCallbackQueryCmdText        = "s_pref_curr"
	EmptyCallbackQueryCmdText                          = "empty"
	DeleteCryptoBotInvoiceQueryCmdText                 = "d_cr_b_inv"
	ConfirmationPayServiceQueryCmdText                 = "con_s_pay"
	RefundAmountFromSMSActivationQueryCmdText          = "ref_sms_act"
	BackQueryCmdText                                   = "back"
	CancelPayTelegramStarsCmdText                      = "c_pay_xtr"
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
	case CryptoBotListPayCurrenciesCallbackQueryCmdText:
		return CryptoBotListPayCurrenciesCallbackQueryCommand
	case SelectTelegramStarsCallbackQueryCmdText:
		return SelectTelegramStarsCallbackQueryCommand
	case SelectPayCurrencyCallbackQueryCmdText:
		return SelectCryptoBotPayCurrencyCallbackQueryCommand
	case PreferredCurrenciesCallbackQueryCmdText:
		return PreferredCurrenciesCallbackQueryCommand
	case SelectPreferredCurrencyCallbackQueryCmdText:
		return SelectPreferredCurrencyCallbackQueryCommand
	case EmptyCallbackQueryCmdText:
		return EmptyCallbackQueryCommand
	case DeleteCryptoBotInvoiceQueryCmdText:
		return DeleteCryptoBotInvoiceCallbackQueryCommand
	case ConfirmationPayServiceQueryCmdText:
		return ConfirmationPayServiceCallbackQueryCommand
	case RefundAmountFromSMSActivationQueryCmdText:
		return RefundAmountFromSMSActivationCallbackQueryCommand
	case BackQueryCmdText:
		return BackCallbackQueryCommand
	case CancelEnterAmountCallbackQueryCmdText:
		return CancelEnterAmountCallbackQueryCommand
	case CancelPayTelegramStarsCmdText:
		return CancelPayTelegramStarsCallbackQueryCommand
	default:
		return NotCallbackQueryCommand
	}
}
