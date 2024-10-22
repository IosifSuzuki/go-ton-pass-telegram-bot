package bot

type Result[T any] struct {
	OK     bool `json:"ok"`
	Result T    `json:"result"`
}
