package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	telebot "gopkg.in/telebot.v4"
)

type MessageRequest struct {
	Message string `json:"message"`
}

func MustLoadDB() (*sql.DB, error) {
	// Подключение к базе данных SQLite
	db, err := sql.Open("sqlite3", "./cmd/telegram-bot/subscriptions.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS subscribers (chat_id INTEGER PRIMARY KEY)`)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания таблицы: %w", err)
	}

	return db, nil
}

func OnTextHandle(db *sql.DB) telebot.HandlerFunc {
	return func(c telebot.Context) error {
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
	}
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		fmt.Println("Необходимо установить переменную окружения TELEGRAM_BOT_TOKEN")
		return
	}

	db, err := MustLoadDB()
	if err != nil {
		fmt.Println("Ошибка подключения к базе данных:", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic("ошибка закрытия базы")
		}
	}(db)

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

	bot.Handle(telebot.OnText, OnTextHandle(db))

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
		if rows.Err() != nil {
			fmt.Println("Ошибка чтения подписчиков:", err)
			http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
			return
		}
		defer func() {
			err := rows.Close()
			if err != nil {
				fmt.Println("ошибка запроса к базе данных базы")
			}
		}()

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
	srv := http.Server{
		Addr:         ":" + port,
		Handler:      nil,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  240 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Завершение с ошибкой %s: %v\n", port, err.Error())
		return
	}
}
