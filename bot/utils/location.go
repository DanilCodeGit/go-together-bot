package bot

import (
	"context"
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
	"ride-together-bot/entity"
	"strconv"
	"strings"
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

func (l Location) HandleLocationUpdate(ctx context.Context, update tgbotapi.Update) (string, error) {
	okAddresses := make([]string, 0)
	okEvents := make([]entity.Event, 0)
	chatID := update.Message.Chat.ID
	var url string
	if update.Message.Location != nil {
		latitude := update.Message.Location.Latitude
		longitude := update.Message.Location.Longitude
		log.Printf("latitude: %f, longitude: %f", latitude, longitude)
		err := l.db.UpsertLocation(update)
		if err != nil {
			return "", errors.WithMessage(err, "insert l error")
		}

		departureAddress, _ := l.db.GetAllDepartureAddresses()
		for _, address := range departureAddress {
			coordinates, err := l.GetCoordinates(address)
			if err != nil {
				log.Printf("get coordinates error: %e", err)
				continue
			}
			ok := l.IsPointInCircle(latitude, longitude, coordinates.Lat, coordinates.Lon)
			if ok {
				okAddresses = append(okAddresses, address)
			}
		}
		for _, address := range okAddresses {
			events, err := l.db.GetAllDataFromEvents(ctx, address)
			if err != nil {
				return "", errors.WithMessage(err, "get all events error")
			}
			okEvents = append(okEvents, events...)
		}
		eventIDs := make([]string, len(okEvents))
		for i, event := range okEvents {
			eventIDs[i] = strconv.Itoa(event.IDEvent)
		}
		eventIDsParam := strings.Join(eventIDs, ",")
		url = fmt.Sprintf("https://cr50181-wordpress-j3047.tw1.ru/maps.php?chat_id=%v&id_events=%s", chatID, eventIDsParam)
	}
	return url, nil
}

/////////// Попадание точки в заданный радиус

// isPointInCircle проверяет, находится ли точка в круге радиусом 1 км от заданной координаты
func (l Location) IsPointInCircle(centerLat, centerLon, pointLat, pointLon float64) bool {
	distance := l.haversine(centerLat, centerLon, pointLat, pointLon)
	return distance <= 1
}

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

////////////////////// Геокодер

type Coordinates struct {
	Lat float64 `json:"lat,string"`
	Lon float64 `json:"lon,string"`
}

// getCoordinates получает координаты по адресу с использованием Nominatim API
func (l Location) GetCoordinates(address string) (*Coordinates, error) {
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
		//log.Printf("координаты не найдены для адреса: %s", address)
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
