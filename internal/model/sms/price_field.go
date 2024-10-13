package sms

import (
	"errors"
	"strconv"
)

type PriceFiled float64

func (p *PriceFiled) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return errors.New("not a json string")
	}
	b = b[1 : len(b)-1]
	text := string(b)
	number, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return err
	}
	*p = PriceFiled(number)
	return nil
}
