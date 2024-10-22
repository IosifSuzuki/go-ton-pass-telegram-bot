package telegram

import "strings"

type User struct {
	ID           int64   `json:"id"`
	IsBot        bool    `json:"is_bot"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	Username     *string `json:"username,omitempty"`
	LanguageCode *string `json:"language_code,omitempty"`
}

func (u *User) FullName() string {
	firstName, lastName := u.FirstName, u.LastName
	nameComponents := make([]string, 0)
	if firstName != nil {
		nameComponents = append(nameComponents, *firstName)
	}
	if lastName != nil {
		nameComponents = append(nameComponents, *lastName)
	}
	return strings.Join(nameComponents, " ")
}
