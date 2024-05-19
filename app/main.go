package main

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"ride-together-bot/bot"
	botUtils "ride-together-bot/bot/utils"
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

	// Initialize additional components needed for the bot
	sticker := botUtils.NewSticker(instance)
	maps := botUtils.NewMaps(instance)
	contact := botUtils.NewContact(instance, conn, sticker)
	location := botUtils.NewLocation(instance, conn)

	// Create an instance of BotApi
	botInstance := bot.NewBot(instance, conn, sticker, maps, contact, location)

	// Start processing updates
	if err := botInstance.Updates(ctx); err != nil {
		log.Panic(err)
	}
}
