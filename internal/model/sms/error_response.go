package sms

type ErrorResponse struct {
	Message string         `json:"msg"`
	Info    map[string]any `json:"info"`
}
