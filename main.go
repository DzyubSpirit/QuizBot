package main

import (
	"flag"
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"net/http"
	"github.com/boltdb/bolt"
)

type BotAPI interface {
	GetUpdatesChan(tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error)
	AnswerInlineQuery(config tgbotapi.InlineConfig) (tgbotapi.APIResponse, error)
	AnswerCallbackQuery(config tgbotapi.CallbackConfig) (tgbotapi.APIResponse, error)
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

func main() {
	botToken := flag.String("bot_token", "", "token of the Telegram bot")
	gameShortName := flag.String("game_short_name", "", "short name of telegram game")
	port := flag.String("port", "80", "port for listening score updates")
	gameURL := flag.String("game_url", GameURLProd, "url for game web page")
	flag.Parse()
	if *botToken == "" {
		log.Fatalf("should be bot_token arg for command")
	}
	if *gameShortName == "" {
		log.Fatalf("should be bot_token arg for command")
	} if *gameURL == "" {
		log.Fatalf("should be game_url arg for command")
	}
	log.Println(*gameURL)

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatalf("error openning database my.db: %v", err)
	}

	tgbot, err := tgbotapi.NewBotAPI(*botToken)
	tgbot.Debug = true
	var bot BotAPI
	bot = tgbot
	if err != nil {
		log.Fatalf("error creating new bot api: %v", err)
	}

	qbot, err := NewQuizBot(bot, db, *gameURL, *gameShortName, Topics)
	if err != nil {
		log.Fatalf("error creating quiz bot: %v", err)
	}

	log.Println("started bot")
	errs, err := qbot.ListenUpdates()
	if err != nil {
		log.Fatalf("error starting listening for telegram bot updates: %v", err)
	}
	go func() {
		for err := range errs {
			log.Printf("ERROR: listening updates: %v", err)
		}
	}()

	http.HandleFunc("/api/", qbot.HandleSetScore)
	http.HandleFunc("/api/topics/", Topics.TopicsHandler)
	http.Handle("/", http.FileServer(http.Dir("www")))
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}
