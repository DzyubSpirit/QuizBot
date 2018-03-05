package main

import (
	"flag"
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"net/url"
	"strconv"
	"net/http"
	"fmt"
	"encoding/json"
)

func main() {
	botToken := flag.String("bot_token", "", "token of the Telegram bot")
	gameShortName := flag.String("game_short_name", "", "short name of telegram game")
	port := flag.String("port", "80", "port for listening score updates")
	flag.Parse()
	if *botToken == "" {
		log.Fatalf("should be bot_token arg for command")
	}
	if *gameShortName == "" {
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
	go func() {
		for update := range updates {
			cq := update.CallbackQuery
			if cq == nil {
				continue
			}
			if cq.GameShortName != *gameShortName {
				continue
			}
			u, err := url.Parse("https://dzyubspirit.github.io/QuizBot/game.html")
			if err != nil {
				log.Printf("error parsing url: %v", err)
				continue
			}
			q := u.Query()
			q.Add("userId", strconv.Itoa(cq.From.ID))
			q.Add("inlineId", cq.InlineMessageID)
			q.Add("chatId", cq.ChatInstance)
			q.Add("messageId", strconv.Itoa(cq.Message.MessageID))
			u.RawQuery = q.Encode()
			bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
				CallbackQueryID: cq.ID,
				URL:             u.String(),
				Text:            "Hello, dear!",
				ShowAlert:       true,
			})
		}
	}()

	type ScoreResult struct {
		UserID    int
		InlineID  string
		ChatID    int
		MessageID int
		Score     int
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var sr ScoreResult
		err := json.NewDecoder(r.Body).Decode(&sr)
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
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
