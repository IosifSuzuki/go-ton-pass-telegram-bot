package sms

type ServicePrice struct {
	CountryCode int
	Code        string
	Cost        float64 `json:"cost"`
	Count       int     `json:"count"`
}
