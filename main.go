// package main
//
// import (
//
//	"context"
//	"embed"
//	"github.com/pkg/errors"
//	"log"
//	"ride-together-bot/bot"
//	"ride-together-bot/conf"
//	"ride-together-bot/db"
//	"ride-together-bot/tgbotapi"
//
//	"github.com/pressly/goose/v3"
//
// )
//
// //go:embed migrations/*.sql
// var embedMigrations embed.FS
//
//	func main() {
//		// Create a background context
//		ctx := context.Background()
//
//		// Initialize the database connection
//		dataBase, err := db.NewDataBase(conf.DSN)
//		if err != nil {
//			log.Println(conf.DSN)
//			log.Panic(errors.WithMessage(err, "ошибка инициализации БД"))
//		}
//
//		// Run migrations
//		goose.SetBaseFS(embedMigrations)
//
//		if err := goose.SetDialect("mysql"); err != nil {
//			log.Fatal(err)
//		}
//
//		if err := goose.Up(dataBase.Conn, "migrations"); err != nil {
//			log.Fatal("goose up: ", err)
//		}
//		// Initialize the bot API instance
//		instance, err := tgbotapi.NewBotAPI(conf.TelegramBotApiKey)
//		if err != nil {
//			log.Panic(err)
//		}
//
//		// Удаление текущего вебхука
//		_, err = instance.Request(tgbotapi.DeleteWebhookConfig{})
//		if err != nil {
//			log.Panic(errors.WithMessage(err, "ошибка удаления вебхука"))
//		}
//
//		// Create an instance of BotApi
//		botInstance := bot.NewBot(instance, dataBase)
//
//		// Start processing updates
//		if err := botInstance.Updates(ctx); err != nil {
//			log.Panic(err)
//		}
//	}
package main

import (
	"context"
	"embed"
	"github.com/pkg/errors"
	tele "gopkg.in/telebot.v3"
	"log"
	"ride-together-bot/bot"
	"ride-together-bot/conf"
	"ride-together-bot/db"
	"time"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	ctx := context.Background()

	dataBase, err := db.NewDataBase(conf.DSN)
	if err != nil {
		log.Println(conf.DSN)
		log.Panic(errors.WithMessage(err, "ошибка инициализации БД"))
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(dataBase.Conn, "migrations"); err != nil {
		log.Fatal("goose up: ", err)
	}

	pref := tele.Settings{
		Token:  conf.TelegramBotApiKey,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Create an instance of BotApi
	botInstance := bot.NewBot(b, dataBase)

	// Start processing updates
	botInstance.Start(ctx)
}
