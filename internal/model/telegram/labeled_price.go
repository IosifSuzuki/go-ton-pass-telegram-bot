package telegram

type LabeledPrice struct {
	Label string `json:"label"`
	Price int64  `json:"amount"`
}
