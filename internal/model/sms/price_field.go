package sms

import (
	"bytes"
	"encoding/binary"
	"strconv"
)

type PriceFiled float64

func (p *PriceFiled) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		var number float64
		buf := bytes.NewReader(b)
		err := binary.Read(buf, binary.LittleEndian, &number)
		if err != nil {
			return nil
		}
		*p = PriceFiled(number)
		return err
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
