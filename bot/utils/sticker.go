package bot

import (
	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"
)

type Sticker struct {
	api *telebot.Bot
}

func NewSticker(api *telebot.Bot) Sticker {
	return Sticker{
		api: api,
	}
}

func (sticker Sticker) SendSticker(chatID int64, stickerID string) error {
	stickerFile := &telebot.Sticker{File: telebot.File{FileID: stickerID}}
	_, err := sticker.api.Send(telebot.ChatID(chatID), stickerFile)
	if err != nil {
		return errors.WithMessage(err, "sticker error")
	}
	return nil
}
