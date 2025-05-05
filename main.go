package main

import (
    "bytes"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MistralResponse struct {
    Choices []struct {
        Text string `json:"text"`
    } `json:"choices"`
}

func main() {
    // Чтение токена из файла
    tokenBytes, err := os.ReadFile("token.txt")
    if err != nil {
        log.Fatalf("Ошибка при чтении token.txt: %v", err)
    }
    botToken := strings.TrimSpace(string(tokenBytes))

    // Прямое задание переменных окружения
    mistralAPIURL := "https://api.mistral.ai/v1/endpoint" // Замените на правильный URL
    mistralAPIKey := "0tWX2Bm4xpx9a0N7Kw9HuqeWpjgkKwYk"   // Замените на правильный ключ

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

        userMessage := update.Message.Text

        // Подготовка запроса к API Mistral
        payload := map[string]string{
            "prompt": userMessage,
        }
        payloadBytes, err := json.Marshal(payload)
        if err != nil {
            log.Printf("Ошибка при подготовке запроса к API Mistral: %v", err)
            continue
        }

        req, err := http.NewRequest("POST", mistralAPIURL, bytes.NewBuffer(payloadBytes))
        if err != nil {
            log.Printf("Ошибка при создании запроса к API Mistral: %v", err)
            continue
        }
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", "Bearer "+mistralAPIKey)

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("Ошибка при отправке запроса к API Mistral: %v", err)
            continue
        }
        defer resp.Body.Close()

        var response MistralResponse
        err = json.NewDecoder(resp.Body).Decode(&response)
        if err != nil {
            log.Printf("Ошибка при расшифровке ответа от API Mistral: %v", err)
            continue
        }

        if len(response.Choices) == 0 {
            log.Println("Пустой ответ от API Mistral")
            continue
        }

        botMessage := tgbotapi.NewMessage(update.Message.Chat.ID, response.Choices[0].Text)
        bot.Send(botMessage)
    }
}
