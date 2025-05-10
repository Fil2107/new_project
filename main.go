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

        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

        switch update.Message.Command() {
        case "add":
            args := update.Message.CommandArguments()
            if args == "" {
                msg.Text = "Пожалуйста, укажите задачу."
            } else {
                tasks = append(tasks, args)
                err := saveTasks()
                if err != nil {
                    msg.Text = fmt.Sprintf("Не удалось сохранить задачу: %v", err)
                } else {
                    msg.Text = "Задача добавлена."
                }
            }
        case "remove":
            args := update.Message.CommandArguments()
            index, err := strconv.Atoi(args)
            if err != nil || index < 0 || index >= len(tasks) {
                msg.Text = "Неверный индекс задачи."
            } else {
                tasks = append(tasks[:index], tasks[index+1:]...)
                err := saveTasks()
                if err != nil {
                    msg.Text = fmt.Sprintf("Не удалось удалить задачу: %v", err)
                } else {
                    msg.Text = "Задача удалена."
                }
            }
        case "list":
            if len(tasks) == 0 {
                msg.Text = "Список задач пуст."
            } else {
                var sb strings.Builder
                sb.WriteString("Список задач:\n")
                for i, task := range tasks {
                    sb.WriteString(fmt.Sprintf("%d. %s\n", i, task))
                }
                msg.Text = sb.String()
            }
        default:
            msg.Text = "Неизвестная команда. Используйте /add, /remove или /list."
        }

        _, err = bot.Send(msg)
        if err != nil {
            log.Printf("Ошибка при отправке сообщения: %v", err)
        }
    }
}
