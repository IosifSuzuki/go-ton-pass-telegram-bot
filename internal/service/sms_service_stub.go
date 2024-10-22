package service

import (
	"go-ton-pass-telegram-bot/internal/container"
)

type smsServiceStub struct {
	smsService
}

func NewSMSServiceStub(container container.Container) SMSService {
	return &smsServiceStub{
		smsService{container: container},
	}
}
