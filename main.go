package main

import (
	"flag"
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"net/http"
	"github.com/boltdb/bolt"
)

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
	}
	if *gameURL == "" {
		log.Fatalf("should be game_url arg for command")
	}
	log.Println(*gameURL)

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatalf("error openning database my.db: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(*botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	qbot, err := NewQuizBot(bot, db, *gameURL, Topics)
	if err != nil {
		log.Fatalf("error creating quiz bot: %v", err)
	}

	log.Println("started bot")
	go func() {
		for update := range updates {
			cq := update.CallbackQuery
			switch {
			case update.InlineQuery != nil:
				qbot.HandleInlineQuery(update.InlineQuery)
			case update.Message != nil && update.Message.Command() != "":
				if update.Message.Chat == nil {
					log.Printf("ERROR: chat is nil when message isn't")
					continue
				}
				if update.Message.From == nil {
					log.Printf("ERROR: from is nil when message isn't")
					continue
				}
				qbot.HandleCommand(update.Message.From.ID, update.Message.Chat.ID, update.Message)
			case cq != nil && cq.GameShortName == *gameShortName:
				qbot.CallbackQuery(cq)
			default:
				qbot.SendURL(update)
			}
		}
	}()

	http.HandleFunc("/api/", qbot.SetScore)
	http.HandleFunc("/api/topics/", Topics.TopicsHandler)
	http.Handle("/", http.FileServer(http.Dir("www")))
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}
