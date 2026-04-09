package main

import (
	"context"
	"log"
	"os"

	tgClient "telegrambot/clients/telegram"
	event_consumer "telegrambot/consumer/event-consumer"
	"telegrambot/events/telegram"
	"telegrambot/storage/sqlite"

	"github.com/joho/godotenv"
)

const (
	tgBotHost         = "api.telegram.org"
	sqliteStoragePath = "data/sqlite/storage.db"
	batchSize         = 100
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	token := os.Getenv("TG_BOT_TOKEN")
	if token == "" {
		log.Fatal("TG_BOT_TOKEN is not set in .env or enviroment")
	}
	//s := files.New(storagePath)
	s, err := sqlite.New(sqliteStoragePath)
	if err != nil {
		log.Fatal("can't connect to storage: %w", err)
	}

	if err = s.Init(context.TODO()); err != nil {
		log.Fatal("can't init storage: %w", err)
	}
	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, token),
		s,
	)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)
	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped. ", err)
	}
}
