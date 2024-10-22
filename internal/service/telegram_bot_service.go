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
	SendResponse(model any, method app.TelegramMethod) error
	UserIsChatMember(chatID string, telegramID int64) (bool, error)
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

const baseTelegramAPI = "https://api.telegram.org/bot"

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

func (t *telegramBotService) prepareRequest(method app.TelegramMethod, model any) (*http.Request, error) {
	config := t.container.GetConfig()
	log := t.container.GetLogger()
	telegramBotToken := config.TelegramBotToken()
	path := fmt.Sprintf("%s%s/%s", baseTelegramAPI, telegramBotToken, method)
	jsonData, err := json.Marshal(model)
	if err != nil {
		log.Error("fail to marshal json model", logger.FError(err))
		return nil, err
	}
	log.Debug("will send response to telegram servers",
		logger.F("message", string(jsonData)),
		logger.F("path", path),
	)
	bodyBuffer := bytes.NewReader(jsonData)
	req, err := http.NewRequest(http.MethodPost, path, bodyBuffer)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Error("fail to create request", logger.FError(err))
		return nil, err
	}
	return req, nil
}

func (t *telegramBotService) SendResponse(model any, method app.TelegramMethod) error {
	log := t.container.GetLogger()
	req, err := t.prepareRequest(method, model)
	if err != nil {
		log.Error("fail to prepare request", logger.FError(err))
		return err
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Error("fail to perform request", logger.FError(err))
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("fail to read body from telegram server", logger.FError(err))
		return err
	}
	var result *telegram.Result[any]
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error("fail to decode body from telegram server", logger.FError(err))
		return err
	}
	if !result.OK {
		log.Debug("telegram server return without status code ok",
			logger.F("description", result.Description),
			logger.F("json", string(body)),
		)
		return nil
	}
	return nil
}

func (t *telegramBotService) UserIsChatMember(chatID string, telegramID int64) (bool, error) {
	log := t.container.GetLogger()
	getChatMember := telegram.GetChatMember{
		ChatID: chatID,
		UserID: telegramID,
	}
	req, err := t.prepareRequest(app.GetChatMemberTelegramMethod, getChatMember)
	if err != nil {
		log.Error("fail to create request", logger.FError(err))
		return false, err
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Error("fail to get perform response", logger.FError(err))
		return false, err
	}
	var result telegram.Result[telegram.ChatMember]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error("fail to decode ChatMember", logger.FError(err))
		return false, err
	}
	log.Debug("get result from telegram for check is user a chat member", logger.F("telegram_id", telegramID), logger.F("status", result.Result.Status))
	switch result.Result.Status {
	case telegram.MemberMemberStatus, telegram.CreatorMemberStatus, telegram.AdministratorMemberStatus:
		return true, nil
	default:
		return false, nil
	}
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
