package worker

import (
	"fmt"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/utils"
	"strings"
)

type FormatterType uint

const (
	DefaultFormatterType FormatterType = iota
)

type Formatter interface {
	Country(country *sms.Country, formatterType FormatterType) string
	Service(service *sms.Service, formatterType FormatterType) string
	SHSHistories(langCode string, smsHistories []domain.SMSHistory) string
	SMSHistory(langCode string, smsHistory domain.SMSHistory) string
}

type formatter struct {
	container container.Container
}

func NewFormatter(container container.Container) Formatter {
	f := formatter{
		container: container,
	}
	return &f
}

func (f *formatter) Country(country *sms.Country, _ FormatterType) string {
	var title string
	name := f.container.GetRepresentableCountryName(country.ID)
	if name == nil {
		name = &country.Title
	}
	flag := f.container.GetFlagEmoji(*name)
	if flag != nil {
		title = fmt.Sprintf("%s %s", *flag, *name)
	} else {
		title = country.Title
	}
	return title
}

func (f *formatter) Service(service *sms.Service, _ FormatterType) string {
	var (
		name  string
		emoji string
	)
	extraService := f.container.GetExtraService(service.Code)
	if extraService != nil {
		name = extraService.Name
		emoji = extraService.Emoji
	} else {
		name = service.Name
		emoji = "🌐"
	}
	return fmt.Sprintf("%s %s", emoji, name)
}

func (f *formatter) SHSHistories(langCode string, smsHistories []domain.SMSHistory) string {
	localizer := f.container.GetLocalizer(langCode)
	newLine := "\n"
	stringBuilder := strings.Builder{}
	stringBuilder.WriteString(localizer.LocalizedString("sms_history_title_markdown"))
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	for _, smsHistory := range smsHistories {
		historyText := f.SMSHistory(langCode, smsHistory)
		stringBuilder.WriteString(historyText)
		stringBuilder.WriteString(newLine)
		stringBuilder.WriteString(newLine)
	}
	return stringBuilder.String()
}

func (f *formatter) SMSHistory(langCode string, smsHistory domain.SMSHistory) string {
	localizer := f.container.GetLocalizer(langCode)
	newLine := "\n"
	stringBuilder := strings.Builder{}
	phoneNumberText := localizer.LocalizedStringWithTemplateData("sms_history_phoneNumber_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(smsHistory.PhoneNumber),
	})
	stringBuilder.WriteString(phoneNumberText)
	stringBuilder.WriteString(newLine)
	if smsHistory.SMSCode != nil {
		smsCode := *smsHistory.SMSCode
		smsCodeText := localizer.LocalizedStringWithTemplateData("sms_history_sms_code_markdown", map[string]any{
			"SMSCode": utils.EscapeMarkdownText(smsCode),
		})
		stringBuilder.WriteString(smsCodeText)
		stringBuilder.WriteString(newLine)
	}
	if smsHistory.ReceivedAt != nil {
		receivedAt := *smsHistory.ReceivedAt
		receivedAtText := utils.EscapeMarkdownText(receivedAt.Format(utils.FullDateFormat))
		receivedAtLocalizedText := localizer.LocalizedStringWithTemplateData("sms_history_received_at_markdown", map[string]any{
			"ReceivedDate": receivedAtText,
		})
		stringBuilder.WriteString(receivedAtLocalizedText)
		stringBuilder.WriteString(newLine)
	} else {
		startAt := *smsHistory.CreatedAt
		startAtText := utils.EscapeMarkdownText(startAt.Format(utils.FullDateFormat))
		receivedAtLocalizedText := localizer.LocalizedStringWithTemplateData("sms_history_start_at_markdown", map[string]any{
			"StartDate": startAtText,
		})
		stringBuilder.WriteString(receivedAtLocalizedText)
		stringBuilder.WriteString(newLine)
	}
	statusText := localizer.LocalizedStringWithTemplateData("sms_history_status_markdown", map[string]any{
		"State": f.Status(app.SMSActivationState(smsHistory.Status)),
	})
	stringBuilder.WriteString(statusText)
	return strings.TrimSpace(stringBuilder.String())
}

func (f *formatter) Status(state app.SMSActivationState) string {
	switch state {
	case app.CancelSMSActivateState:
		return "Cancel"
	case app.PendingSMSActivateState:
		return "Pending"
	case app.DoneSMSActivateState:
		return "Done"
	}
	return "Unknown"
}
