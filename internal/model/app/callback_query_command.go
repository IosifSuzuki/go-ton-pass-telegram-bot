package app

type CallbackQueryCommand uint

const (
	NotCallbackQueryCommand CallbackQueryCommand = iota
	BalanceCallbackQueryCommand
	BuyNumberCallbackQueryCommand
	HelpCallbackQueryCommand
	HistoryCallbackQueryCommand
	LanguageCallbackQueryCommand
)
