package main

import (
	"go-ton-pass-telegram-bot/internal/config"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/router"
	"log"
	"net/http"
	"time"
)

func main() {
	conf := config.ParseConfig()
	box := container.NewContainer(conf)
	RunServer(box)
}

func RunServer(box container.Container) {
	r := router.PrepareAndConfigureRouter(box)
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
