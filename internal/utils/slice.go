package utils

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func Contains[T any](ss []T, test func(T) bool) (ret bool) {
	for _, s := range ss {
		if test(s) {
			return true
		}
	}
	return false
}
