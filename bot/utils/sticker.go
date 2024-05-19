package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

type Sticker struct {
	api *tgbotapi.BotAPI
}

func NewSticker(api *tgbotapi.BotAPI) Sticker {
	return Sticker{
		api: api,
	}
}

func (sticker Sticker) SendSticker(stickerID string, update tgbotapi.Update) error {
	stickerFile := tgbotapi.FileID(stickerID)
	stickers := tgbotapi.NewSticker(update.Message.Chat.ID, stickerFile)
	_, err := sticker.api.Send(stickers)
	if err != nil {
		return errors.WithMessage(err, "sticker error")
	}
	return nil
}
