package main

import (
	"context"
	"embed"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	tele "gopkg.in/telebot.v3"
	"log"
	"ride-together-bot/bot"
	"ride-together-bot/conf"
	"ride-together-bot/db"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	ctx := context.Background()
	cfg := conf.NewConfig()

	dataBase, err := db.NewDataBase(cfg.DSN.DSNLocal)
	if err != nil {
		log.Println(cfg.DSN.DSNLocal)
		log.Panic(errors.WithMessage(err, "ошибка инициализации БД"))
	}

	goose.SetBaseFS(embedMigrations)

	err = goose.SetDialect("mysql")
	if err != nil {
		log.Fatal(err)
	}

	err = goose.Up(dataBase.Conn, "migrations")
	if err != nil {
		log.Fatal("goose up: ", err)
	}

	pref := tele.Settings{
		Token: cfg.TelegramBotApiKey,
		Poller: &tele.LongPoller{
			LastUpdateID: 0,
		},
		Verbose: true,
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "create telegram bot"))
		return
	}

	botInstance := bot.NewBot(cfg, b, dataBase)

	// Start processing updates
	botInstance.Start(ctx)
}
