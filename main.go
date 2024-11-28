package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type Handler func(bot *tgbotapi.BotAPI, update tgbotapi.Update)

var (
	// ownerID — ID владельца бота.
	ownerID int64 = 1281548385
	// states — хранит текущее состояние для каждого пользователя.
	states = make(map[int64]string)
	// data — хранит данные пользователей. Внешний ключ — ID пользователя, внутренняя карта хранит пары ключ-значение данных.
	data = make(map[int64]map[string]string)
)

func main() {
	// Логирование в файл
	logFile, err := os.OpenFile("my_bot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Ошибка открытия лог-файла: %v", err)
	}
	defer logFile.Close()
	logrus.SetOutput(logFile)

	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	bot, err := tgbotapi.NewBotAPI("8088688317:AAE3gPdA8VO9ykyNvFe-TGx5rsAUpkW5cWU")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Авторизация аккаунта %s", bot.Self.UserName)
	// Список хендлеров
	handlers := map[string]Handler{
		"start": startHandler,
		"help":  helpHandler,
		"nuber": nuberHandler,
		"sites": sitesHandler,
		"photo": photoHandler,
		"call":  callHandler,
		"late":  lateHandler,
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			// Проверяем состояние пользователя
			if state, ok := states[update.Message.Chat.ID]; ok {
				processState(bot, update, state)
				continue
			}

			if update.Message.IsCommand() {
				command := update.Message.Command()

				if handler, exists := handlers[command]; exists {
					start := time.Now()
					logrus.WithFields(logrus.Fields{
						"command":   command,
						"chat_id":   update.Message.Chat.ID,
						"user_id":   update.Message.From.ID,
						"timestamp": start,
					}).Info("Начало обработки команды")

					handler(bot, update)

					logrus.WithFields(logrus.Fields{
						"command":   command,
						"chat_id":   update.Message.Chat.ID,
						"user_id":   update.Message.From.ID,
						"timestamp": time.Since(start),
					}).Info("Команда обработана")

					// НАЧАЛО ЗАПИСИ КОМАНД В БД
					conn, err := pgx.Connect(context.Background(), "postgres://postgres:5280065@localhost:5432/postgres")
					if err != nil {
						fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
						os.Exit(1)
					}
					defer conn.Close(context.Background())

					conn.QueryRow(context.Background(),
						"INSERT INTO logi(user_id, chat_id, time, comand) "+
							"VALUES ("+strconv.FormatInt(update.Message.From.ID, 10)+", "+strconv.FormatInt(update.Message.Chat.ID, 10)+", NOW(), '"+command+"');")

					// КОНЕЦ ЗАПИСИ КОМАНД В БД
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не знаю эту команду.")
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Ошибка при отправке сообщения: %v", err)
					}
				}
			}
		}
	}
}

// Команды обработчиков

func startHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я ваш помощник в учёбе. Напишите /help, чтобы узнать, что я могу.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func helpHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я могу помочь вам с помощью следующих команд:\n/start - Запуск бота\n/help - Вывести сопроводительное сообщение\n/photo - Фоточка расписания\n/call - Фоточка звонков\n/nuber - Номера\n/sites - Площадки\n/late - Сообщить об опоздании")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func nuberHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Номера:\n+7 903 185-96-82 Анна Сергеевна администрация:\n+7 926 124-07-19 Вероника Сергеевна соцпедагог(по всем социальным вопросам):\n+7 977 268-33-61 Артемий Александрович куратор")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}

}

func sitesHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Площадки:\n город Москва, улица Большие Каменщики, дом 7 «Таганское - 1» Центр Телекоммуникации\n город Москва, улица Речников, дом 28, корпус 1, строение 2 «Коломенское - 2» Центр Общеобразовательной подготовки\n город Москва, улица Судакова, дом 18А «Люблинское - 3» Центр Дополнительного образования\n город Москва, улица Корнейчука, дом 55А «Бибиревское - 4» Центр Дополнительного образования «Юный Автомобилист»\n город Москва, Кирпичная улица, дом 33 «Семеновское - 5» Центр Информационной безопасности\n город Москва, Рязанский проспект, дом 8, строение 1 «Рязанское - 6» Центр Автоматизации и ИТ\n город Москва, Рабочая улица, дом 12, строение 1 «Римское - 7» Центр Радиоэлектроники\n город Москва, Басовская улица, дом 12 «Авиамоторное - 8» Центр Электроснабжения и Автотранспорта")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

// Отправка фото в формате png так как это единственный формат который поддерживается
func photoHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FilePath("C:/Users/хрю/Desktop/Новая папка (2)/1.png"))
	if _, err := bot.Send(photo); err != nil {
		log.Printf("Ошибка при отправке изображения: %v", err)
		msg :=
			tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось отправить изображение.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Картинка успешно отправлена!")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения: %v", err)
		}
	}
}
func callHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FilePath("C:/Users/хрю/Desktop/Новая папка (2)/2.png"))
	if _, err := bot.Send(photo); err != nil {
		log.Printf("Ошибка при отправке изображения: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось отправить изображение.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Картинка успешно отправлена!")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения: %v", err)
		}
	}
}

// Хендлер для опозданий
func lateHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ваше ФИО:")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
	states[update.Message.Chat.ID] = "awaiting_name"
	data[update.Message.Chat.ID] = make(map[string]string)
}

// Обработка состояний (шагов)
func processState(bot *tgbotapi.BotAPI, update tgbotapi.Update, state string) {
	switch state {
	case "awaiting_name":
		data[update.Message.Chat.ID]["name"] = update.Message.Text
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На сколько минут вы опоздаете?")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения: %v", err)
		}
		states[update.Message.Chat.ID] = "awaiting_time"
	case "awaiting_time":
		data[update.Message.Chat.ID]["time"] = update.Message.Text
		name := data[update.Message.Chat.ID]["name"]
		time := data[update.Message.Chat.ID]["time"]

		// Отправляем сообщение владельцу
		msg := tgbotapi.NewMessage(ownerID, "Студент: "+name+"\nОпоздает на: "+time+" минут.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения: %v", err)
		}

		// Подтверждаем пользователю
		confirmation := tgbotapi.NewMessage(update.Message.Chat.ID, "Спасибо, ваше сообщение отправлено.")
		if _, err := bot.Send(confirmation); err != nil {
			log.Printf("Ошибка при отправке сообщения: %v", err)
		}

		// Очищаем состояние
		delete(states, update.Message.Chat.ID)
		delete(data, update.Message.Chat.ID)
	}
}
