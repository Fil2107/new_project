package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var tasks []string

func saveTasks() error {
    data, err := json.Marshal(tasks)
    if err != nil {
        return err
    }
    return os.WriteFile("tasks.json", data, 0644)
}

func loadTasks() error {
    data, err := os.ReadFile("tasks.json")
    if err != nil {
        if os.IsNotExist(err) {
            return nil
        }
        return err
    }
    return json.Unmarshal(data, &tasks)
}

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
        log.Panicf("Ошибка при инициализации бота: %v", err)
    }

    log.Printf("Бот авторизован как %s", bot.Self.UserName)

    // Загрузка задач из файла
    err = loadTasks()
    if err != nil {
        log.Fatalf("Ошибка при загрузке задач: %v", err)
    }

    // Настройка получения обновлений
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates := bot.GetUpdatesChan(u)

    // Обработка сообщений
    for update := range updates {
        if update.Message == nil {
            continue
        }

        log.Printf("Получено сообщение: %s", update.Message.Text)

        // Обработка только команд
        switch update.Message.Command() {
        case "add":
            args := update.Message.CommandArguments()
            if args == "" {
                sendMessage(bot, update.Message.Chat.ID, "Пожалуйста, укажите задачу.")
            } else {
                tasks = append(tasks, args)
                err := saveTasks()
                if err != nil {
                    sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("Не удалось сохранить задачу: %v", err))
                } else {
                    sendMessage(bot, update.Message.Chat.ID, "Задача добавлена.")
                }
            }
        case "remove":
            args := update.Message.CommandArguments()
            index, err := strconv.Atoi(args)
            if err != nil || index < 0 || index >= len(tasks) {
                sendMessage(bot, update.Message.Chat.ID, "Неверный индекс задачи.")
            } else {
                tasks = append(tasks[:index], tasks[index+1:]...)
                err := saveTasks()
                if err != nil {
                    sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("Не удалось удалить задачу: %v", err))
                } else {
                    sendMessage(bot, update.Message.Chat.ID, "Задача удалена.")
                }
            }
        case "list":
            if len(tasks) == 0 {
                sendMessage(bot, update.Message.Chat.ID, "Список задач пуст.")
            } else {
                var sb strings.Builder
                sb.WriteString("Список задач:\n")
                for i, task := range tasks {
                    sb.WriteString(fmt.Sprintf("%d. %s\n", i, task))
                }
                sendMessage(bot, update.Message.Chat.ID, sb.String())
            }
        default:
            // Игнорируем неизвестные команды
            log.Printf("Неизвестная команда: %s", update.Message.Command())
        }
    }
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
    msg := tgbotapi.NewMessage(chatID, text)
    _, err := bot.Send(msg)
    if err != nil {
        log.Printf("Ошибка при отправке сообщения: %v", err)
    }
}
