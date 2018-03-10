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
	"io/ioutil"
)

func callbackQuery(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	u, err := url.Parse("http://212.237.53.191:8526/game.html")
	if err != nil {
		log.Fatalf("error parsing url: %v", err)
	}

	q := u.Query()
	q.Add("userId", strconv.Itoa(cq.From.ID))
	q.Add("inlineId", cq.InlineMessageID)
	q.Add("topicId", "en")
	if cq.Message != nil {
		q.Add("chatId", cq.ChatInstance)
		q.Add("messageId", strconv.Itoa(cq.Message.MessageID))
	}
	u.RawQuery = q.Encode()
	bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
		CallbackQueryID: cq.ID,
		URL:             u.String(),
		Text:            "Hello, dear!",
		ShowAlert:       true,
	})
}

func sendURL(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil || update.Message.Chat == nil {
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Play the game: t.me/KICCBibleQuizBot?game=bible_quiz")
	bot.Send(msg)
}

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
			switch {
			case cq != nil && cq.GameShortName == *gameShortName:
				callbackQuery(bot, cq)
			default:
				sendURL(bot, update)
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
	http.HandleFunc("/api/topics/", TopicsHandler)
	http.Handle("/", http.FileServer(http.Dir("www")))
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}
