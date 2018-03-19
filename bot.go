package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/boltdb/bolt"
	"gopkg.in/telegram-bot-api.v4"
)

type topic = string
type chat struct {
	Topic   topic
	Ratings map[topic]rating
}

type userID = int
type rating map[userID]int

type user struct {
	LastChatInstance string
}

type QuizBot struct {
	*tgbotapi.BotAPI
	DB      *bolt.DB
	GameURL string
	Topics  TopicsMap

	chats map[string]*chat
	users map[int]*user
}

func readUsers(tx *bolt.Tx) (map[int]*user, error) {
	users := make(map[int]*user)
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
		users[kInt] = &u
	}
	return users, nil
}

func readChats(tx *bolt.Tx) (map[string]*chat, error) {
	chats := make(map[string]*chat)
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
		chats[string(k)] = &ch
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
	var users map[int]*user
	var chats map[string]*chat
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
		return nil
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

func (bot QuizBot) HandleCommand(id userID, chatID int64, message *tgbotapi.Message) {
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
		if _, ok := bot.Topics[topic]; !ok {
			bot.Send(tgbotapi.NewMessage(chatID, "Unknown topic"))
			return
		}
		user, ok := bot.users[message.From.ID]
		if !ok {
			bot.Send(tgbotapi.NewMessage(chatID, "Play the game and then change topic"))
			return
		}

		msgStr := fmt.Sprintf("Topic %q selected", topic)
		lastTopic := bot.chats[user.LastChatInstance].Topic
		chat := bot.chats[user.LastChatInstance]
		chat.Topic = topic
		err := saveChat(bot.DB, user.LastChatInstance, *bot.chats[user.LastChatInstance])
		if err != nil {
			log.Printf("error saving chat: %v", err)
		}
		bot.Send(tgbotapi.NewMessage(chatID, msgStr))

		update := make(map[userID]int, len(chat.Ratings[lastTopic]))
		for id := range chat.Ratings[lastTopic] {
			update[id] = 0
		}
		for id, score := range chat.Ratings[topic] {
			update[id] = score
		}
		chatID, err := strconv.ParseInt(user.LastChatInstance, 10, 64)
		if err != nil {
			log.Printf("ERROR: can't parse chat instance %q: %v", user.LastChatInstance, err)
			return
		}
		for id, score := range update {
			bot.Send(tgbotapi.SetGameScoreConfig{
				Score:  score,
				UserID: id,
				ChatID: int(chatID),
				Force:  true,
			})
		}
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

	topicID := DefaultTopic
	if topic, ok := bot.chats[cq.ChatInstance]; ok {
		topicID = topic.Topic
	}
	us := user{cq.ChatInstance}
	bot.users[cq.From.ID] = &us
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

type ScoreResult struct {
	UserID    int
	InlineID  string
	ChatID    string
	MessageID int
	Score     int
}

func (bot QuizBot) SetScore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	byt, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error reading request body: %v", err)
		return
	}
	log.Printf("bytes: %s", byt)
	var sr ScoreResult
	err = json.Unmarshal(byt, &sr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(w, "error parsing score result: %v", err)
		return
	}

	chat := bot.chats[sr.ChatID]
	if sr.Score <= chat.Ratings[chat.Topic][sr.UserID] {
		fmt.Fprint(w, "Okay")
		return
	}

	chat.Ratings[chat.Topic][sr.UserID] = sr.Score
	err = saveChat(bot.DB, sr.ChatID, *chat)
	if err != nil {
		log.Printf("ERROR: can't save chat: %v", err)
	}

	cfg := tgbotapi.SetGameScoreConfig{
		Score:  sr.Score,
		UserID: sr.UserID,
	}
	if sr.InlineID != "" {
		cfg.InlineMessageID = sr.InlineID
	}
	if sr.ChatID != "" {
		chatID, err := strconv.ParseInt(sr.ChatID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("ERROR: can't parse chat instance when saving score")
			return
		}
		cfg.ChatID = int(chatID)
		cfg.MessageID = sr.MessageID
	}
	_, err = bot.Send(cfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error updating user score for scoreresult %v, err: %v", sr, err)
		return
	}
	fmt.Fprint(w, "Okay")
}
