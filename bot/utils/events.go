package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"ride-together-bot/conf/stickers"
	"ride-together-bot/db"
	"strconv"
	"strings"
)

type Event struct {
	api *tgbotapi.BotAPI
	db  *db.DB
	s   Sticker
}

func NewEvent(api *tgbotapi.BotAPI, db *db.DB, s Sticker) *Event {
	return &Event{
		api: api,
		db:  db,
		s:   s,
	}
}

func (e *Event) CreateEvent(chatID int64, update tgbotapi.Update) {
	// Формируем URL с параметром запроса с именем пользователя
	url := fmt.Sprintf("https://cr50181-wordpress-j3047.tw1.ru/create_event_page.php?chatID=%d", chatID)
	webappInfo := tgbotapi.WebAppInfo{URL: url}
	btn := tgbotapi.NewKeyboardButtonWebApp("Создать поездку", webappInfo)
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(btn),
	)
	msg := tgbotapi.NewMessage(chatID, "ㅤ")
	msg.ReplyMarkup = keyboard
	e.api.Send(msg)
	e.s.SendSticker(stickers.CreateEvent, update)
}

func (e *Event) ActiveEvents(ctx context.Context, update tgbotapi.Update) error {
	chatID := update.Message.Chat.ID
	isDriver, err := e.db.IsDriver(ctx, update)
	if err != nil {
		return errors.WithMessage(err, "IsDriver")
	}
	url := fmt.Sprintf("https://cr50181-wordpress-j3047.tw1.ru/driver_events.php?chatID=%d&isDriver=%v", chatID, isDriver)
	webappInfo := tgbotapi.WebAppInfo{URL: url}
	btn := tgbotapi.NewKeyboardButtonWebApp("Активные поездки", webappInfo)
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(btn),
	)
	msg := tgbotapi.NewMessage(chatID, "ㅤ")
	msg.ReplyMarkup = keyboard
	e.api.Send(msg)
	e.s.SendSticker(stickers.Cat, update)
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
