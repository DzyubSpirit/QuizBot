package main

import (
	"net/url"
	"strconv"
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"bytes"
	"fmt"
)

type chat struct {
	Topic string
}

type user struct {
	LastChatInstance string
}

type QuizBot struct {
	*tgbotapi.BotAPI
	GameURL string
	Topics  TopicsMap

	chats map[string]chat
	users map[int]user
}

func NewQuizBot(botAPI *tgbotapi.BotAPI, gameURL string, topics map[string]Topic) QuizBot {
	return QuizBot{
		BotAPI:  botAPI,
		Topics:  topics,
		GameURL: gameURL,
		chats:   make(map[string]chat),
		users:   make(map[int]user),
	}
}

func (bot QuizBot) HandleInlineQuery(query *tgbotapi.InlineQuery) {
	u := "https://google.com/"
	bot.AnswerInlineQuery(tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results: []interface{}{
			tgbotapi.InlineQueryResultGame{
				Type:          "game",
				ID:            "0",
				GameShortName: "test_game",
				ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{{
						{URL: &u},
					}},
				},
			},
		},
	})
}

func (bot QuizBot) HandleCommand(userID int, chatID int64, message *tgbotapi.Message) {
	command := message.Command()
	switch command {
	case "topics":
		var buffer bytes.Buffer
		buffer.WriteString("Topics:\n")
		for topic := range bot.Topics {
			buffer.WriteString(topic)
			buffer.WriteString("\n")
		}
		msg := tgbotapi.NewMessage(chatID, buffer.String())
		bot.Send(msg)
	case "topic":
		topic := message.CommandArguments()
		var msgStr string
		if _, ok := bot.Topics[topic]; !ok {
			msgStr = "Unknown topic"
		} else {
			if user, ok := bot.users[message.From.ID]; ok {
				msgStr = fmt.Sprintf("Topic %q selected", topic)
				bot.chats[user.LastChatInstance] = chat{topic}
			} else {
				msgStr = fmt.Sprintf("Play the game and then change topic")
			}
		}
		bot.Send(tgbotapi.NewMessage(chatID, msgStr))
	default:
		msg := tgbotapi.NewMessage(chatID, "Unknown command")
		bot.Send(msg)
	}
}

func (bot QuizBot) CallbackQuery(cq *tgbotapi.CallbackQuery) {
	u, err := url.Parse(bot.GameURL)
	if err != nil {
		log.Fatalf("error parsing url: %v", err)
	}

	topicID := "en"
	if topic, ok := bot.chats[cq.ChatInstance]; ok {
		topicID = topic.Topic
	}
	bot.users[cq.From.ID] = user{cq.ChatInstance}

	q := u.Query()
	q.Add("userId", strconv.Itoa(cq.From.ID))
	q.Add("topicId", topicID)
	q.Add("inlineId", cq.InlineMessageID)
	log.Printf("inlinceId: %v\n", cq.InlineMessageID)
	if cq.Message != nil {
		q.Add("chatId", cq.ChatInstance)
		q.Add("messageId", strconv.Itoa(cq.Message.MessageID))
		log.Printf("chatId: %v", cq.ChatInstance)
		log.Printf("messageId: %v", cq.Message.MessageID)
	}
	u.RawQuery = q.Encode()
	bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
		CallbackQueryID: cq.ID,
		URL:             u.String(),
		Text:            "Hello, dear!",
		ShowAlert:       true,
	})
}

func (bot QuizBot) SendURL(update tgbotapi.Update) {
	if update.Message == nil || update.Message.Chat == nil {
		return
	}
	msg := tgbotapi.GameConfig{
		GameShortName: "test_game",
		BaseChat: tgbotapi.BaseChat{
			ChatID: update.Message.Chat.ID,
		},
	}
	//msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Play the game: t.me/KICCBibleQuizBot?game=bible_quiz")
	bot.Send(msg)
}
