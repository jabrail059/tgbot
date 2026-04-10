package main

import (
	"context"
	"log"
	"os"

	tgClient "telegrambot/clients/telegram"
	event_consumer "telegrambot/consumer/event-consumer"
	"telegrambot/events/telegram"
	"telegrambot/storage/postgres"

	"github.com/joho/godotenv"
)

const (
	tgBotHost = "api.telegram.org"
	batchSize = 100
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	token := os.Getenv("TG_BOT_TOKEN")
	if token == "" {
		log.Fatal("TG_BOT_TOKEN is not set in .env or enviroment")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	s, err := postgres.New(dbURL)
	if err != nil {
		log.Fatal("can't connect to storage: ", err)
	}

	if err = s.Init(context.TODO()); err != nil {
		log.Fatal("can't init storage: ", err)
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
