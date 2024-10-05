package app

type ExchangeRate struct {
	Currency string  `json:"abbr"`
	Rate     float64 `json:"rate"`
}

type ExchangeRateResponse struct {
	Result []ExchangeRate `json:"result"`
}
