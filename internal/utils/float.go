package utils

import (
	"regexp"
	"strconv"
)

func GetFloat64(value any) float64 {
	switch value := value.(type) {
	case float32:
		return float64(value)
	case float64:
		return value
	}
	return 0
}

func ParseFloat64FromText(text string) (float64, error) {
	re := regexp.MustCompile(`([-+]?\d*)([.,])(\d+)`)
	text = re.ReplaceAllString(text, "${1}.${3}")
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}
