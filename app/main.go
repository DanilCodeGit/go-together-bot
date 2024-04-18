package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"ride-together-bot/conf"
	"ride-together-bot/db"
	"ride-together-bot/entitiy"
)

type BotAPI struct {
	API *tgbotapi.BotAPI
	db  *db.DB
}

func NewBot(api *tgbotapi.BotAPI, db *db.DB) BotAPI {
	api.Debug = true
	return BotAPI{
		API: api,
		db:  db,
	}
}

func (bot BotAPI) requestContact(chatID int64) {
	// Создаём сообщение
	requestContactMessage := tgbotapi.NewMessage(chatID, "Согласны ли вы предоставить ваш номер телефона для регистрации в системе?")

	// Создаём кнопку отправки контакта
	acceptButton := tgbotapi.NewKeyboardButtonContact("Да")
	declineButton := tgbotapi.NewKeyboardButton("Нет")

	// Создаём клавиатуру
	requestContactReplyKeyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{acceptButton, declineButton})
	requestContactMessage.ReplyMarkup = requestContactReplyKeyboard
	bot.API.Send(requestContactMessage) // Отправляем сообщение
}

// Проверка принятого контакта
func (bot BotAPI) checkRequestContactReply(ctx context.Context, update tgbotapi.Update, user entitiy.User) {
	if update.Message.Contact != nil { // Проверяем, содержит ли сообщение контакт
		if update.Message.Contact.UserID == update.Message.From.ID { // Проверяем действительно ли это контакт отправителя
			err := bot.db.Registration(ctx, user, update)
			if err != nil {
				return
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо!")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false) // Убираем клавиатуру
			bot.API.Send(msg)
			stickerID := tgbotapi.FileID("CAACAgIAAxkBAAEL3MtmEnBdKVlozRT-Lm9SdbTUFGwaKQACNRMAAq07CEgEFLhcMipUIDQE")
			sticker := tgbotapi.NewSticker(update.Message.Chat.ID, stickerID)
			_, err = bot.API.Send(sticker)
			if err != nil {
				return
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер телефона, который вы предоставили, принадлежит не вам!")
			bot.API.Send(msg)
			bot.requestContact(update.Message.Chat.ID)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Если вы не предоставите ваш номер телефона, вы не сможете пользоваться системой!")
		bot.API.Send(msg)
		bot.requestContact(update.Message.Chat.ID)
	}
}

func (bot BotAPI) inlineContact(chatID int64) {
	// Формируем URL с параметром запроса с именем пользователя
	url := fmt.Sprintf("https://cr50181-wordpress-j3047.tw1.ru/create_event_page.php?chatID=%d", chatID)
	webappInfo := tgbotapi.WebAppInfo{URL: url}
	btn := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonWebApp("my btn", webappInfo),
	)

	reply := tgbotapi.NewInlineKeyboardMarkup(btn)
	msg := tgbotapi.NewMessage(chatID, "Кнопка")
	msg.ReplyMarkup = reply
	bot.API.Send(msg)
}

func (bot BotAPI) Updates(ctx context.Context, user entitiy.User) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.API.GetUpdatesChan(u)
	for update := range updates {

		chatID := update.Message.Chat.ID
		commands := update.Message.Command()
		msg := tgbotapi.NewMessage(chatID, update.Message.Text)

		ok, err := bot.db.IsExists(ctx, update.Message.Chat.UserName)
		if err != nil {
			return errors.WithMessage(err, "is exists")
		}
		if !ok {
			msg.Text = "Требуется регистрация"
			bot.API.Send(msg)
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			bot.requestContact(chatID)
			for u := range updates {
				bot.checkRequestContactReply(ctx, u, user)
				break
			}
			//continue
		}
		switch commands {
		case "start":
			msg.Text = "Привет, я бот для поиска попутчиков в любой системе каршеринга.\nПриятной экономии"
			bot.API.Send(msg)

		case "auth":
			ok, err := bot.db.IsExists(ctx, update.Message.Chat.UserName)
			if err != nil {
				return errors.WithMessage(err, "is exists")
			}
			if ok {
				msg.Text = "Пользователь уже зарегистрирован"
				stickerID := tgbotapi.FileID("CAACAgIAAxkBAAEL3MlmEm_CzZTbjq297QhPpvUjGIDQ8gACTBQAAuxLAUh_I_vdpHUhwzQE")
				sticker := tgbotapi.NewSticker(chatID, stickerID)
				_, err := bot.API.Send(sticker)
				if err != nil {
					return errors.WithMessage(err, "sticker")
				}
				bot.API.Send(msg)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				continue
			}
			msg.Text = "Регистрация пользователя"
			bot.API.Send(msg)
			bot.requestContact(chatID)
			for u := range updates {
				bot.checkRequestContactReply(ctx, u, user)
				break
			}

		case "create_event":
			bot.inlineContact(chatID)

		case "find_ride":
			continue
		}

	}
	return nil
}

func main() {
	ctx := context.Background()
	// Инициализация базы данных
	conn, err := db.NewDataBase(ctx, conf.DSN)
	if err != nil {
		log.Panic(errors.WithMessage(err, "ошибка инициализации БД"))
	}

	// Инициализация бота
	newBot, err := tgbotapi.NewBotAPI(conf.TelegramBotApiKey)
	if err != nil {
		log.Panic(err)
	}

	// Создание экземпляра BotAPI
	bot := NewBot(newBot, conn)

	// Создание экземпляра пользователя
	var user entitiy.User

	// Запуск обновлений
	err = bot.Updates(ctx, user)
	if err != nil {
		log.Panic(err)
	}

}
