package telegram

type AnswerCallbackQuery struct {
	ID        string  `json:"callback_query_id"`
	Text      *string `json:"text,omitempty"`
	ShowAlert bool    `json:"show_alert"`
	CacheTime int64   `json:"cache_time"`
}
