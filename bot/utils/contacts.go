package bot

import (
	"context"
	"fmt"
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
func (contact Contact) CheckRequestContactReply(ctx context.Context, update tgbotapi.Update) {
	if update.Message.Contact != nil {
		if update.Message.Contact.UserID == update.Message.From.ID {
			err := contact.db.Registration(ctx, update)
			if err != nil {
				return
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо!")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false) // Убираем клавиатуру
			contact.api.Send(msg)
			err = contact.sticker.SendSticker(stickers.Cat, update)
			if err != nil {
				return
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер телефона, который вы предоставили, принадлежит не вам!")
			contact.api.Send(msg)
			contact.RequestContact(update.Message.Chat.ID)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Если вы не предоставите ваш номер телефона, вы не сможете пользоваться системой!")
		contact.api.Send(msg)
		contact.RequestContact(update.Message.Chat.ID)
	}
}

func (contact Contact) InlineContact(chatID int64) {
	// Формируем URL с параметром запроса с именем пользователя
	url := fmt.Sprintf("https://cr50181-wordpress-j3047.tw1.ru/create_event_page.php?chatID=%d", chatID)
	webappInfo := tgbotapi.WebAppInfo{URL: url}
	btn := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonWebApp("my btn", webappInfo),
	)
	reply := tgbotapi.NewInlineKeyboardMarkup(btn)
	msg := tgbotapi.NewMessage(chatID, "Кнопка")
	msg.ReplyMarkup = reply
	contact.api.Send(msg)
}

func (contact Contact) RequestContact(chatID int64) {
	// Создаём сообщение
	requestContactMessage := tgbotapi.NewMessage(chatID, "Согласны ли вы предоставить ваш номер телефона для регистрации в системе?")

	// Создаём кнопку отправки контакта
	acceptButton := tgbotapi.NewKeyboardButtonContact("Да")
	declineButton := tgbotapi.NewKeyboardButton("Нет")

	// Создаём клавиатуру
	requestContactReplyKeyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{acceptButton, declineButton})
	requestContactMessage.ReplyMarkup = requestContactReplyKeyboard
	contact.api.Send(requestContactMessage)
}
