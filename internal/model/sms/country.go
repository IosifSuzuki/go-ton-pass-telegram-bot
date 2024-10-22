package sms

type Country struct {
	ID           int64  `json:"id"`
	Title        string `json:"eng"`
	Visible      int    `json:"visible"`
	Retry        int    `json:"retry"`
	Rent         int    `json:"rent"`
	MultiService int    `json:"multiService"`
}
