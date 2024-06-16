package bot

import (
	"context"
	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"
	"log"
	bot "ride-together-bot/bot/utils"
	"ride-together-bot/conf/stickers"
	"ride-together-bot/db"
)

type Api struct {
	api      *telebot.Bot
	db       *db.DB
	sticker  bot.Sticker
	contact  bot.Contact
	location bot.Location
	event    bot.Event
}

func NewBot(api *telebot.Bot, db *db.DB) *Api {
	newSticker := bot.NewSticker(api)
	newContact := bot.NewContact(api, db, newSticker)
	newLocation := bot.NewLocation(api, db, newSticker)
	newEvent := bot.NewEvent(api, db, newSticker)
	return &Api{
		api:      api,
		db:       db,
		sticker:  newSticker,
		contact:  *newContact,
		location: *newLocation,
		event:    *newEvent,
	}
}

func (bot *Api) Start(ctx context.Context) {
	bot.api.Handle("/start", bot.handleStart)
	bot.api.Handle("/auth", bot.handleAuth(ctx))
	bot.api.Handle("/new_ride", bot.handleNewRide)
	bot.api.Handle("/find", bot.handleFind(ctx))
	bot.api.Handle("/trips_management", bot.handleTripsManagement(ctx))
	bot.api.Handle("/history", bot.handleHistory())

	bot.api.Start()
}

func (bot *Api) handleStart(c telebot.Context) error {
	user := c.Sender()
	msg := "Привет, я бот для поиска попутчиков в любой системе каршеринга.\nПриятной экономии"
	if _, err := bot.api.Send(c.Sender(), msg); err != nil {
		return errors.WithMessage(err, "handleStart")
	}
	return bot.sticker.SendSticker(user.ID, stickers.Start)
}

func (bot *Api) handleAuth(ctx context.Context) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		ok, err := bot.db.IsExists(ctx, c.Sender().Username)
		if err != nil {
			return errors.WithMessage(err, "ошибка проверки существования пользователя")
		}
		if ok {
			c.Send("Пользователь уже зарегистрирован")
			return bot.sticker.SendSticker(chatID, stickers.Shrek)
		}
		c.Send("Регистрация пользователя")
		bot.contact.RequestContact(chatID)
		update := bot.waitForUpdate(c.Bot(), "request_contact")
		if update.Message.Contact == nil {
			return errors.New("не удалось получить данные пользователя")
		}

		log.Printf("Получены данные о пользователе: %+v\n", update.Message.Contact)

		bot.contact.CheckRequestContactReply(ctx, update.Message)
		return nil
	}
}

func (bot *Api) handleNewRide(c telebot.Context) error {
	bot.event.CreateEvent(c.Chat().ID)
	return nil
}

func (bot *Api) waitForUpdate(botUpd *telebot.Bot, updateType string) telebot.Update {
	for update := range botUpd.Updates {
		switch updateType {
		case "request_contact":
			if update.Message.Contact != nil {
				return update
			}
		case "location":
			if update.Message.Location != nil {
				return update
			}
		}
	}
	return telebot.Update{}
}

func (bot *Api) handleFind(ctx context.Context) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		bot.location.GeolocationRequest(c.Chat().ID)

		log.Println("Ожидание обновления с геолокацией...")

		update := bot.waitForUpdate(c.Bot(), "location")
		if update.Message.Location == nil {
			return errors.New("не удалось получить геолокацию")
		}

		log.Printf("Получено местоположение: %+v\n", update.Message.Location)

		url, err := bot.location.HandleLocationUpdate(ctx, update.Message)
		if err != nil {
			return errors.WithMessage(err, "ошибка обработки местоположения")
		}

		webappInfo := &telebot.WebApp{URL: url}
		btn := telebot.Btn{Text: "Поездки", WebApp: webappInfo}
		keyboard := &telebot.ReplyMarkup{ResizeKeyboard: true}
		keyboard.Reply(keyboard.Row(btn))

		return c.Send("Список поездок в радиусе 1км", keyboard)
	}
}

func (bot *Api) handleTripsManagement(ctx context.Context) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		return bot.event.TripsManagement(ctx, c.Message())
	}
}

func (bot *Api) handleHistory() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		usr := c.Sender()
		url, err := bot.event.EventHistory(c.Message())
		if err != nil {
			return errors.WithMessage(err, "ошибка получения истории событий")
		}
		webappInfo := &telebot.WebApp{URL: url}
		btn := telebot.Btn{Text: "История", WebApp: webappInfo}
		keyboard := &telebot.ReplyMarkup{ResizeKeyboard: true}
		keyboard.Reply(keyboard.Row(btn))
		if err := c.Send("ㅤ", keyboard); err != nil {
			return err
		}
		return bot.sticker.SendSticker(usr.ID, stickers.Cat)
	}
}
