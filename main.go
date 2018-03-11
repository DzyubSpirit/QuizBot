package main

import (
	"flag"
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"net/http"
	"fmt"
	"encoding/json"
	"io/ioutil"
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

	bot, err := tgbotapi.NewBotAPI(*botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	log.Println("started bot")
	go func() {
		qbot := NewQuizBot(bot, *gameURL, Topics)
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

	type ScoreResult struct {
		UserID    int
		InlineID  string
		ChatID    int
		MessageID int
		Score     int
	}
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error reading request body: %v", err)
			return
		}
		log.Printf("bytes: %s", bytes)
		var sr ScoreResult
		err = json.Unmarshal(bytes, &sr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(w, "error parsing score result: %v", err)
			return
		}

		cfg := tgbotapi.SetGameScoreConfig{
			Score:  sr.Score,
			UserID: sr.UserID,
		}
		if sr.InlineID != "" {
			cfg.InlineMessageID = sr.InlineID
		}
		if sr.ChatID != 0 {
			cfg.ChatID = sr.ChatID
			cfg.MessageID = sr.MessageID
		}
		_, err = bot.Send(cfg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error updating user score for scoreresult %v, err: %v", sr, err)
			return
		}
		fmt.Fprint(w, "Okay")
	})
	http.HandleFunc("/api/topics/", Topics.TopicsHandler)
	http.Handle("/", http.FileServer(http.Dir("www")))
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}
