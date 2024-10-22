package test

import (
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/utils"
	"testing"
)

func TestCurrencyTextFormatted(t *testing.T) {
	t.Run("well formatted currency", func(t *testing.T) {
		usdCurrency := app.Currency{
			Name:   "United States Dollar",
			ABBR:   "USD",
			Symbol: "$",
		}
		euroCurrency := app.Currency{
			Name:   "Euro",
			ABBR:   "EUR",
			Symbol: "€",
		}
		usdCurrencyTextFormat := utils.CurrencyAmountTextFormat(2, usdCurrency)
		if usdCurrencyTextFormat != "$2.00" {
			t.Errorf("unexpected text format: %v", usdCurrencyTextFormat)
		}
		euroTextFormat := utils.CurrencyAmountTextFormat(2, euroCurrency)
		if euroTextFormat != "2.00€" {
			t.Errorf("unexpected text format: %v", euroTextFormat)
		}
	})
}
