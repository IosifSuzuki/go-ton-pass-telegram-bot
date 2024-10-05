package app

import "time"

type CacheResponse[T any] struct {
	Result      T         `json:"result"`
	TimeFetched time.Time `json:"time_fetched"`
}
