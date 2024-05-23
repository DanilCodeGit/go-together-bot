package main

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"ride-together-bot/bot"
	"ride-together-bot/conf"
	"ride-together-bot/db"
)

func main() {
	// Create a background context
	ctx := context.Background()

	// Initialize the database connection
	conn, err := db.NewDataBase(conf.DSN)
	if err != nil {
		log.Panic(errors.WithMessage(err, "ошибка инициализации БД"))
	}

	// Initialize the bot API instance
	instance, err := tgbotapi.NewBotAPI(conf.TelegramBotApiKey)
	if err != nil {
		log.Panic(err)
	}

	// Create an instance of BotApi
	botInstance := bot.NewBot(instance, conn)

	// Start processing updates
	if err := botInstance.Updates(ctx); err != nil {
		log.Panic(err)
	}
}
