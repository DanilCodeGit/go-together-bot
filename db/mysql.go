package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"

	"database/sql" // Import SQL package for MySQL
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"ride-together-bot/domain"
	"ride-together-bot/entitiy"
)

type DB struct {
	Conn *sql.DB // Change to MySQL database connection
}

func NewDataBase(ctx context.Context, dsn string) (*DB, error) {
	db, err := sql.Open("mysql", dsn) // Use MySQL driver and connection string
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v", err)
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, errors.WithMessage(err, "Error pinging database")
	}
	log.Println("Successful database connection")

	return &DB{Conn: db}, nil
}

func (conn DB) GetAllData(ctx context.Context) ([]entitiy.User, error) {
	query := `SELECT name, phone, login, chatid FROM users`
	data := make([]entitiy.User, 0)
	rows, err := conn.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.WithMessage(err, "Error executing query")
	}
	defer rows.Close()

	for rows.Next() {
		var user entitiy.User
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
func (conn DB) Registration(ctx context.Context, user entitiy.User, update tgbotapi.Update) error {
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
		// Error executing query
		return false, errors.WithMessage(err, "Error checking if user exists")
	}
	// Record found
	return true, nil
}

//func (conn DB) UpdateNumber(ctx context.Context, user entitiy.User) error {
//	query := `
//       INSERT INTO users (chatid, phone) VALUES (?, ?)
//       ON DUPLICATE KEY UPDATE phone = VALUES(phone)
//   `
//	_, err := conn.Conn.ExecContext(ctx, query, user.ChatID, user.Phone)
//	if err != nil {
//		return errors.WithMessage(err, "Error executing query")
//	}
//	return nil
//}
