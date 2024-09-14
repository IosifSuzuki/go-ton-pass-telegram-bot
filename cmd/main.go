package main

import (
	"encoding/json"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-ton-pass-telegram-bot/internal/config"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/router"
	"go-ton-pass-telegram-bot/pkg/logger"
	"golang.org/x/text/language"
	"log"
	"net/http"
	"time"
)

func main() {
	conf, err := config.ParseConfig()
	if err != nil {
		log.Fatalln(err)
	}
	bundle := loadBundle()
	l := logger.NewLogger(logger.DEV, logger.LevelDebug)
	box := container.NewContainer(l, conf, bundle)
	RunServer(box)
}

func loadBundle() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustLoadMessageFile("locales/en.json")
	bundle.MustLoadMessageFile("locales/sk.json")
	bundle.MustLoadMessageFile("locales/uk.json")
	return bundle
}

func RunServer(box container.Container) {
	r := router.PrepareAndConfigureRouter(box)
	server := &http.Server{
		Handler:      r,
		Addr:         box.GetConfig().Address(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
