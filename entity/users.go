package entity

type User struct {
	Name   string `json:"name"`
	Login  string `json:"login"`
	Phone  string `json:"phone"`
	ChatID int64  `json:"chatId"`
}
