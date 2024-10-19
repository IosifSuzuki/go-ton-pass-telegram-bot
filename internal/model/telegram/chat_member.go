package telegram

type ChatMember struct {
	Status MemberStatus `json:"status"`
	User   User         `json:"user"`
}
