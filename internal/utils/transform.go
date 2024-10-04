package utils

import (
	"encoding/base64"
	"encoding/json"
	"github.com/vmihailenco/msgpack/v5"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
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

func EncodeCryptoBotInvoicePayload(invoicePayload bot.InvoicePayload) (*string, error) {
	data, err := json.Marshal(invoicePayload)
	if err != nil {
		return nil, err
	}
	text := base64.StdEncoding.EncodeToString(data)
	return &text, nil
}

func DecodeCryptoBotInvoicePayload(text string) (*bot.InvoicePayload, error) {
	data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	var invoicePayload bot.InvoicePayload
	if err := json.Unmarshal(data, &invoicePayload); err != nil {
		return nil, err
	}
	return &invoicePayload, nil
}
