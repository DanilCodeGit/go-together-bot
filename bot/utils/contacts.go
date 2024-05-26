package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ride-together-bot/conf/stickers"
	"ride-together-bot/db"
)

type Contact struct {
	api     *tgbotapi.BotAPI
	db      *db.DB
	sticker Sticker
}

func NewContact(api *tgbotapi.BotAPI, db *db.DB, sticker Sticker) *Contact {
	return &Contact{
		api:     api,
		db:      db,
		sticker: sticker,
	}
}

// Проверка принятого контакта
func (e Contact) CheckRequestContactReply(ctx context.Context, update tgbotapi.Update) {
	if update.Message.Contact != nil {
		if update.Message.Contact.UserID == update.Message.From.ID {
			err := e.db.Registration(ctx, update)
			if err != nil {
				return
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо!")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false) // Убираем клавиатуру
			e.api.Send(msg)
			err = e.sticker.SendSticker(stickers.Cat, update)
			if err != nil {
				return
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер телефона, который вы предоставили, принадлежит не вам!")
			e.api.Send(msg)
			e.RequestContact(update.Message.Chat.ID)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Если вы не предоставите ваш номер телефона, вы не сможете пользоваться системой!")
		e.api.Send(msg)
		e.RequestContact(update.Message.Chat.ID)
	}
}

func (e Contact) RequestContact(chatID int64) {
	// Создаём сообщение
	requestContactMessage := tgbotapi.NewMessage(chatID, "Согласны ли вы предоставить ваш номер телефона для регистрации в системе?")

	// Создаём кнопку отправки контакта
	acceptButton := tgbotapi.NewKeyboardButtonContact("Да")
	declineButton := tgbotapi.NewKeyboardButton("Нет")

	// Создаём клавиатуру
	requestContactReplyKeyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{acceptButton, declineButton})
	requestContactMessage.ReplyMarkup = requestContactReplyKeyboard
	e.api.Send(requestContactMessage)
}
