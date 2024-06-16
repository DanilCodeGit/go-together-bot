package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"
	"log"
	"math"
	"net/http"
	"net/url"
	"ride-together-bot/conf"
	"ride-together-bot/conf/stickers"
	"ride-together-bot/db"
	"ride-together-bot/entity"
	"strconv"
	"strings"
)

type Location struct {
	api *telebot.Bot
	db  *db.DB
	s   Sticker
}

func NewLocation(api *telebot.Bot, db *db.DB, s Sticker) *Location {
	return &Location{
		api: api,
		db:  db,
		s:   s,
	}
}

// GeolocationRequest sends a location request to the user
func (l Location) GeolocationRequest(chatID int64) {
	btn := telebot.Btn{Text: "Запрос геолокации", Location: true}
	keyboard := telebot.ReplyMarkup{ResizeKeyboard: true}
	keyboard.Reply(keyboard.Row(btn))
	l.api.Send(telebot.ChatID(chatID), "Пожалуйста, предоставьте свою геолокацию", &keyboard)
	l.s.SendSticker(chatID, stickers.Location)
}

// HandleLocationUpdate handles the user's location update and returns a URL for displaying events on a map
func (l Location) HandleLocationUpdate(ctx context.Context, message *telebot.Message) (string, error) {
	okAddresses := make([]string, 0)
	okEvents := make([]entity.Event, 0)
	chatID := message.Chat.ID
	var u string

	if message.Location != nil {
		latitude := message.Location.Lat
		longitude := message.Location.Lng
		log.Printf("latitude: %f, longitude: %f", latitude, longitude)
		err := l.db.UpsertLocation(message)
		if err != nil {
			return "", errors.WithMessage(err, "insert l error")
		}

		departureAddresses, _ := l.db.GetAllDepartureAddresses()
		for _, address := range departureAddresses {
			coordinates, err := l.GetCoordinates(address)
			if err != nil {
				log.Printf("get coordinates error: %e", err)
				continue
			}
			if l.IsPointInCircle(latitude, longitude, float32(coordinates.Lat), float32(coordinates.Lon)) {
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
		u = fmt.Sprintf("https://cr50181-wordpress-j3047.tw1.ru/maps.php?chat_id=%v&id_events=%s", telebot.ChatID(chatID), eventIDsParam)
	}
	return u, nil
}

// IsPointInCircle checks if a point is within a 1 km radius of a given coordinate
func (l Location) IsPointInCircle(centerLat, centerLon, pointLat, pointLon float32) bool {
	distance := l.haversine(centerLat, centerLon, pointLat, pointLon)
	return distance <= 1
}

// toRadians converts degrees to radians
func (l Location) toRadians(deg float32) float32 {
	return deg * math.Pi / 180
}

// haversine calculates the distance between two points on the Earth's surface
func (l Location) haversine(lat1, lon1, lat2, lon2 float32) float32 {
	dLat := l.toRadians(lat2 - lat1)
	dLon := l.toRadians(lon2 - lon1)

	lat1 = l.toRadians(lat1)
	lat2 = l.toRadians(lat2)

	a := math.Sin(float64(dLat/2))*math.Sin(float64(dLat/2)) +
		math.Sin(float64(dLon/2))*math.Sin(float64(dLon/2))*math.Cos(float64(lat1))*math.Cos(float64(lat2))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return float32(conf.EarthRadius * c)
}

type Coordinates struct {
	Lat float64 `json:"lat,string"`
	Lon float64 `json:"lon,string"`
}

// GetCoordinates gets coordinates by address using Nominatim API
func (l Location) GetCoordinates(address string) (*Coordinates, error) {
	encodedAddress := url.QueryEscape(address)
	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?format=json&q=%s", encodedAddress)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неожиданный статус код: %d", resp.StatusCode)
	}

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("ошибка при декодировании ответа: %v", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("координаты не найдены для адреса: %s", address)
	}

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
