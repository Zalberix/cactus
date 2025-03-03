package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	telebot "gopkg.in/telebot.v4"
)

type MessageRequest struct {
	Message string `json:"message"`
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		fmt.Println("Необходимо установить переменную окружения TELEGRAM_BOT_TOKEN")
		return
	}

	// Подключение к базе данных SQLite
	db, err := sql.Open("sqlite3", "./cmd/telegram-bot/subscriptions.db")
	if err != nil {
		fmt.Println("Ошибка подключения к базе данных:", err)
		return
	}
	defer db.Close()

	// Создание таблицы для хранения подписчиков
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS subscribers (chat_id INTEGER PRIMARY KEY)`)
	if err != nil {
		fmt.Println("Ошибка создания таблицы:", err)
		return
	}

	p := telebot.Settings{
		Token: botToken,
	}

	bot, err := telebot.NewBot(p)
	if err != nil {
		fmt.Println("Ошибка создания бота:", err)
		return
	}

	bot.Handle("/start", func(c telebot.Context) error {
		keyboard := &telebot.ReplyMarkup{ResizeKeyboard: true}
		subscribeBtn := keyboard.Text("Подписаться")
		keyboard.Reply(keyboard.Row(subscribeBtn))

		return c.Send("Добро пожаловать! Нажмите кнопку, чтобы подписаться.", keyboard)
	})

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		if c.Text() == "Подписаться" {
			chatID := c.Chat().ID

			// Сохранение ID чата в базе данных
			_, err := db.Exec("INSERT OR IGNORE INTO subscribers (chat_id) VALUES (?)", chatID)
			if err != nil {
				fmt.Println("Ошибка сохранения подписчика:", err)
				return c.Send("Произошла ошибка при подписке.")
			}

			return c.Send("Вы успешно подписались!")
		}
		return nil
	})

	http.HandleFunc("POST /send", func(w http.ResponseWriter, r *http.Request) {
		var request MessageRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Ошибка парсинга запроса", http.StatusBadRequest)
			return
		}

		if request.Message == "" {
			http.Error(w, "Сообщение не должно быть пустым", http.StatusBadRequest)
			return
		}

		// Получение всех подписчиков из базы данных
		rows, err := db.Query("SELECT chat_id FROM subscribers")
		if err != nil {
			fmt.Println("Ошибка получения подписчиков:", err)
			http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var chatID int64
			if err := rows.Scan(&chatID); err != nil {
				fmt.Println("Ошибка чтения данных:", err)
				continue
			}

			recipient := &telebot.Chat{ID: chatID}
			if _, err := bot.Send(recipient, request.Message); err != nil {
				fmt.Println("Ошибка отправки сообщения:", err)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Сообщение отправлено подписчикам"))
	})

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Сервис запущен на порту", port)
	go bot.Start()
	http.ListenAndServe(":"+port, nil)
}
