package telegram

type SendResponse struct {
	ChatId int64  `json:"chat_id"`
	Text   string `json:"text"`
}
