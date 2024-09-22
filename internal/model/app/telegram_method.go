package app

type TelegramMethod string

const (
	SendMessageTelegramMethod         TelegramMethod = "sendMessage"
	SetMyCommandsTelegramMethod                      = "setMyCommands"
	SetMyDescriptionTelegramMethod                   = "setMyDescription"
	SetMyNameTelegramMethod                          = "setMyName"
	AnswerCallbackQueryTelegramMethod                = "answerCallbackQuery"
	SendPhotoTelegramMethod                          = "sendPhoto"
	EditMessageTextTelegramMethod                    = "editMessageText"
)
