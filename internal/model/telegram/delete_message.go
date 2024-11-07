package telegram

type DeleteMessage struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int64 `json:"message_id"`
}
