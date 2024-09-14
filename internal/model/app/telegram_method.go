package app

type TelegramMethod string

const (
	SendMessageTelegramMethod      TelegramMethod = "sendMessage"
	SetMyCommandsTelegramMethod                   = "setMyCommands"
	SetMyDescriptionTelegramMethod                = "setMyDescription"
	SetMyNameTelegramMethod                       = "setMyName"
)
