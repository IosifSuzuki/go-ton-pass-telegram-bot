package app

type CallbackQueryCommand uint

const (
	NotCallbackQueryCommand CallbackQueryCommand = iota
	SelectInitialLanguageCallbackQueryCommand
	SelectInitialPreferredCurrencyCallbackQueryCommand
	BalanceCallbackQueryCommand
	BuyNumberCallbackQueryCommand
	HelpCallbackQueryCommand
	HistoryCallbackQueryCommand
	LanguageCallbackQueryCommand
	MainMenuCallbackQueryCommand
	SelectSMSServiceCallbackQueryCommand
	PayServiceCallbackQueryCommand
	SelectLanguageCallbackQueryCommand
	ListPayCurrenciesCallbackQueryCommand
	SelectPayCurrencyCallbackQueryCommand
	PreferredCurrenciesCallbackQueryCommand
	SelectPreferredCurrencyCallbackQueryCommand
	EmptyCallbackQueryCommand
)
