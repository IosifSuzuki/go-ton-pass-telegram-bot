package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/redis/go-redis/v9"
	"go-ton-pass-telegram-bot/internal/config"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/router"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/service/postpone"
	"go-ton-pass-telegram-bot/pkg/logger"
	"go.temporal.io/sdk/client"
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
	db, err := openConnectionToDB(conf.DB())
	defer db.Close()
	if err != nil {
		log.Fatalln(err)
	}
	bundle := loadBundle()
	l := logger.NewLogger(logger.DEV, logger.LevelDebug)
	box := container.NewContainer(l, conf, bundle)
	if err := box.PreloadData(); err != nil {
		log.Fatalln(err)
	}
	redisClient := configureAndConnectToRedisClient(conf)
	defer redisClient.Close()
	sessionService := service.NewSessionService(box, redisClient)
	cacheService := service.NewCache(box, redisClient)
	temporalClient, err := client.Dial(client.Options{
		HostPort:  conf.Temporal().Address(),
		Namespace: "default",
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer temporalClient.Close()
	//updateTelegramBotProfile(box)
	RunServer(box, db, sessionService, temporalClient, cacheService)
}

func configureAndConnectToRedisClient(conf config.Config) *redis.Client {
	redisConfig := conf.Redis()
	c := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Address(),
		Password: redisConfig.Password,
		DB:       redisConfig.DataBase,
	})
	status, err := c.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf(
			"ping to redis failed with status: %s; error: %v; connection addr: %s",
			status,
			err,
			redisConfig.Address(),
		)
		return nil
	}
	return c
}

func RunServer(box container.Container, conn *sql.DB, sessionService service.SessionService, temporalClient client.Client, cacheService service.Cache) {
	profileRepository := repository.NewProfileRepository(conn)
	smsHistoryRepository := repository.NewSMSHistoryRepository(conn)
	smsService := service.NewSMSService(box)
	postponeService := postpone.NewPostpone(box, temporalClient, profileRepository, smsHistoryRepository)
	if err := postponeService.Prepare(); err != nil {
		log.Fatalln("fail to prepare postpone service", logger.FError(err))
	}

	r := router.PrepareAndConfigureRouter(box, sessionService, cacheService, smsService, postponeService, profileRepository, smsHistoryRepository)
	openServer := &http.Server{
		Handler:      r,
		Addr:         box.GetConfig().OpenConnectionAddress(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	secureServer := &http.Server{
		Handler:      r,
		Addr:         box.GetConfig().SecureConnectionAddress(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := openServer.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()
	if err := secureServer.ListenAndServeTLS("tls/certificate.crt", "tls/private.key"); err != nil {
		log.Fatalln(err)
	}
}

func loadBundle() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustLoadMessageFile("locales/en.json")
	bundle.MustLoadMessageFile("locales/sk.json")
	bundle.MustLoadMessageFile("locales/uk.json")
	return bundle
}

func openConnectionToDB(db config.DB) (*sql.DB, error) {
	psqlConn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host,
		db.Port,
		db.User,
		db.Password,
		db.Name,
		db.Mode,
	)
	conn, err := sql.Open("postgres", psqlConn)
	if err != nil {
		return conn, err
	}
	err = conn.Ping()
	return conn, err
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
