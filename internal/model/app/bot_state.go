package app

type BotState uint

const (
	IDLEState BotState = iota
	SelectLanguageBotState
	SelectCurrencyBotState
)
