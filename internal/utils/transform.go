package utils

import (
	"encoding/base64"
	"github.com/vmihailenco/msgpack/v5"
	"go-ton-pass-telegram-bot/internal/model/app"
)

func EncodeTelegramCallbackData(callbackData app.TelegramCallbackData) (*string, error) {
	data, err := msgpack.Marshal(callbackData)
	if err != nil {
		return nil, err
	}
	encodedText := base64.StdEncoding.EncodeToString(data)
	return &encodedText, nil
}

func DecodeTelegramCallbackData(text string) (*app.TelegramCallbackData, error) {
	decodedData, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	var telegramCallbackData app.TelegramCallbackData
	err = msgpack.Unmarshal(decodedData, &telegramCallbackData)
	return &telegramCallbackData, err
}
