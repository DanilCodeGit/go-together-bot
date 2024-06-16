package bot

import (
	"context"
	"gopkg.in/telebot.v3"
	"ride-together-bot/conf/stickers"
	"ride-together-bot/db"
)

type Contact struct {
	api     *telebot.Bot
	db      *db.DB
	sticker Sticker
}

func NewContact(api *telebot.Bot, db *db.DB, sticker Sticker) *Contact {
	return &Contact{
		api:     api,
		db:      db,
		sticker: sticker,
	}
}

// Проверка принятого контакта
func (c Contact) CheckRequestContactReply(ctx context.Context, message *telebot.Message) {
	if message.Contact != nil {
		if message.Contact.UserID == message.Sender.ID {
			err := c.db.Registration(ctx, message)
			if err != nil {
				return
			}
			reply := "Спасибо!"
			c.api.Send(message.Sender, reply, &telebot.ReplyMarkup{RemoveKeyboard: true})
			err = c.sticker.SendSticker(message.Chat.ID, stickers.Cat)
			if err != nil {
				return
			}
		}
	} else {
		reply := "Если вы не предоставите ваш номер телефона, вы не сможете пользоваться системой!"
		c.api.Send(message.Sender, reply)
		c.RequestContact(message.Chat.ID)
	}
}

func (c Contact) RequestContact(chatID int64) {
	requestContactMessage := "Согласны ли вы предоставить ваш номер телефона для регистрации в системе?"

	acceptButton := telebot.ReplyButton{Text: "Yes", Contact: true}
	declineButton := telebot.ReplyButton{Text: "No"}

	replyKeyboard := &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{acceptButton, declineButton},
		},
		ResizeKeyboard: true,
	}

	c.api.Send(telebot.ChatID(chatID), requestContactMessage, replyKeyboard)
}
