package entitiy

type User struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	ChatID   int64  `json:"chatID"`
}
