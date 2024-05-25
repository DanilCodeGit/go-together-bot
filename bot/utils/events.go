package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"ride-together-bot/db"
	"strconv"
	"strings"
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

func (e *Event) EventHistory(update tgbotapi.Update) (string, error) {
	chatID := update.Message.Chat.ID
	ids, err := e.db.GetEventsID(chatID, update)
	if err != nil {
		return "", errors.WithMessage(err, "GetEventsID")
	}
	uniqueIds := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		uniqueIds[int(id)] = struct{}{}
	}
	strIds := make([]string, len(uniqueIds))
	idx := 0
	for id := range uniqueIds {
		strIds[idx] = strconv.Itoa(id)
		idx++
	}
	eventIDsParam := strings.Join(strIds, ",")
	url := fmt.Sprintf("https://cr50181-wordpress-j3047.tw1.ru/events_history.php?chat_id=%v&id_events=%s", chatID, eventIDsParam)
	return url, nil
}
