package sms

type RequestedNumber struct {
	ActivationID       string     `json:"activationId"`
	PhoneNumber        string     `json:"phoneNumber"`
	ActivationCost     PriceFiled `json:"activationCost"`
	CountryCode        string     `json:"countryCode"`
	ActivationTime     Datetime   `json:"activationTime"`
	ActivationOperator string     `json:"activationOperator"`
}
