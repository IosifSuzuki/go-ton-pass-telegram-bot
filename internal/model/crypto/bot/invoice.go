package bot

type Invoice struct {
	ID            int64   `json:"invoice_id"`
	Hash          string  `json:"hash"`
	CurrencyType  string  `json:"currency_type"`
	Asset         *string `json:"asset"`
	Fiat          *string `json:"fiat"`
	Amount        string  `json:"amount"`
	PaidAsset     *string `json:"paid_asset"`
	PaidFiat      *string `json:"paid_fiat"`
	FeeAsset      *string `json:"fee_asset"`
	FeeAmount     *string `json:"fee_amount"`
	BotInvoiceURL string  `json:"bot_invoice_url"`
	PaidUsdRate   *string `json:"paid_usd_rate"`
	Status        string  `json:"status"`
	PaidAt        string  `json:"paid_at"`
	Payload       *string `json:"payload"`
}
