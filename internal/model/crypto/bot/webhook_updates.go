package bot

type WebhookUpdates struct {
	ID             int      `json:"update_id"`
	UpdateType     string   `json:"update_type"`
	PayloadInvoice *Invoice `json:"payload"`
}
