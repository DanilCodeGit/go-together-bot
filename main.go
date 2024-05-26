package main

import (
	"context"
	"embed"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"ride-together-bot/bot"
	"ride-together-bot/conf"
	"ride-together-bot/db"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	// Create a background context
	ctx := context.Background()

	// Initialize the database connection
	dataBase, err := db.NewDataBase(conf.DSN)
	if err != nil {
		log.Panic(errors.WithMessage(err, "ошибка инициализации БД"))
	}

	// Run migrations
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(dataBase.Conn, "migrations"); err != nil {
		log.Fatal("goose up: ", err)
	}
	// Initialize the bot API instance
	instance, err := tgbotapi.NewBotAPI(conf.TelegramBotApiKey)
	if err != nil {
		log.Panic(err)
	}

	// Create an instance of BotApi
	botInstance := bot.NewBot(instance, dataBase)

	// Start processing updates
	if err := botInstance.Updates(ctx); err != nil {
		log.Panic(err)
	}
}
