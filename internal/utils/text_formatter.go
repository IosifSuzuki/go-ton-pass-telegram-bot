package utils

import (
	"fmt"
	"go-ton-pass-telegram-bot/internal/model/app"
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
	return fmt.Sprintf("%s %s", currency.Symbol, currency.ABBR)
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
