package telegram

type ReplyKeyboardMarkup struct {
	Keyboard                  [][]KeyboardButton `json:"keyboard"`
	PersistentDisplayKeyboard bool               `json:"is_persistent"`
	ResizeKeyboard            bool               `json:"resize_keyboard"`
	OneTimeKeyboard           bool               `json:"one_time_keyboard"`
	Placeholder               *string            `json:"input_field_placeholder,omitempty"`
}
