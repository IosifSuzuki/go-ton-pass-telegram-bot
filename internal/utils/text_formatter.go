package utils

import (
	"fmt"
	"go-ton-pass-telegram-bot/internal/model/app"
	"regexp"
	"strings"
)

func CurrencyAmountTextFormat(amount float64, currency app.Currency) string {
	var builder strings.Builder
	amountText := fmt.Sprintf("%.2f", amount)
	if strings.EqualFold(currency.ABBR, "USD") {
		builder.WriteString(currency.Symbol)
		builder.WriteString(amountText)
	} else {
		builder.WriteString(amountText)
		builder.WriteString(currency.Symbol)
	}
	return builder.String()
}

func ShortCurrencyTextFormat(currency app.Currency) string {
	return fmt.Sprintf("%s | %s | %s", currency.Symbol, currency.ABBR, currency.Emoji)
}

func LanguageTextFormat(language app.Language) string {
	return fmt.Sprintf("%s %s", language.FlagEmoji, language.NativeName)
}

func ButtonTitle(title string, emojiIcon string) string {
	return fmt.Sprintf("%s %s", emojiIcon, title)
}

func Equal(lhs string, rhs string) bool {
	return strings.EqualFold(lhs, rhs)
}

func PhoneNumberTitle(number string) string {
	return fmt.Sprintf("+%s", number)
}

func ParsePhoneNumber(phoneNumber string) (*app.PhoneNumber, error) {
	re := regexp.MustCompile(`^\+(\d+)(\d{10,})$`)
	matches := re.FindStringSubmatch(phoneNumber)
	if len(matches) == 0 {
		return nil, app.UnknownPhoneNumberFormatError
	}
	return &app.PhoneNumber{
		CountryCode:      matches[1],
		ShortPhoneNumber: matches[2],
	}, nil
}

func EscapeMarkdownText(text string) string {
	escapedText := text
	specialChars := []string{"\\", "*", "_", "{", "}", "[", "]", "(", ")", "#", "+", "-", ".", "!", "|"}
	for _, char := range specialChars {
		escapedText = strings.ReplaceAll(escapedText, char, "\\"+char)
	}
	return escapedText
}
