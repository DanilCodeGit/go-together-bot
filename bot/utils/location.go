package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"ride-together-bot/db"
)

type Location struct {
	api *tgbotapi.BotAPI
	db  *db.DB
}

func NewLocation(api *tgbotapi.BotAPI, db *db.DB) *Location {
	return &Location{
		api: api,
		db:  db,
	}
}

func (location Location) GeolocationRequest(chatID int64) {
	btn := tgbotapi.NewKeyboardButtonLocation("запрос геолокации")
	keyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn})
	msg := tgbotapi.NewMessage(chatID, "Отправьте вашу геолокацию")
	msg.ReplyMarkup = keyboard
	location.api.Send(msg)
}

func (location Location) HandleLocationUpdate(update tgbotapi.Update) error {
	if update.Message.Location != nil {
		latitude := update.Message.Location.Latitude
		longitude := update.Message.Location.Longitude
		log.Printf("latitude: %f, longitude: %f", latitude, longitude)
		err := location.db.UpsertLocation(update)
		if err != nil {
			return errors.WithMessage(err, "insert location error")
		}
	}
	return nil
}
