package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"ride-together-bot/entity"
	"strings"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	if err := db.Ping(); err != nil {
		return nil, errors.WithMessage(err, "Error pinging database")
	}
	log.Println("Successful database connection")

	return &DB{Conn: db}, nil
}

func (conn DB) GetAllData(ctx context.Context) ([]entity.User, error) {
	query := `SELECT name, phone, login, chatid FROM users`
	data := make([]entity.User, 0)
	rows, err := conn.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.WithMessage(err, "Error executing query")
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User
		err := rows.Scan(&user.Name, &user.Phone, &user.Login, &user.ChatID)
		if err != nil {
			return nil, errors.WithMessage(err, "Error scanning data")
		}
		data = append(data, user)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WithMessage(err, "Error retrieving data")
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
func (conn DB) Registration(ctx context.Context, update tgbotapi.Update) error {
	query := `INSERT INTO users (id, name, phone, login, chatid) VALUES (?, ?, ?, ?, ?)`
	_, err := conn.Conn.ExecContext(ctx, query, rand.Int(), latinOnly(update.Message.Chat.FirstName), update.Message.Contact.PhoneNumber, update.Message.Chat.UserName, update.Message.Chat.ID)
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

func (conn DB) UpsertLocation(update tgbotapi.Update) error {
	userId, err := conn.getUserID(update)
	if err != nil {
		return err
	}

	query := `INSERT INTO user_location (user_id, latitude, longitude, created_at) 
               VALUES (?, ?, ?, ?) 
               ON DUPLICATE KEY UPDATE user_id = values(user_id)`

	_, err = conn.Conn.ExecContext(context.Background(), query,
		userId, update.Message.Location.Latitude, update.Message.Location.Longitude, time.Now())
	if err != nil {
		return errors.WithMessage(err, "Error upserting location")
	}
	return nil
}

func (conn DB) getUserID(update tgbotapi.Update) (int64, error) {
	query := `SELECT id FROM users WHERE chatID = ?`
	var id int64
	err := conn.Conn.QueryRowContext(context.Background(), query, update.Message.Chat.ID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.WithMessage(err, "Error getting user ID")
		}
	}
	return id, nil
}

func (conn DB) IsDriver(ctx context.Context, update tgbotapi.Update) (bool, error) {
	userId, err := conn.getUserID(update)
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
