package app

type SendTelegramStarsInvoice struct {
	ChatID      int64
	Title       string
	Description string
	Stars       int64
	Payload     any
	PhotoURL    string
}
