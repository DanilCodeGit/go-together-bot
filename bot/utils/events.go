package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"ride-together-bot/db"
)

type Event struct {
	api *tgbotapi.BotAPI
	db  *db.DB
}

func NewEvent(api *tgbotapi.BotAPI, db *db.DB) *Event {
	return &Event{
		api: api,
		db:  db,
	}
}

func (e *Event) ActiveEvents(ctx context.Context, update tgbotapi.Update) error {
	chatID := update.Message.Chat.ID
	ok, err := e.db.IsDriver(ctx, update)
	if err != nil {
		return errors.WithMessage(err, "IsDriver")
	}

	url := fmt.Sprintf("https://cr50181-wordpress-j3047.tw1.ru/driver_events.php?isDriver=%t&chatID=%d", ok, chatID)
	webappInfo := tgbotapi.WebAppInfo{URL: url}
	btn := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonWebApp("Активные поездки", webappInfo),
	)

	reply := tgbotapi.NewInlineKeyboardMarkup(btn)
	msg := tgbotapi.NewMessage(chatID, "Активные поездки")
	msg.ReplyMarkup = reply
	e.api.Send(msg)
	return nil
}
