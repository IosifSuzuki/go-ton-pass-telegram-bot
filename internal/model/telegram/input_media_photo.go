package telegram

type InputPhotoMedia struct {
	Type                  string  `json:"type"`
	Media                 string  `json:"media"`
	Caption               *string `json:"caption,omitempty"`
	ParseMode             *string `json:"parse_mode,omitempty"`
	ShowCaptionAboveMedia *bool   `json:"show_caption_above_media,omitempty"`
}
