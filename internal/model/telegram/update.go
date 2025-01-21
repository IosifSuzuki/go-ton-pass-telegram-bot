package telegram

type Update struct {
	ID            int64          `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

func (u *Update) GetChatID() int64 {
	if u.Message != nil {
		return u.Message.Chat.ID
	}
	return u.CallbackQuery.Message.Chat.ID
}

func (u *Update) GetTelegramID() int64 {
	if u.Message != nil {
		return u.Message.From.ID
	}
	return u.CallbackQuery.Message.From.ID
}
