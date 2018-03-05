package main

import (
	"flag"
	"log"
	"gopkg.in/telegram-bot-api.v4"
)

//421904983:AAFYGcsueGwqVX1gYSW9YcoNXntt-xxaWqo
func main() {
	botToken := flag.String("bot_token", "", "token of the Telegram bot")
	flag.Parse()
	if *botToken == "" {
		log.Fatalf("should be bot_token arg for command")
	}

	bot, err := tgbotapi.NewBotAPI(*botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	log.Println("started bot")
	for update := range updates {
		cq := update.CallbackQuery
		if cq == nil {
			continue
		}
		if cq.GameShortName != "test_game" {
			continue
		}
		bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
			CallbackQueryID: cq.ID,
			URL:             "https://dzyubspirit.github.io/QuizBot/game.html",
			Text:            "Hello, dear!",
			ShowAlert:       true,
		})

		/*
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		*/
	}
}
