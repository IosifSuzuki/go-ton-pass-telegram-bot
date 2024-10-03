package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"io"
	"net/http"
	"strings"
)

type TelegramBotService interface {
	ParseTelegramCommand(update *telegram.Update) (app.TelegramCommand, error)
	ParseTelegramCallbackData(callbackQuery *telegram.CallbackQuery) (*app.TelegramCallbackData, error)
	GetLanguagesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup
	GetCurrenciesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup
	SendResponse(model any, method app.TelegramMethod) error

	GetSetMyCommands() *telegram.SetMyCommands
	GetSetMyDescription() *telegram.SetMyDescription
	GetSetMyName() *telegram.SetMyName
}

type telegramBotService struct {
	container container.Container
}

const (
	startCmdText = "/start"
	helpCmdText  = "/help"
)

func NewTelegramBot(container container.Container) TelegramBotService {
	return &telegramBotService{
		container: container,
	}
}

func (t *telegramBotService) ParseTelegramCallbackData(callbackQuery *telegram.CallbackQuery) (*app.TelegramCallbackData, error) {
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		return nil, err
	}
	return telegramCallbackData, nil
}

func (t *telegramBotService) ParseTelegramCommand(update *telegram.Update) (app.TelegramCommand, error) {
	var text = ""
	if update.Message != nil {
		text = *update.Message.Text
	}
	return parseTelegramCommand(text)
}

func (t *telegramBotService) GetLanguagesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup {
	log := t.container.GetLogger()
	languages := t.container.GetConfig().AvailableLanguages()

	log.Debug("AvailableLanguages from configuration", logger.F("AvailableLanguages", languages))

	keyboardButtons := make([][]telegram.KeyboardButton, 0, len(languages))
	for _, language := range languages {
		buttonText := fmt.Sprintf("%s %s", language.FlagEmoji, language.NativeName)
		keyboardButtons = append(keyboardButtons, []telegram.KeyboardButton{
			{
				Text: buttonText,
			},
		})
	}
	return &telegram.ReplyKeyboardMarkup{
		Keyboard:                  keyboardButtons,
		PersistentDisplayKeyboard: false,
		ResizeKeyboard:            true,
		OneTimeKeyboard:           true,
		Placeholder:               nil,
	}
}

func (t *telegramBotService) GetSetMyCommands() *telegram.SetMyCommands {
	startBotCommand := telegram.BotCommand{
		Command:     "start",
		Description: t.container.GetLocalizer("en").LocalizedString("start_cmd_short_description_cmd"),
	}
	helpBotCommand := telegram.BotCommand{
		Command:     "help",
		Description: t.container.GetLocalizer("en").LocalizedString("help_cmd_short_description_cmd"),
	}
	return &telegram.SetMyCommands{Commands: []telegram.BotCommand{
		startBotCommand,
		helpBotCommand,
	}}
}

func (t *telegramBotService) GetSetMyDescription() *telegram.SetMyDescription {
	return &telegram.SetMyDescription{
		Description: t.container.GetLocalizer("en").LocalizedString("bot_description"),
	}
}

func (t *telegramBotService) GetSetMyName() *telegram.SetMyName {
	return &telegram.SetMyName{
		Name: t.container.GetLocalizer("en").LocalizedString("bot_name"),
	}
}

func (t *telegramBotService) GetCurrenciesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup {
	log := t.container.GetLogger()
	currencies := t.container.GetConfig().AvailablePreferredCurrencies()

	log.Debug("AvailablePreferredCurrencies from configuration", logger.F("AvailablePreferredCurrencies", currencies))
	keyboardButtons := make([][]telegram.KeyboardButton, 0, len(currencies))
	for _, currency := range currencies {
		buttonText := fmt.Sprintf("%s %s", currency.Symbol, currency.ABBR)
		keyboardButtons = append(keyboardButtons, []telegram.KeyboardButton{
			{
				Text: buttonText,
			},
		})
	}
	return &telegram.ReplyKeyboardMarkup{
		Keyboard:                  keyboardButtons,
		PersistentDisplayKeyboard: false,
		ResizeKeyboard:            true,
		OneTimeKeyboard:           true,
		Placeholder:               nil,
	}
}

func (t *telegramBotService) SendResponse(model any, method app.TelegramMethod) error {
	log := t.container.GetLogger()
	telegramBotToken := t.container.GetConfig().TelegramBotToken()
	const baseTelegramAPI = "https://api.telegram.org/bot"
	path := fmt.Sprintf("%s%s/%s", baseTelegramAPI, telegramBotToken, method)
	sendBody, err := json.Marshal(model)
	if err != nil {
		log.Error("fail to encode telegram message", logger.FError(err))
		return err
	}
	log.Debug("send response to telegram servers",
		logger.F("message", string(sendBody)),
		logger.F("path", path),
	)
	bodyBuffer := bytes.NewBuffer(sendBody)
	resp, err := http.Post(path, "application/json", bodyBuffer)
	if err != nil {
		log.Error("fail to send data to telegram server", logger.FError(err))
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("fail to read body from telegram server", logger.FError(err))
		return err
	}
	var result *telegram.Result
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error("fail to decode body from telegram server", logger.FError(err))
		return err
	}
	if !result.OK {
		log.Debug("telegram server return without status code ok",
			logger.F("description", result.Description),
			logger.F("json", string(sendBody)),
		)
		return nil
	}
	return nil
}

func parseTelegramCommand(text string) (app.TelegramCommand, error) {
	switch text {
	case startCmdText:
		return app.StartTelegramCommand, nil
	case helpCmdText:
		return app.HelpTelegramCommand, nil
	default:
		break
	}
	if strings.HasPrefix(text, "/") {
		return app.UnknownTelegramCommand, app.NotSupportedTelegramCommandError
	}
	return app.NotTelegramCommand, app.NotTelegramCommandError
}
