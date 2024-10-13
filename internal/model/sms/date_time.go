package sms

import (
	"errors"
	"fmt"
	"time"
)

type Datetime time.Time

func (d *Datetime) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return errors.New("not a json string")
	}
	b = b[1 : len(b)-1]
	t, err := time.Parse("2006-01-02 15:04:05", string(b))
	if err != nil {
		return fmt.Errorf("failed to parse time: %w", err)
	}
	*d = Datetime(t)
	return nil
}
