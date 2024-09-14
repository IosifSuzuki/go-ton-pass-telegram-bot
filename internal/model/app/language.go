package app

type Language struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"nativeName"`
	FlagEmoji  string `json:"flagEmoji"`
}
