package telegram

import "strings"

type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

func (u *User) FullName() string {
	return strings.Join([]string{u.FirstName, u.LastName}, " ")
}
