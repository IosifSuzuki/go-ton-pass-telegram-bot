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
	ConfirmationPay(langCode string, service *sms.Service, country *sms.Country, amount float64, preferredCurrency app.Currency) string
	StartSMSActivation(langCode string, smsHistory *domain.SMSHistory) string
	CompleteSMSActivation(langCode string, smsHistory *domain.SMSHistory) string
	FailSMSActivation(langCode string, smsHistory *domain.SMSHistory) string
	ManualCancelActivation(langCode string, smsHistory *domain.SMSHistory) string
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
	return f.representableCountry(country.Title, country.ID)
}

func (f *formatter) Service(service *sms.Service, _ FormatterType) string {
	return f.representableService(service.Name, service.Code)
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
	phoneNumber := app.PhoneNumber{
		CountryCode:      smsHistory.PhoneCodeNumber,
		ShortPhoneNumber: smsHistory.PhoneShortNumber,
	}
	localPhoneNumberRow := localizer.LocalizedStringWithTemplateData("sms_activation_short_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.ShortPhoneNumber),
	})
	internationPhoneNumberRow := localizer.LocalizedStringWithTemplateData("sms_activation_full_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.FullNumber()),
	})
	serviceRow := localizer.LocalizedStringWithTemplateData("sms_activation_service_markdown", map[string]any{
		"Service": f.representableService(smsHistory.ServiceName, smsHistory.ServiceCode),
	})
	countryRow := localizer.LocalizedStringWithTemplateData("sms_activation_country_markdown", map[string]any{
		"Country": f.representableCountry(smsHistory.CountryName, smsHistory.CountryID),
	})

	stringBuilder.WriteString(localPhoneNumberRow)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(internationPhoneNumberRow)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(serviceRow)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(countryRow)
	stringBuilder.WriteString(newLine)
	if smsHistory.SMSCode != nil {
		smsCode := *smsHistory.SMSCode
		smsCodeText := localizer.LocalizedStringWithTemplateData("sms_activation_code_markdown", map[string]any{
			"SMSCode": utils.EscapeMarkdownText(smsCode),
		})
		stringBuilder.WriteString(smsCodeText)
		stringBuilder.WriteString(newLine)
	}
	if smsHistory.ReceivedAt != nil {
		receivedAt := *smsHistory.ReceivedAt
		receivedAtText := utils.EscapeMarkdownText(receivedAt.Format(utils.FullDateFormat))
		receivedAtLocalizedText := localizer.LocalizedStringWithTemplateData("sms_activation_received_at_markdown", map[string]any{
			"ReceivedDate": receivedAtText,
		})
		stringBuilder.WriteString(receivedAtLocalizedText)
		stringBuilder.WriteString(newLine)
	} else {
		startAt := *smsHistory.CreatedAt
		startAtText := utils.EscapeMarkdownText(startAt.Format(utils.FullDateFormat))
		receivedAtLocalizedText := localizer.LocalizedStringWithTemplateData("sms_activation_start_at_markdown", map[string]any{
			"StartDate": startAtText,
		})
		stringBuilder.WriteString(receivedAtLocalizedText)
		stringBuilder.WriteString(newLine)
	}
	statusText := localizer.LocalizedStringWithTemplateData("sms_activation_status_markdown", map[string]any{
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

func (f *formatter) ConfirmationPay(langCode string, service *sms.Service, country *sms.Country, amount float64, preferredCurrency app.Currency) string {
	localizer := f.container.GetLocalizer(langCode)
	newLine := "\n"
	stringBuilder := strings.Builder{}
	title := localizer.LocalizedString("confirm_sms_activation_title_markdown")
	selectedService := localizer.LocalizedStringWithTemplateData("confirm_sms_activation_selected_service_markdown", map[string]any{
		"ServiceName": utils.EscapeMarkdownText(f.Service(service, DefaultFormatterType)),
	})
	priceForService := localizer.LocalizedStringWithTemplateData("confirm_sms_activation_price_for_service_markdown", map[string]any{
		"Price": utils.EscapeMarkdownText(utils.CurrencyAmountTextFormat(amount, preferredCurrency)),
	})
	selectedCountry := localizer.LocalizedStringWithTemplateData("confirm_sms_activation_selected_country_markdown", map[string]any{
		"Country": utils.EscapeMarkdownText(f.Country(country, DefaultFormatterType)),
	})
	footer := localizer.LocalizedString("confirm_sms_activation_footer_markdown")
	stringBuilder.WriteString(title)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedService)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(priceForService)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedCountry)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(footer)
	return stringBuilder.String()
}

