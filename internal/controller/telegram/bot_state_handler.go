package telegram

import (
	"context"
	"fmt"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
)

func (b *botController) userSelectedLanguageBotStageHandler(ctx context.Context, update *telegram.Update) error {
	telegramID := update.Message.From.ID
	if update.Message.Text == nil {
		return app.NilError
	}
	selectedLanguageNativeName := *update.Message.Text

	availableLanguages := b.container.GetConfig().AvailableLanguages()
	filteredLanguages := utils.Filter(availableLanguages, func(language app.Language) bool {
		presentableLanguageText := fmt.Sprintf("%s %s", language.FlagEmoji, language.NativeName)
		return presentableLanguageText == selectedLanguageNativeName
	})
	if len(filteredLanguages) == 0 {
		return app.UnknownLanguageError
	}
	selectedLanguage := filteredLanguages[0]

	if err := b.profileRepository.SetPreferredLanguage(ctx, telegramID, selectedLanguage.Code); err != nil {
		return err
	}
	if err := b.messageWelcome(ctx, update); err != nil {
		return err
	}
	return b.messageToSelectCurrency(ctx, update)
}

func (b *botController) userSelectedCurrencyBotStageHandler(ctx context.Context, update *telegram.Update) error {
	telegramID := update.Message.From.ID
	if update.Message.Text == nil {
		return app.NilError
	}
	selectedCurrencyText := *update.Message.Text

	availableCurrencies := b.container.GetConfig().AvailableCurrencies()
	filteredCurrencies := utils.Filter(availableCurrencies, func(currency app.Currency) bool {
		presentableCurrencyText := fmt.Sprintf("%s %s", currency.Symbol, currency.ABBR)
		return presentableCurrencyText == selectedCurrencyText
	})
	if len(filteredCurrencies) == 0 {
		return app.UnknownLanguageError
	}
	selectedCurrency := filteredCurrencies[0]
	if err := b.profileRepository.SetPreferredCurrency(ctx, telegramID, selectedCurrency.ABBR); err != nil {
		return err
	}
	if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
		return err
	}
	return b.messageMainMenu(ctx, update)
}
