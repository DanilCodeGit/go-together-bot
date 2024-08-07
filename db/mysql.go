package db

import (
	"context"
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"math/rand"
	"os"
	"regexp"
	"ride-together-bot/entity"
	"strings"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"ride-together-bot/domain"
)

type DB struct {
	Conn *sql.DB
}

func NewDataBase(dsn string) (*DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v", err)
		return nil, errors.WithMessage(err, "connection")
	}
	err = db.Ping()
	if err != nil {
		return nil, errors.WithMessage(err, "Error pinging database")
	}
	log.Println("Successful database connection")

	return &DB{Conn: db}, nil
}

func (conn DB) GetAllDataFromEvents(ctx context.Context, departureAddress string) ([]entity.Event, error) {
	query := `SELECT * FROM events WHERE departure_address = ?`
	data := make([]entity.Event, 0)
	rows, err := conn.Conn.QueryContext(ctx, query, departureAddress)
	if err != nil {
		return nil, errors.WithMessage(err, "Ошибка выполнения запроса")
	}
	defer rows.Close()

	for rows.Next() {
		var event entity.Event
		var dateOfTrip []byte
		var costPerPerson sql.NullFloat64 // Используем sql.NullFloat64 для costPerPerson
		err := rows.Scan(
			&event.IDEvent,
			&dateOfTrip,
			&event.AvailableSeats,
			&event.TripCost,
			&costPerPerson, // Сканируем в sql.NullFloat64
			&event.DepartureAddress,
			&event.ArrivalAddress,
			&event.CarNumber,
			&event.UserID,
			&event.DriverName,
			&event.Status,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "Ошибка сканирования данных")
		}

		// Ручное преобразование dateOfTrip
		if dateOfTrip != nil {
			parsedDate, err := time.Parse("2006-01-02 15:04:05", string(dateOfTrip))
			if err != nil {
				return nil, errors.WithMessage(err, "Ошибка парсинга date_of_trip")
			}
			event.DateOfTrip = sql.NullTime{Time: parsedDate, Valid: true}
		} else {
			event.DateOfTrip = sql.NullTime{Valid: false}
		}

		// Проверка на нулевую или NULL стоимость на человека
		if !costPerPerson.Valid || costPerPerson.Float64 == 0 {
			event.CostPerPerson = 0
		}

		// Присваиваем значение из sql.NullFloat64 в обычное поле структуры Event
		event.CostPerPerson = float32(costPerPerson.Float64)

		data = append(data, event)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WithMessage(err, "Ошибка извлечения данных")
	}

	return data, nil
}

func latinOnly(name string) string {
	// Создаем регулярное выражение, которое находит только латинские символы
	reg := regexp.MustCompile("[^a-zA-Z]+")

	// Преобразуем все символы, кроме латинских букв, в пробелы
	cleaned := reg.ReplaceAllString(name, " ")

	// Удаляем пробелы из начала и конца строки
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

func (conn DB) Registration(ctx context.Context, message *telebot.Message) error {
	query := `INSERT INTO users (id, name, phone, login, chatid) VALUES (?, ?, ?, ?, ?)`
	_, err := conn.Conn.ExecContext(ctx, query, rand.Int(), latinOnly(message.Sender.FirstName), message.Contact.PhoneNumber, message.Sender.Username, message.Chat.ID)
	if err != nil {
		return errors.WithMessage(err, "Registration error")
	}
	return nil
}

func (conn DB) IsExists(ctx context.Context, login string) (bool, error) {
	query := `SELECT phone FROM users WHERE login = ?`
	var res domain.User
	err := conn.Conn.QueryRowContext(ctx, query, login).Scan(&res.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Record not found
			return false, nil
		}
		return false, errors.WithMessage(err, "Error checking if user exists")
	}

	return true, nil
}

func (conn DB) UpsertLocation(message *telebot.Message) error {
	userId, err := conn.GetUserID(message)
	if err != nil {
		return errors.WithMessage(err, "get user id")
	}

	query := `INSERT INTO user_location (user_id, latitude, longitude, created_at) 
               VALUES (?, ?, ?, ?) 
               ON DUPLICATE KEY UPDATE user_id = values(user_id)`

	_, err = conn.Conn.ExecContext(context.Background(), query,
		userId, message.Location.Lat, message.Location.Lng, time.Now())
	if err != nil {
		return errors.WithMessage(err, "Error upserting location")
	}
	return nil
}

func (conn DB) GetUserID(message *telebot.Message) (int64, error) {
	query := `SELECT id FROM users WHERE chatID = ?`
	var id int64
	err := conn.Conn.QueryRowContext(context.Background(), query, message.Chat.ID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.WithMessage(err, "Error getting user ID")
		}
	}
	return id, nil
}

func (conn DB) IsDriver(ctx context.Context, message *telebot.Message) (bool, error) {
	userId, err := conn.GetUserID(message)
	if err != nil {
		return false, err
	}

	query := `SELECT COUNT(*) FROM events WHERE user_id = ?`
	var count int
	err = conn.Conn.QueryRowContext(ctx, query, userId).Scan(&count)
	if err != nil {
		return false, errors.WithMessage(err, "Error checking if user exists")
	}
	return count > 0, nil
}

func (conn DB) GetLastLocation(ctx context.Context, userId int64) (entity.Coordinates, error) {
	query := `SELECT user_location.latitude, user_location.longitude
            FROM user_location
            WHERE user_id = ?
            ORDER BY id DESC
            LIMIT 1`

	var coordinates entity.Coordinates
	_ = conn.Conn.QueryRowContext(ctx, query, userId).Scan(&coordinates.Lat, &coordinates.Lon)

	return coordinates, nil
}

func (conn DB) GetAllDepartureAddresses() ([]string, error) {
	departures := make([]string, 0)
	query := `select departure_address from events where status = 1`
	rows, _ := conn.Conn.QueryContext(context.Background(), query)
	defer rows.Close()
	for rows.Next() {
		var departureAddress string
		err := rows.Scan(&departureAddress)
		if err != nil {
			return nil, errors.WithMessage(err, "Error scanning departure address")
		}
		departures = append(departures, departureAddress)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WithMessage(err, "Error scanning departure address")
	}
	return departures, nil
}

func (conn DB) GetEventsID(chatID int64, message *telebot.Message) ([]int64, error) {
	query := `select event_id FROM passengers WHERE user_chatID = ?`
	ids := make([]int64, 0)
	rows, err := conn.Conn.QueryContext(context.Background(), query, chatID)
	if err != nil {
		return nil, errors.WithMessage(err, "Error getting events ID")
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return nil, errors.WithMessage(err, "Error scanning events ID")
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WithMessage(err, "rows Error scanning events ID")
	}
	if len(ids) == 0 {
		return nil, errors.WithMessage(err, "events ID len == 0")
	}
	userID, _ := conn.GetUserID(message)
	query = `select id_events FROM events WHERE user_id = ?`

	rows, err = conn.Conn.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, errors.WithMessage(err, "Error getting events ID")
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return nil, errors.WithMessage(err, "Error scanning events ID")
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WithMessage(err, "rows Error scanning events ID")
	}
	if len(ids) == 0 {
		return nil, errors.WithMessage(err, "events ID len == 0")
	}

	return ids, nil
}