func (f *formatter) StartSMSActivation(langCode string, smsHistory *domain.SMSHistory) string {
	localizer := f.container.GetLocalizer(langCode)
	newLine := "\n"
	stringBuilder := strings.Builder{}
	phoneNumber := app.PhoneNumber{
		CountryCode:      smsHistory.PhoneCodeNumber,
		ShortPhoneNumber: smsHistory.PhoneShortNumber,
	}
	title := localizer.LocalizedString("start_sms_activation_title_markdown")
	localPhoneNumber := localizer.LocalizedStringWithTemplateData("sms_activation_short_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.ShortPhoneNumber),
	})
	internationPhoneNumber := localizer.LocalizedStringWithTemplateData("sms_activation_full_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.FullNumber()),
	})
	selectedService := localizer.LocalizedStringWithTemplateData("sms_activation_service_markdown", map[string]any{
		"Service": f.representableService(smsHistory.ServiceName, smsHistory.ServiceCode),
	})
	selectedCountry := localizer.LocalizedStringWithTemplateData("sms_activation_country_markdown", map[string]any{
		"Country": f.representableCountry(smsHistory.CountryName, smsHistory.CountryID),
	})
	stringBuilder.WriteString(title)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(localPhoneNumber)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(internationPhoneNumber)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedService)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedCountry)
	return stringBuilder.String()
}

func (f *formatter) CompleteSMSActivation(langCode string, smsHistory *domain.SMSHistory) string {
	localizer := f.container.GetLocalizer(langCode)
	newLine := "\n"
	stringBuilder := strings.Builder{}
	phoneNumber := app.PhoneNumber{
		CountryCode:      smsHistory.PhoneCodeNumber,
		ShortPhoneNumber: smsHistory.PhoneShortNumber,
	}
	title := localizer.LocalizedString("success_received_sms_code_markdown")
	localPhoneNumber := localizer.LocalizedStringWithTemplateData("sms_activation_short_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.ShortPhoneNumber),
	})
	internationPhoneNumber := localizer.LocalizedStringWithTemplateData("sms_activation_full_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.FullNumber()),
	})
	selectedService := localizer.LocalizedStringWithTemplateData("sms_activation_service_markdown", map[string]any{
		"Service": f.representableService(smsHistory.ServiceName, smsHistory.ServiceCode),
	})
	selectedCountry := localizer.LocalizedStringWithTemplateData("sms_activation_country_markdown", map[string]any{
		"Country": f.representableCountry(smsHistory.CountryName, smsHistory.CountryID),
	})
	smsCode := localizer.LocalizedStringWithTemplateData("sms_activation_code_markdown", map[string]any{
		"SMSCode": utils.EscapeMarkdownText(*smsHistory.SMSCode),
	})
	footer := localizer.LocalizedString("success_received_sms_code_footer_markdown")
	stringBuilder.WriteString(title)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(localPhoneNumber)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(internationPhoneNumber)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedService)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedCountry)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(smsCode)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(footer)
	return stringBuilder.String()
}

