package app

import "fmt"

type PhoneNumber struct {
	CountryCode      string
	ShortPhoneNumber string
}

func (p PhoneNumber) FullNumber() string {
	return fmt.Sprintf("+%s%s", p.CountryCode, p.ShortPhoneNumber)
}
