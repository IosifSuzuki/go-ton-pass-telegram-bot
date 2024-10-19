package telegram

import "go-ton-pass-telegram-bot/internal/model/app"

type MemberStatus string

const (
	MemberMemberStatus        MemberStatus = "member"
	CreatorMemberStatus       MemberStatus = "creator"
	AdministratorMemberStatus MemberStatus = "administrator"
	LeftMemberStatus          MemberStatus = "left"
	BannedMemberStatus        MemberStatus = "kicked"
)

func (m *MemberStatus) UnmarshalJSON(data []byte) error {
	text := string(data)
	if text == "null" || text == `""` || len(text) <= 2 {
		return app.NilError
	}
	data = data[1 : len(data)-1]
	memberStatus := MemberStatus(data)
	switch memberStatus {
	case MemberMemberStatus, CreatorMemberStatus, AdministratorMemberStatus, LeftMemberStatus, BannedMemberStatus:
		*m = memberStatus
		return nil
	default:
		return app.UnknownValueError
	}
}
