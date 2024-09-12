package telegram

type Message struct {
	ID   int    `json:"message_id"`
	From User   `json:"from"`
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
}
