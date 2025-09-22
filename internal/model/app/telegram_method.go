package app

type TelegramMethod string

const (
	DeleteMessageTelegramMethod          TelegramMethod = "deleteMessage"
	SendMessageTelegramMethod            TelegramMethod = "sendMessage"
	SetMyCommandsTelegramMethod          TelegramMethod = "setMyCommands"
	SetMyDescriptionTelegramMethod       TelegramMethod = "setMyDescription"
	SetMyNameTelegramMethod              TelegramMethod = "setMyName"
	AnswerCallbackQueryTelegramMethod    TelegramMethod = "answerCallbackQuery"
	SendPhotoTelegramMethod              TelegramMethod = "sendPhoto"
	EditMessageTextTelegramMethod        TelegramMethod = "editMessageText"
	EditCaptionMessageTelegramMethod     TelegramMethod = "editMessageCaption"
	EditMessageMediaTelegramMethod       TelegramMethod = "editMessageMedia"
	GetChatMemberTelegramMethod          TelegramMethod = "getChatMember"
	SendInvoiceTelegramMethod            TelegramMethod = "sendInvoice"
	AnswerPreCheckoutQueryTelegramMethod TelegramMethod = "answerPreCheckoutQuery"
)
