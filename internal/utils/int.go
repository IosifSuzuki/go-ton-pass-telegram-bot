package utils

func GetInt64(value any) int64 {
	switch value := value.(type) {
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	}
	return 0
}
