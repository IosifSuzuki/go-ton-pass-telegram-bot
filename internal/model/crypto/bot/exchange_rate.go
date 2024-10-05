package bot

type ExchangeRate struct {
	IsValid  bool   `json:"is_valid"`
	IsCrypto bool   `json:"is_crypto"`
	IsFiat   bool   `json:"is_fiat"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Rate     string `json:"rate"`
}
