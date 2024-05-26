package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	bot "ride-together-bot/bot/utils"
	"ride-together-bot/conf/stickers"
	"ride-together-bot/db"
)

type BotApi struct {
	api      *tgbotapi.BotAPI
	db       *db.DB
	sticker  bot.Sticker
	contact  bot.Contact
	location bot.Location
	event    bot.Event
}

func NewBot(api *tgbotapi.BotAPI, db *db.DB) *BotApi {
	newSticker := bot.NewSticker(api)
	newContact := bot.NewContact(api, db, newSticker)
	newLocation := bot.NewLocation(api, db)
	newEvent := bot.NewEvent(api, db, newSticker)
	api.Debug = true
	return &BotApi{
		api:      api,
		db:       db,
		sticker:  newSticker,
		contact:  *newContact,
		location: *newLocation,
		event:    *newEvent,
	}
}

func (bot BotApi) Updates(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)

	updates := bot.api.GetUpdatesChan(u)

	for update := range updates {
		chatID := update.Message.Chat.ID
		commands := update.Message.Command()
		msg := tgbotapi.NewMessage(chatID, update.Message.Text)

		ok, err := bot.db.IsExists(ctx, update.Message.Chat.UserName)
		if err != nil {
			return errors.WithMessage(err, "is exists")
		}
		if !ok {
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			bot.contact.RequestContact(chatID)
			update = bot.waitForUpdate(updates, "contact")
			bot.contact.CheckRequestContactReply(ctx, update)
			continue
		}
		switch commands {
		case "start":
			msg.Text = "Привет, я бот для поиска попутчиков в любой системе каршеринга.\nПриятной экономии"
			bot.api.Send(msg)
		case "auth":
			ok, err := bot.db.IsExists(ctx, update.Message.Chat.UserName)
			if err != nil {
				return errors.WithMessage(err, "is exists")
			}
			if ok {
				msg.Text = "Пользователь уже зарегистрирован"
				err = bot.sticker.SendSticker(stickers.Shrek, update)
				if err != nil {
					return err
				}
				bot.api.Send(msg)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				continue
			}
			msg.Text = "Регистрация пользователя"
			bot.api.Send(msg)
			bot.contact.RequestContact(chatID)
			update = bot.waitForUpdate(updates, "contact")
			bot.contact.CheckRequestContactReply(ctx, update)

		case "create_event":
			bot.event.CreateEvent(chatID, update)

		case "find_ride":
			bot.location.GeolocationRequest(chatID)
			update = bot.waitForUpdate(updates, "location")
			url, err := bot.location.HandleLocationUpdate(ctx, update)
			if err != nil {
				return errors.WithMessage(err, "handle location error")
			}
			webappInfo := tgbotapi.WebAppInfo{URL: url}
			btn := tgbotapi.NewKeyboardButtonWebApp("Поездки", webappInfo)
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(btn),
			)
			msg := tgbotapi.NewMessage(chatID, "Список поездок в радиусе 1км")
			msg.ReplyMarkup = keyboard
			bot.api.Send(msg)
		case "active_events":
			err := bot.event.ActiveEvents(ctx, update)
			if err != nil {
				return errors.WithMessage(err, "get active events error")
			}
		case "history":
			url, err := bot.event.EventHistory(update)
			if err != nil {
				return errors.WithMessage(err, "get history error")
			}
			webappInfo := tgbotapi.WebAppInfo{URL: url}
			btn := tgbotapi.NewKeyboardButtonWebApp("История", webappInfo)
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(btn),
			)
			msg := tgbotapi.NewMessage(chatID, "ㅤ")
			msg.ReplyMarkup = keyboard
			bot.api.Send(msg)
			bot.sticker.SendSticker(stickers.Cat, update)
		}
	}
	return nil
}

// Метод для ожидания следующего обновления определённого типа
func (bot BotApi) waitForUpdate(updates tgbotapi.UpdatesChannel, updateType string) tgbotapi.Update {
	for update := range updates {
		switch updateType {
		case "contact":
			if update.Message.Contact != nil {
				return update
			}
		case "location":
			if update.Message.Location != nil {
				return update
			}
		}
	}
	return tgbotapi.Update{}
}
