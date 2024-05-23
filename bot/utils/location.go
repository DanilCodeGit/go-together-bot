package bot

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"math"
	"net/http"
	"net/url"
	"ride-together-bot/conf"
	"ride-together-bot/db"
	"strconv"
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

func (l Location) GeolocationRequest(chatID int64) {
	btn := tgbotapi.NewKeyboardButtonLocation("запрос геолокации")
	keyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn})
	msg := tgbotapi.NewMessage(chatID, "Отправьте вашу геолокацию")
	msg.ReplyMarkup = keyboard
	l.api.Send(msg)
}

func (l Location) HandleLocationUpdate(update tgbotapi.Update) error {
	if update.Message.Location != nil {
		latitude := update.Message.Location.Latitude
		longitude := update.Message.Location.Longitude
		log.Printf("latitude: %f, longitude: %f", latitude, longitude)
		err := l.db.UpsertLocation(update)
		if err != nil {
			return errors.WithMessage(err, "insert l error")
		}
	}
	return nil
}

/////////// Попадание точки в заданный радиус

// toRadians преобразует градусы в радианы
func (l Location) toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

// haversine вычисляет расстояние между двумя точками на поверхности Земли
func (l Location) haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := l.toRadians(lat2 - lat1)
	dLon := l.toRadians(lon2 - lon1)

	lat1 = l.toRadians(lat1)
	lat2 = l.toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return conf.EarthRadius * c
}

// isPointInCircle проверяет, находится ли точка в круге радиусом 1 км от заданной координаты
func (l Location) isPointInCircle(centerLat, centerLon, pointLat, pointLon float64) bool {
	distance := l.haversine(centerLat, centerLon, pointLat, pointLon)
	return distance <= 1
}

/// Геокодер

type Coordinates struct {
	Lat float64 `json:"lat,string"`
	Lon float64 `json:"lon,string"`
}

// getCoordinates получает координаты по адресу с использованием Nominatim API
func (l Location) getCoordinates(address string) (*Coordinates, error) {
	// Кодируем адрес для использования в URL
	encodedAddress := url.QueryEscape(address)
	// Формируем URL для запроса к Nominatim API
	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?format=json&q=%s", encodedAddress)

	// Выполняем HTTP GET запрос
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус код ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неожиданный статус код: %d", resp.StatusCode)
	}

	// Декодируем ответ
	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("ошибка при декодировании ответа: %v", err)
	}

	// Проверяем, есть ли результаты
	if len(results) == 0 {
		return nil, fmt.Errorf("координаты не найдены для адреса: %s", address)
	}

	// Парсим широту и долготу
	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге широты: %v", err)
	}
	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге долготы: %v", err)
	}

	return &Coordinates{Lat: lat, Lon: lon}, nil
}