func (f *formatter) FailSMSActivation(langCode string, smsHistory *domain.SMSHistory) string {
	localizer := f.container.GetLocalizer(langCode)
	newLine := "\n"
	stringBuilder := strings.Builder{}
	phoneNumber := app.PhoneNumber{
		CountryCode:      smsHistory.PhoneCodeNumber,
		ShortPhoneNumber: smsHistory.PhoneShortNumber,
	}
	title := localizer.LocalizedString("not_receive_sms_code_title_markdown")
	footer := localizer.LocalizedString("not_receive_sms_code_footer_markdown")
	localPhoneNumber := localizer.LocalizedStringWithTemplateData("sms_activation_short_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.ShortPhoneNumber),
	})
	internationPhoneNumber := localizer.LocalizedStringWithTemplateData("sms_activation_full_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.FullNumber()),
	})
	selectedService := localizer.LocalizedStringWithTemplateData("sms_activation_service_markdown", map[string]any{
		"Service": f.representableService(smsHistory.ServiceName, smsHistory.ServiceCode),
	})
	selectedCountry := localizer.LocalizedStringWithTemplateData("sms_activation_country_markdown", map[string]any{
		"Country": f.representableCountry(smsHistory.CountryName, smsHistory.CountryID),
	})
	startAt := localizer.LocalizedStringWithTemplateData("sms_activation_start_at_markdown", map[string]any{
		"StartDate": utils.EscapeMarkdownText(smsHistory.CreatedAt.Format(utils.FullDateFormat)),
	})
	stringBuilder.WriteString(title)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(localPhoneNumber)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(internationPhoneNumber)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedService)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedCountry)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(startAt)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(footer)
	return stringBuilder.String()
}

func (f *formatter) ManualCancelActivation(langCode string, smsHistory *domain.SMSHistory) string {
	localizer := f.container.GetLocalizer(langCode)
	newLine := "\n"
	stringBuilder := strings.Builder{}
	phoneNumber := app.PhoneNumber{
		CountryCode:      smsHistory.PhoneCodeNumber,
		ShortPhoneNumber: smsHistory.PhoneShortNumber,
	}
	title := localizer.LocalizedString("cancel_sms_activation_title_markdown")
	footer := localizer.LocalizedString("cancel_sms_activation_footer_markdown")
	localPhoneNumber := localizer.LocalizedStringWithTemplateData("sms_activation_short_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.ShortPhoneNumber),
	})
	internationPhoneNumber := localizer.LocalizedStringWithTemplateData("sms_activation_full_phone_number_markdown", map[string]any{
		"PhoneNumber": utils.EscapeMarkdownText(phoneNumber.FullNumber()),
	})
	selectedService := localizer.LocalizedStringWithTemplateData("sms_activation_service_markdown", map[string]any{
		"Service": f.representableService(smsHistory.ServiceName, smsHistory.ServiceCode),
	})
	selectedCountry := localizer.LocalizedStringWithTemplateData("sms_activation_country_markdown", map[string]any{
		"Country": f.representableCountry(smsHistory.CountryName, smsHistory.CountryID),
	})
	stringBuilder.WriteString(title)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(localPhoneNumber)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(internationPhoneNumber)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedService)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(selectedCountry)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(newLine)
	stringBuilder.WriteString(footer)
	return stringBuilder.String()
}

func (f *formatter) representableCountry(countryName string, countryID int64) string {
	var title string
	name := f.container.GetRepresentableCountryName(countryID)
	if name == nil {
		name = &countryName
	}
	flag := f.container.GetFlagEmoji(*name)
	if flag != nil {
		title = fmt.Sprintf("%s %s", *flag, *name)
	} else {
		title = countryName
	}
	return title
}

func (f *formatter) representableService(serviceName string, serviceCode string) string {
	var (
		name  string
		emoji string
	)
	extraService := f.container.GetExtraService(serviceCode)
	if extraService != nil {
		name = extraService.Name
		emoji = extraService.Emoji
	} else {
		name = serviceName
		emoji = "🌐"
	}
	return fmt.Sprintf("%s %s", emoji, name)
}
