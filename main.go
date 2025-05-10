package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var tasks = make(map[int64][]string) // Хранилище задач (по chatID)

func main() {
	// Чтение токена из файла
	tokenBytes, err := os.ReadFile("token.txt")
	if err != nil {
		log.Fatalf("Ошибка при чтении token.txt: %v", err)
	}
	botToken := strings.TrimSpace(string(tokenBytes))

	// Инициализация бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Бот авторизован как %s", bot.Self.UserName)

	// Настройка получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Обработка сообщений
	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text
		var reply string

		switch {
		case text == "/start":
			reply = "Привет! Я бот для управления задачами. Используй команды:\n" +
				"/add <задача> - добавить задачу\n" +
				"/list - показать список задач\n" +
				"/del <номер> - удалить задачу\n" +
				"/help - показать справку"
		
		case text == "/help":
			reply = "Доступные команды:\n" +
				"/add <задача> - добавить задачу\n" +
				"/list - показать список задач\n" +
				"/del <номер> - удалить задачу\n" +
				"/help - показать справку"
		
		case strings.HasPrefix(text, "/add"):
			task := strings.TrimSpace(strings.TrimPrefix(text, "/add"))
			if task == "" {
				reply = "Пожалуйста, укажите задачу после команды /add"
			} else {
				tasks[chatID] = append(tasks[chatID], task)
				reply = fmt.Sprintf("Задача добавлена: %s", task)
			}
		
		case text == "/list":
			if len(tasks[chatID]) == 0 {
				reply = "Нет задач в списке"
			} else {
				reply = "Список задач:\n"
				for i, task := range tasks[chatID] {
					reply += fmt.Sprintf("%d. %s\n", i+1, task)
				}
			}
		
		case strings.HasPrefix(text, "/del"):
			arg := strings.TrimSpace(strings.TrimPrefix(text, "/del"))
			if arg == "" {
				reply = "Пожалуйста, укажите номер задачи для удаления"
			} else {
				num, err := strconv.Atoi(arg)
				if err != nil || num <= 0 {
					reply = "Некорректный номер задачи"
				} else if num > len(tasks[chatID]) {
					reply = "Задачи с таким номером не существует"
				} else {
					deleted := tasks[chatID][num-1]
					tasks[chatID] = append(tasks[chatID][:num-1], tasks[chatID][num:]...)
					reply = fmt.Sprintf("Задача удалена: %s", deleted)
				}
			}
		
		default:
			reply = "Неизвестная команда. Введите /help для списка команд"
		}

		msg := tgbotapi.NewMessage(chatID, reply)
		bot.Send(msg)
	}
}
