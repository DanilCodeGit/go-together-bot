package bot

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"
	"log"
	bot "ride-together-bot/bot/utils"
	"ride-together-bot/conf"
	"ride-together-bot/conf/stickers"
	"ride-together-bot/db"
	"ride-together-bot/entity"
)

type Api struct {
	api      *telebot.Bot
	db       *db.DB
	sticker  bot.Sticker
	contact  bot.Contact
	location bot.Location
	event    bot.Event
}

func NewBot(conf *conf.Config, api *telebot.Bot, db *db.DB) *Api {
	newSticker := bot.NewSticker(api)
	newContact := bot.NewContact(api, db, newSticker)
	newLocation := bot.NewLocation(api, db, newSticker, conf)
	newEvent := bot.NewEvent(conf, api, db, newSticker)
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
	hi := fmt.Sprintf("Привет, %s. Я бот для поиска попутчиков в любой системе каршеринга.\nПриятной экономии!\n\n", user.FirstName)
	msg := hi
	msg += "⚠️*Внимание!* Перед началом работы требуется авторизация."

	_, err := bot.api.Send(c.Sender(), msg, &telebot.SendOptions{
		ParseMode: "Markdown",
	})
	if err != nil {
		return errors.WithMessage(err, "handleStart")
	}

	err = bot.sticker.SendSticker(user.ID, stickers.Start)
	if err != nil {
		return errors.WithMessage(err, "send sticker")
	}
	return nil
}

func (bot *Api) handleAuth(ctx context.Context) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		ok, err := bot.db.IsExists(ctx, c.Sender().Username)
		if err != nil {
			return errors.WithMessage(err, "ошибка проверки существования пользователя")
		}
		if ok {
			err = c.Send("Пользователь уже зарегистрирован")
			if err != nil {
				return entity.ErrSendMsg
			}
			return bot.sticker.SendSticker(chatID, stickers.Shrek)
		}

		err = c.Send("Регистрация пользователя")
		if err != nil {
			return entity.ErrSendMsg
		}

		err = bot.contact.RequestContact(chatID)
		if err != nil {
			return errors.WithMessage(err, "requestContact")
		}
		update := bot.waitForUpdate(c.Bot(), "request_contact")
		if update.Message.Contact == nil {
			return errors.New("не удалось получить данные пользователя")
		}

		log.Printf("Получены данные о пользователе: %+v\n", update.Message.Contact)

		err = bot.contact.CheckRequestContactReply(ctx, update.Message)
		if err != nil {
			return errors.WithMessage(err, "checkRequestContactReply")
		}
		return nil
	}
}

func (bot *Api) handleNewRide(c telebot.Context) error {
	ok, err := bot.db.IsExists(context.Background(), c.Sender().Username)
	if err != nil {
		return errors.WithMessage(err, "ошибка проверки существования пользователя")
	}

	if !ok {
		msg := entity.NeedAuth

		_, err = bot.api.Send(c.Sender(), msg, &telebot.SendOptions{
			ParseMode: "Markdown",
		})
		if err != nil {
			return errors.WithMessage(err, "handleStart")
		}
		return nil
	}

	err = bot.event.CreateEvent(c.Chat().ID)
	if err != nil {
		return errors.WithMessage(err, "create event")
	}
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
		ok, err := bot.db.IsExists(ctx, c.Sender().Username)
		if err != nil {
			return errors.WithMessage(err, "ошибка проверки существования пользователя")
		}
		if !ok {
			msg := entity.NeedAuth

			if _, err := bot.api.Send(c.Sender(), msg, &telebot.SendOptions{
				ParseMode: "Markdown",
			}); err != nil {
				return errors.WithMessage(err, "handleStart")
			}
			return nil
		}

		err = bot.location.GeolocationRequest(c.Chat().ID)
		if err != nil {
			return errors.WithMessage(err, "geolocationRequest")
		}

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
		ok, err := bot.db.IsExists(ctx, c.Sender().Username)
		if err != nil {
			return errors.WithMessage(err, "ошибка проверки существования пользователя")
		}
		if !ok {
			msg := entity.NeedAuth

			_, err := bot.api.Send(c.Sender(), msg, &telebot.SendOptions{
				ParseMode: "Markdown",
			})
			if err != nil {
				return entity.ErrSendMsg
			}
			return nil
		}
		return bot.event.TripsManagement(ctx, c.Message())
	}
}

func (bot *Api) handleHistory() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ok, err := bot.db.IsExists(context.Background(), c.Sender().Username)
		if err != nil {
			return errors.WithMessage(err, "ошибка проверки существования пользователя")
		}

		if !ok {
			msg := entity.NeedAuth

			if _, err := bot.api.Send(c.Sender(), msg, &telebot.SendOptions{
				ParseMode: "Markdown",
			}); err != nil {
				return errors.WithMessage(err, "handleStart")
			}
			return nil
		}

		usr := c.Sender()
		url, err := bot.event.EventHistory(c.Message())
		if err != nil {
			return errors.WithMessage(err, "ошибка получения истории событий")
		}

		webappInfo := &telebot.WebApp{URL: url}
		btn := telebot.Btn{Text: "История", WebApp: webappInfo}
		keyboard := &telebot.ReplyMarkup{ResizeKeyboard: true}
		keyboard.Reply(keyboard.Row(btn))

		err = c.Send("ㅤ", keyboard)
		if err != nil {
			return errors.WithMessage(err, "history")
		}
		err = bot.sticker.SendSticker(usr.ID, stickers.Cat)
		if err != nil {
			return entity.ErrSendSticker
		}
		return nil
	}
}
