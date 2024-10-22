package telegram

type GetChatMember struct {
	ChatID any   `json:"chat_id"`
	UserID int64 `json:"user_id"`
}
