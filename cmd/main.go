package main

import (
	"context"
	"encoding/json"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/redis/go-redis/v9"
	"go-ton-pass-telegram-bot/internal/config"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/router"
	"go-ton-pass-telegram-bot/internal/service"
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
	redisClient := configureAndConnectToRedisClient(conf)
	sessionService := service.NewSessionService(box, redisClient)
	updateTelegramBotProfile(box)
	RunServer(box, sessionService)
}

func configureAndConnectToRedisClient(conf config.Config) *redis.Client {
	redisConfig := conf.Redis()
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Address(),
		Password: redisConfig.Password,
		DB:       redisConfig.DataBase,
	})
	status, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf(
			"ping to redis failed with status: %s; error: %v; connection addr: %s",
			status,
			err,
			redisConfig.Address(),
		)
		return nil
	}
	return client
}

func loadBundle() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustLoadMessageFile("locales/en.json")
	bundle.MustLoadMessageFile("locales/sk.json")
	bundle.MustLoadMessageFile("locales/uk.json")
	return bundle
}

func RunServer(box container.Container, sessionService service.SessionService) {
	r := router.PrepareAndConfigureRouter(box, sessionService)
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

func updateTelegramBotProfile(box container.Container) {
	telegramService := service.NewTelegramBot(box)

	go setBotCommands(telegramService)
	go setBotDescription(telegramService)
	go setBotName(telegramService)
}

func setBotCommands(telegramService service.TelegramBotService) {
	model := telegramService.GetSetMyCommands()
	if err := telegramService.SendResponse(model, app.SetMyCommandsTelegramMethod); err != nil {
		log.Println("setBotCommands: ", err)
	}
}

func setBotDescription(telegramService service.TelegramBotService) {
	model := telegramService.GetSetMyDescription()
	if err := telegramService.SendResponse(model, app.SetMyDescriptionTelegramMethod); err != nil {
		log.Println("setBotDescription: ", err)
	}
}

func setBotName(telegramService service.TelegramBotService) {
	model := telegramService.GetSetMyName()
	if err := telegramService.SendResponse(model, app.SetMyNameTelegramMethod); err != nil {
		log.Println("setMyName: ", err)
	}
}
