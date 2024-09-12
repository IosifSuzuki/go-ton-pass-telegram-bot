package main

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-ton-pass-telegram-bot/internal/config"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/router"
	"go-ton-pass-telegram-bot/internal/service/telegramBot"
	"go-ton-pass-telegram-bot/pkg/logger"
	"golang.org/x/text/language"
	"log"
	"net/http"
	"time"
)

func main() {
	conf := config.ParseConfig()
	bundle := loadBundle()
	l := logger.NewLogger(logger.DEV, logger.LevelDebug)
	box := container.NewContainer(l, conf, bundle)
	RunServer(box)
}

func loadBundle() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.MustLoadMessageFile("locales/en.json")
	return bundle
}

func RunServer(box container.Container) {
	telegramBotService := telegramBot.NewTelegramBot(box)
	r := router.PrepareAndConfigureRouter(box, telegramBotService)
	server := &http.Server{
		Handler:      r,
		Addr:         box.GetServerAddress(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
