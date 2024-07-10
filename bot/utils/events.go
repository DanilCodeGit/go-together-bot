package bot

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"
	"ride-together-bot/conf"
	"ride-together-bot/conf/stickers"
	"ride-together-bot/db"
	"ride-together-bot/entity"
	"strconv"
	"strings"
)

type Event struct {
	conf *conf.Config
	api  *telebot.Bot
	db   *db.DB
	s    Sticker
}

func NewEvent(conf *conf.Config, api *telebot.Bot, db *db.DB, s Sticker) *Event {
	return &Event{
		conf: conf,
		api:  api,
		db:   db,
		s:    s,
	}
}

func (e *Event) CreateEvent(chatID int64) error {
	// Формируем URL с параметром запроса с именем пользователя
	url := fmt.Sprintf("%s?chatID=%d", e.conf.URLs.CreateEventPage, chatID)

	// Создаем кнопку для открытия webapp
	btn := telebot.ReplyButton{
		Text:   "Создать поездку",
		WebApp: &telebot.WebApp{URL: url},
	}

	// Создаем клавиатуру с одной кнопкой
	keyboard := telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{{btn}},
	}

	// Отправляем сообщение с клавиатурой
	_, err := e.api.Send(telebot.ChatID(chatID), "ㅤ", &keyboard)
	if err != nil {
		_ = fmt.Errorf("Error sending message: %v\n", err)
	}

	// Отправляем стикер
	err = e.s.SendSticker(chatID, stickers.CreateEvent)
	if err != nil {
		return entity.ErrSendSticker
	}
	return nil
}

func (e *Event) TripsManagement(ctx context.Context, message *telebot.Message) error {
	chatID := message.Chat.ID
	isDriver, err := e.db.IsDriver(ctx, message)
	if err != nil {
		return errors.WithMessage(err, "IsDriver")
	}

	url := fmt.Sprintf("%s?chatID=%d&isDriver=%v", e.conf.URLs.DriverEvents, chatID, isDriver)
	webappInfo := &telebot.WebApp{URL: url}
	btn := telebot.Btn{Text: "Менеджер поездок", WebApp: webappInfo}
	keyboard := &telebot.ReplyMarkup{ResizeKeyboard: true}
	keyboard.Reply(keyboard.Row(btn))

	_, err = e.api.Send(telebot.ChatID(chatID), "Используйте кнопку ниже для управления поездками", &telebot.SendOptions{
		ReplyMarkup: keyboard,
	})
	if err != nil {
		return errors.WithMessage(err, "SendMessage")
	}

	err = e.s.SendSticker(chatID, stickers.Cat)
	if err != nil {
		return entity.ErrSendSticker
	}

	return nil
}

func (e *Event) EventHistory(message *telebot.Message) (string, error) {
	chatID := message.Chat.ID
	ids, err := e.db.GetEventsID(chatID, message)
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
	url := fmt.Sprintf("%s?chat_id=%v&id_events=%s", e.conf.URLs.EventsHistory, chatID, eventIDsParam)

	return url, nil
}
