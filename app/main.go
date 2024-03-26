package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"html/template"
	"log"
	"net/http"
	"ride-together-bot/conf"
	"ride-together-bot/db"
	"ride-together-bot/entitiy"
)

type BotAPI struct {
	API *tgbotapi.BotAPI
	db  *db.DB
}

func NewBot(api *tgbotapi.BotAPI, db *db.DB) BotAPI {
	api.Debug = true
	return BotAPI{
		API: api,
		db:  db,
	}
}

func (bot BotAPI) viewHandler(w http.ResponseWriter, r *http.Request) {
	// Получение данных из базы данных
	data, err := bot.db.GetAllData(context.Background())
	if err != nil {
		http.Error(w, errors.WithMessage(err, "ошибка получения данных из БД").Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Данные: ", data)
	// Отображение данных на веб-странице
	tmpl, err := template.ParseFiles("cite/index.html") // Путь к вашему HTML-шаблону
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (bot BotAPI) requestContact(chatID int64) {
	// Создаём сообщение
	requestContactMessage := tgbotapi.NewMessage(chatID, "Согласны ли вы предоставить ваш номер телефона для регистрации в системе?")

	// Создаём кнопку отправки контакта
	acceptButton := tgbotapi.NewKeyboardButtonContact("Да")
	declineButton := tgbotapi.NewKeyboardButton("Нет")

	// Создаём клавиатуру
	requestContactReplyKeyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{acceptButton, declineButton})
	requestContactMessage.ReplyMarkup = requestContactReplyKeyboard
	bot.API.Send(requestContactMessage) // Отправляем сообщение
}

// Проверка принятого контакта
func (bot BotAPI) checkRequestContactReply(ctx context.Context, update tgbotapi.Update, user entitiy.User) {
	if update.Message.Contact != nil { // Проверяем, содержит ли сообщение контакт
		if update.Message.Contact.UserID == update.Message.From.ID { // Проверяем действительно ли это контакт отправителя
			bot.db.Registration(ctx, user, update)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо!")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false) // Убираем клавиатуру
			bot.API.Send(msg)

		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номер телефона, который вы предоставили, принадлежит не вам!")
			bot.API.Send(msg)
			bot.requestContact(update.Message.Chat.ID)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Если вы не предоставите ваш номер телефона, вы не сможете пользоваться системой!")
		bot.API.Send(msg)
		bot.requestContact(update.Message.Chat.ID)
	}
}

// Тест клавиатура для тестовой веб страницы (auth)
func (bot BotAPI) inlineContact(chatID int64) {
	url := "https://danilcodegit.github.io/tg_bot_cite.github.io/"
	webappInfo := tgbotapi.WebAppInfo{URL: url}
	btn := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonWebApp("my btn", webappInfo),
	)

	reply := tgbotapi.NewInlineKeyboardMarkup(btn)
	msg := tgbotapi.NewMessage(chatID, "Кнопка")
	msg.ReplyMarkup = reply
	bot.API.Send(msg)

}

func (bot BotAPI) Updates(ctx context.Context, user entitiy.User) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.API.GetUpdatesChan(u)
	for update := range updates {
		chatID := update.Message.Chat.ID
		commands := update.Message.Command()
		msg := tgbotapi.NewMessage(chatID, update.Message.Text)
		switch commands {
		case "start":
			msg.Text = "Привет, я бот для поиска попутчиков в любой системе каршеринга.\nПриятной экономии"
			bot.API.Send(msg)

		case "registration":
			ok, err := bot.db.IsExists(ctx, update.Message.Chat.UserName)
			if err != nil {
				return errors.WithMessage(err, "is exists")
			}
			if ok {
				msg.Text = "Пользователь уже зарегистрирован"
				bot.API.Send(msg)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
				continue
			}
			msg.Text = "Регистрация пользователя"
			bot.API.Send(msg)
			bot.requestContact(chatID)
			for u := range updates {
				bot.checkRequestContactReply(ctx, u, user)
				break
			}

		case "auth":
			bot.inlineContact(chatID)
		}

	}
	return nil
}

func main() {
	ctx := context.Background()
	// Инициализация базы данных
	conn, err := db.NewDataBase(ctx, conf.DSN)
	if err != nil {
		log.Panic(errors.WithMessage(err, "ошибка инициализации БД"))
	}
	//err = conn.Conn.Ping()
	//if err != nil {
	//	return
	//}

	// Инициализация бота
	newBot, err := tgbotapi.NewBotAPI(conf.TelegramBotApiKey)
	if err != nil {
		log.Panic(err)
	}

	// Создание экземпляра BotAPI
	bot := NewBot(newBot, conn)

	// Создание экземпляра пользователя
	var user entitiy.User

	//// Регистрация обработчика HTTP
	//fs := http.FileServer(http.Dir("cite/assets"))
	//mux := http.NewServeMux()
	//mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	//mux.HandleFunc("/", bot.viewHandler)
	////if err := http.ListenAndServe("localhost:8080", mux); err != nil {
	////	log.Println(err)
	////}
	//if err := http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", mux); err != nil {
	//	log.Println(err)
	//}

	// Запуск обновлений
	err = bot.Updates(ctx, user)
	if err != nil {
		log.Panic(err)
	}

}
