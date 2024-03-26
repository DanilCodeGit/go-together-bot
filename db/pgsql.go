package db

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"log"
	"os"
	"ride-together-bot/domain"
	"ride-together-bot/entitiy"
)

type DB struct {
	Conn *pgxpool.Pool
}

func NewDataBase(ctx context.Context, dsn string) (*DB, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v", err)
		return nil, err
	}
	log.Println("Успешное подключение")

	return &DB{Conn: conn}, nil
}

func (conn DB) GetAllData(ctx context.Context) ([]entitiy.User, error) {
	query := `select name, phone, login, password, chatid from users`
	data := make([]entitiy.User, 0)
	res, err := conn.Conn.Query(ctx, query)
	if err != nil {
		return nil, errors.WithMessage(err, "ошибка выполнения запроса")
	}
	defer res.Close()

	for res.Next() {
		var user entitiy.User
		err := res.Scan(&user.Name, &user.Phone, &user.Login, &user.Password, &user.ChatID)
		if err != nil {
			return nil, errors.WithMessage(err, "ошибка сканирования данных")
		}
		data = append(data, user)
	}

	if err := res.Err(); err != nil {
		return nil, errors.WithMessage(err, "ошибка получения данных")
	}

	return data, nil
}

func (conn DB) Registration(ctx context.Context, user entitiy.User, update tgbotapi.Update) error {
	query := `insert into users (name, phone, login, password, chatid) VALUES ($1,$2,$3,$4, $5)`

	_, err := conn.Conn.Exec(ctx, query, update.Message.Contact.FirstName, update.Message.Contact.PhoneNumber, update.Message.From.UserName, user.Password, update.Message.Chat.ID)
	if err != nil {
		return errors.WithMessage(err, "registration error")
	}

	return nil
}

func (conn DB) IsExists(ctx context.Context, login string) (bool, error) {
	query := `SELECT phone FROM users WHERE login = $1`
	var res domain.User
	row := conn.Conn.QueryRow(ctx, query, login)
	err := row.Scan(&res.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Запись не найдена
			return false, nil
		}
		// Произошла ошибка при выполнении запроса
		return false, errors.WithMessage(err, "Error checking if user exists")
	}
	// Запись найдена
	return true, nil
}

func (conn DB) UpdateNumber(ctx context.Context, user entitiy.User) error {
	query := `
       INSERT INTO users (chatid, phone) VALUES ($1, $2)
       ON CONFLICT (chatid)
       DO UPDATE SET phone = $2
   `

	_, err := conn.Conn.Exec(ctx, query, user.ChatID, user.Phone)
	if err != nil {
		return errors.WithMessage(err, "exec query")
	}
	return nil
}

//package db
//
//import (
//	"context"
//	"fmt"
//	"log"
//	"os"
//
//	"database/sql" // Import SQL package for MySQL
//	_ "github.com/go-sql-driver/mysql"
//	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
//	"github.com/pkg/errors"
//	"ride-together-bot/domain"
//	"ride-together-bot/entitiy"
//)
//
//type DB struct {
//	Conn *sql.DB // Change to MySQL database connection
//}
//
//func NewDataBase(ctx context.Context, dsn string) (*DB, error) {
//	db, err := sql.Open("mysql", dsn) // Use MySQL driver and connection string
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v", err)
//		return nil, err
//	}
//	if err := db.Ping(); err != nil {
//		return nil, errors.WithMessage(err, "Error pinging database")
//	}
//	log.Println("Successful database connection")
//
//	return &DB{Conn: db}, nil
//}
//
//func (conn DB) GetAllData(ctx context.Context) ([]entitiy.User, error) {
//	query := `SELECT name, phone, login, password, chatid FROM users`
//	data := make([]entitiy.User, 0)
//	rows, err := conn.Conn.QueryContext(ctx, query)
//	if err != nil {
//		return nil, errors.WithMessage(err, "Error executing query")
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//		var user entitiy.User
//		err := rows.Scan(&user.Name, &user.Phone, &user.Login, &user.Password, &user.ChatID)
//		if err != nil {
//			return nil, errors.WithMessage(err, "Error scanning data")
//		}
//		data = append(data, user)
//	}
//
//	if err := rows.Err(); err != nil {
//		return nil, errors.WithMessage(err, "Error retrieving data")
//	}
//
//	return data, nil
//}
//
//func (conn DB) Registration(ctx context.Context, user entitiy.User, update tgbotapi.Update) error {
//	query := `INSERT INTO users (name, phone, login, password, chatid) VALUES (?, ?, ?, ?, ?)`
//	_, err := conn.Conn.ExecContext(ctx, query, update.Message.Contact.FirstName, update.Message.Contact.PhoneNumber, update.Message.From.UserName, user.Password, update.Message.Chat.ID)
//	if err != nil {
//		return errors.WithMessage(err, "Registration error")
//	}
//	return nil
//}
//
//func (conn DB) IsExists(ctx context.Context, login string) (bool, error) {
//	query := `SELECT phone FROM users WHERE login = ?`
//	var res domain.User
//	err := conn.Conn.QueryRowContext(ctx, query, login).Scan(&res.Login)
//	if err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			// Record not found
//			return false, nil
//		}
//		// Error executing query
//		return false, errors.WithMessage(err, "Error checking if user exists")
//	}
//	// Record found
//	return true, nil
//}
//
//func (conn DB) UpdateNumber(ctx context.Context, user entitiy.User) error {
//	query := `
//        INSERT INTO users (chatid, phone) VALUES (?, ?)
//        ON DUPLICATE KEY UPDATE phone = VALUES(phone)
//    `
//	_, err := conn.Conn.ExecContext(ctx, query, user.ChatID, user.Phone)
//	if err != nil {
//		return errors.WithMessage(err, "Error executing query")
//	}
//	return nil
//}
