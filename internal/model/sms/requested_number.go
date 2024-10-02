package sms

type RequestedNumber struct {
	ActivationID       string `json:"activationId"`
	PhoneNumber        string `json:"phoneNumber"`
	ActivationCost     string `json:"activationCost"`
	ActivationOperator string `json:"activationOperator"`
}
