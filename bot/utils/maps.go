package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Maps struct {
	api *tgbotapi.BotAPI
}

func NewMaps(api *tgbotapi.BotAPI) *Maps {
	return &Maps{
		api: api,
	}
}

func (bot Maps) ShowMaps(chatID int64) {
	url := fmt.Sprintf("https://cr50181-wordpress-j3047.tw1.ru/maps.php?chatID=%d", chatID)
	webappInfo := tgbotapi.WebAppInfo{URL: url}
	btn := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonWebApp("карты", webappInfo),
	)

	reply := tgbotapi.NewInlineKeyboardMarkup(btn)
	msg := tgbotapi.NewMessage(chatID, "карты")
	msg.ReplyMarkup = reply
	bot.api.Send(msg)
}
