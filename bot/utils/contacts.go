package bot

import (
	"context"
	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"
	"ride-together-bot/conf/stickers"
	"ride-together-bot/db"
	"ride-together-bot/entity"
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
func (c Contact) CheckRequestContactReply(ctx context.Context, message *telebot.Message) error {
	if message.Contact == nil {
		reply := "Если вы не предоставите ваш номер телефона, вы не сможете пользоваться системой!"
		_, err := c.api.Send(message.Sender, reply)
		if err != nil {
			return entity.ErrSendMsg
		}
		c.RequestContact(message.Chat.ID)

		return nil
	}
	if message.Contact.UserID == message.Sender.ID {
		err := c.db.Registration(ctx, message)
		if err != nil {
			return errors.WithMessage(err, "registration contact")
		}
		reply := "Спасибо!"
		_, err = c.api.Send(message.Sender, reply, &telebot.ReplyMarkup{RemoveKeyboard: true})
		if err != nil {
			return entity.ErrSendMsg
		}
		err = c.sticker.SendSticker(message.Chat.ID, stickers.Cat)
		if err != nil {
			return entity.ErrSendSticker
		}
	}

	return nil
}

func (c Contact) RequestContact(chatID int64) error {
	requestContactMessage := "Согласны ли вы предоставить ваш номер телефона для регистрации в системе?"

	acceptButton := telebot.ReplyButton{Text: "Yes", Contact: true}
	declineButton := telebot.ReplyButton{Text: "No"}

	replyKeyboard := &telebot.ReplyMarkup{
		ReplyKeyboard: [][]telebot.ReplyButton{
			{acceptButton, declineButton},
		},
		ResizeKeyboard: true,
	}

	_, err := c.api.Send(telebot.ChatID(chatID), requestContactMessage, replyKeyboard)
	if err != nil {
		return entity.ErrSendMsg
	}
	return nil
}
