package app

type ExchangeRate struct {
	SourceCurrency string  `json:"source_currency"`
	TargetCurrency string  `json:"target_currency"`
	Rate           float64 `json:"rate"`
}

type ExchangeRateResponse struct {
	Result []ExchangeRate `json:"result"`
}
