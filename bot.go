package main

import (
	"net/url"
	"strconv"
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"encoding/gob"
)

type chat struct {
	Topic string
}

type user struct {
	LastChatInstance string
}

type QuizBot struct {
	*tgbotapi.BotAPI
	DB      *bolt.DB
	GameURL string
	Topics  TopicsMap

	chats map[string]chat
	users map[int]user
}

func readUsers(tx *bolt.Tx) (map[int]user, error) {
	users := make(map[int]user)
	b, err := tx.CreateBucketIfNotExists([]byte("users"))
	if err != nil {
		return nil, fmt.Errorf("error creating bucket: %v", err)
	}
	cur := b.Cursor()
	for k, v := cur.First(); k != nil; k, v = cur.Next() {
		var kInt int
		err := gob.NewDecoder(bytes.NewReader(k)).Decode(&kInt)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling key for users: %v", err)
		}

		var u user
		err = gob.NewDecoder(bytes.NewReader(v)).Decode(&u)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling user: %v", err)
		}
		users[kInt] = u
	}
	return users, nil
}

func readChats(tx *bolt.Tx) (map[string]chat, error) {
	chats := make(map[string]chat)
	b, err := tx.CreateBucketIfNotExists([]byte("chats"))
	if err != nil {
		return nil, fmt.Errorf("error creating bucket: %v", err)
	}
	cur := b.Cursor()
	for k, v := cur.First(); k != nil; k, v = cur.Next() {
		var ch chat
		err = gob.NewDecoder(bytes.NewReader(v)).Decode(&ch)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling chat: %v", err)
		}
		chats[string(k)] = ch
	}
	return chats, nil
}

func saveUser(db *bolt.DB, kInt int, u user) error {
	return db.Update(func(tx *bolt.Tx) error {
		var k bytes.Buffer
		err := gob.NewEncoder(&k).Encode(kInt)
		if err != nil {
			return fmt.Errorf("error encoding key: %v", err)
		}
		var v bytes.Buffer
		err = gob.NewEncoder(&v).Encode(u)
		if err != nil {
			return fmt.Errorf("error encoding value: %v", err)
		}
		err = tx.Bucket([]byte("users")).Put(k.Bytes(), v.Bytes())
		if err != nil {
			return fmt.Errorf("error putting user: %v", err)
		}
		return nil
	})
}

func saveChat(db *bolt.DB, id string, ch chat) error {
	return db.Update(func(tx *bolt.Tx) error {
		var v bytes.Buffer
		err := gob.NewEncoder(&v).Encode(ch)
		if err != nil {
			return fmt.Errorf("error encoding value: %v", err)
		}
		err = tx.Bucket([]byte("chats")).Put([]byte(id), v.Bytes())
		if err != nil {
			return fmt.Errorf("error putting user: %v", err)
		}
		return nil
	})
}

func NewQuizBot(botAPI *tgbotapi.BotAPI, db *bolt.DB, gameURL string, topics map[string]Topic) (*QuizBot, error) {
	var users map[int]user
	var chats map[string]chat
	err := db.Update(func(tx *bolt.Tx) error {
		var err error
		users, err = readUsers(tx)
		if err != nil {
			return fmt.Errorf("error reading users: %v", err)
		}

		chats, err = readChats(tx)
		if err != nil {
			return fmt.Errorf("error reading chats: %v", err)
		}
	})
	if err != nil {
		return nil, err
	}

	return &QuizBot{
		BotAPI:  botAPI,
		DB:      db,
		Topics:  topics,
		GameURL: gameURL,
		chats:   chats,
		users:   users,
	}, nil
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
				ch :=  chat{topic}
				bot.chats[user.LastChatInstance] = ch
				err := saveChat(bot.DB, user.LastChatInstance, ch)
				if err != nil {
					log.Printf("error saving chat: %v", err)
				}
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
	us := user{cq.ChatInstance}
	bot.users[cq.From.ID] = us
	err = saveUser(bot.DB, cq.From.ID, us)
	if err != nil {
		log.Printf("error saving user to bolt: %v", err)
	}

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
