package app

type SMSActivationState string

const (
	UnknownSMSActivateState SMSActivationState = "STATUS_UNKNOWN"
	CancelSMSActivateState  SMSActivationState = "STATUS_CANCEL"
	PendingSMSActivateState SMSActivationState = "STATUS_PENDING"
	DoneSMSActivateState    SMSActivationState = "STATUS_OK"
)
