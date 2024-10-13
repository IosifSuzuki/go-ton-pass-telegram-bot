package utils

func GetFloat64(value any) float64 {
	switch value := value.(type) {
	case float32:
		return float64(value)
	case float64:
		return value
	}
	return 0
}
