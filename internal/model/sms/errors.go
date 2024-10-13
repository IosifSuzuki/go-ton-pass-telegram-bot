package sms

import "strings"

var (
	WrongMaxPriceErrorName = "WRONG_MAX_PRICE"
	BadServiceErrorName    = "BAD_SERVICE"
	SqlErrorName           = "ERROR_SQL"
	NoNumbersErrorName     = "NO_NUMBERS"
)

type Error struct {
	Name string
}

func (e Error) Error() string {
	return e.Name
}

func DecodeError(text string) *Error {
	if strings.HasPrefix(text, WrongMaxPriceErrorName) {
		return &Error{Name: WrongMaxPriceErrorName}
	} else if strings.HasPrefix(text, BadServiceErrorName) {
		return &Error{Name: BadServiceErrorName}
	} else if strings.HasPrefix(text, SqlErrorName) {
		return &Error{Name: SqlErrorName}
	} else if strings.HasPrefix(text, NoNumbersErrorName) {
		return &Error{Name: NoNumbersErrorName}
	}
	return nil
}
