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

const MaxErrCount = 100

type TopicName = string
type Chat struct {
	Topic   TopicName
	Ratings map[TopicName]Rating
}

type UserID = int
type Rating = map[UserID]int

type User struct {
	LastChatInstance    string
	LastInlineMessageID string
}

type DB interface {
	Update(fn func(*bolt.Tx) error) error
}

type QuizBot struct {
	BotAPI
	DB            DB //*bolt.DB
	GameURL       string
	GameShortName string
	Topics        TopicsMap

	Chats map[string]*Chat
	Users map[int]*User
}

func readUsers(tx *bolt.Tx) (map[int]*User, error) {
	users := make(map[int]*User)
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

		var u User
		err = gob.NewDecoder(bytes.NewReader(v)).Decode(&u)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling user: %v", err)
		}
		users[kInt] = &u
	}
	return users, nil
}

func readChats(tx *bolt.Tx) (map[string]*Chat, error) {
	chats := make(map[string]*Chat)
	b, err := tx.CreateBucketIfNotExists([]byte("chats"))
	if err != nil {
		return nil, fmt.Errorf("error creating bucket: %v", err)
	}
	cur := b.Cursor()
	for k, v := cur.First(); k != nil; k, v = cur.Next() {
		var ch Chat
		err = gob.NewDecoder(bytes.NewReader(v)).Decode(&ch)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling chat: %v", err)
		}
		chats[string(k)] = &ch
	}
	return chats, nil
}

func saveUser(db DB, kInt int, u User) error {
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

func saveChat(db DB, id string, ch Chat) error {
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

func NewQuizBot(botAPI BotAPI, db *bolt.DB, gameURL, gameShortName string, topics map[string]Topic) (*QuizBot, error) {
	var users map[int]*User
	var chats map[string]*Chat
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
		BotAPI:        botAPI,
		DB:            db,
		Topics:        topics,
		GameURL:       gameURL,
		GameShortName: gameShortName,
		Chats:         chats,
		Users:         users,
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

func (bot QuizBot) ListTopics(chatID int64) {
	var buffer bytes.Buffer
	buffer.WriteString("Topics:\n")
	for topic := range bot.Topics {
		buffer.WriteString(topic)
		buffer.WriteString("\n")
	}
	msg := tgbotapi.NewMessage(chatID, buffer.String())
	bot.Send(msg)
}

func TopicSelected(topic string) string {
	return fmt.Sprintf("Topic %q selected", topic)
}

func (bot QuizBot) SetTopic(chatID int64, fromID int, topic string) {
	if _, ok := bot.Topics[topic]; !ok {
		bot.Send(tgbotapi.NewMessage(chatID, "Unknown topic"))
		return
	}
	user, ok := bot.Users[fromID]
	if !ok {
		bot.Send(tgbotapi.NewMessage(chatID, "Play the game and then change topic"))
		return
	}

	msgStr := TopicSelected(topic)
	chat, ok := bot.Chats[user.LastChatInstance]
	if !ok {
		return
	}
	lastTopic := chat.Topic
	chat.Topic = topic
	err := saveChat(bot.DB, user.LastChatInstance, *bot.Chats[user.LastChatInstance])
	if err != nil {
		log.Printf("error saving chat: %v", err)
	}
	bot.Send(tgbotapi.NewMessage(chatID, msgStr))

	log.Printf("lastTopic rating: %v", chat.Ratings[lastTopic])
	log.Printf("topic rating: %v", chat.Ratings[topic])
	update := make(map[UserID]int, len(chat.Ratings[lastTopic]))
	for id := range chat.Ratings[lastTopic] {
		update[id] = 1
	}
	for id, score := range chat.Ratings[topic] {
		update[id] = score
	}
	gameChatID, err := strconv.ParseInt(user.LastChatInstance, 10, 64)
	if err != nil {
		log.Printf("ERROR: can't parse chat instance %q: %v", user.LastChatInstance, err)
		return
	}
	for id, score := range update {
		bot.Send(tgbotapi.SetGameScoreConfig{
			Score:           score,
			UserID:          id,
			ChatID:          int(gameChatID),
			InlineMessageID: user.LastInlineMessageID,
			Force:           true,
		})
	}
}

func (bot QuizBot) HandleCommand(id UserID, chatID int64, message *tgbotapi.Message) {
	command := message.Command()
	switch command {
	case "topics":
		bot.ListTopics(chatID)
	case "topic":
		topic := message.CommandArguments()
		bot.SetTopic(chatID, message.From.ID, topic)
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
	if topic, ok := bot.Chats[cq.ChatInstance]; ok {
		topicID = topic.Topic
	}
	us := User{cq.ChatInstance, cq.InlineMessageID}
	bot.Users[cq.From.ID] = &us
	err = saveUser(bot.DB, cq.From.ID, us)
	if err != nil {
		log.Printf("error saving user to bolt: %v", err)
	}

	q := u.Query()
	q.Add("userId", strconv.Itoa(cq.From.ID))
	q.Add("topicId", topicID)
	q.Add("chatId", cq.ChatInstance)
	q.Add("inlineId", cq.InlineMessageID)
	log.Printf("inlinceId: %v\n", cq.InlineMessageID)
	if cq.Message != nil {
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

func (bot QuizBot) SetScore(sr ScoreResult) (int, error) {
	if bot.Chats[sr.ChatID] == nil {
		bot.Chats[sr.ChatID] = &Chat{DefaultTopic, make(map[TopicName]Rating)}
		log.Printf("created: %v", bot.Chats[sr.ChatID])
	}
	if bot.Chats[sr.ChatID].Ratings == nil {
		bot.Chats[sr.ChatID].Ratings = make(map[TopicName]Rating)
	}
	chat := bot.Chats[sr.ChatID]
	if chat.Ratings[chat.Topic] == nil {
		chat.Ratings[chat.Topic] = make(Rating)
	}
	if sr.Score <= chat.Ratings[chat.Topic][sr.UserID] {
		return chat.Ratings[chat.Topic][sr.UserID], nil
	}

	chat.Ratings[chat.Topic][sr.UserID] = sr.Score
	log.Printf("rating for %q: %v", chat.Topic, chat.Ratings[chat.Topic])
	err := saveChat(bot.DB, sr.ChatID, *chat)
	if err != nil {
		return 0, fmt.Errorf("error saving chat: %v", err)
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
			return 0, fmt.Errorf("error parsing chat instance when saving score")
		}
		cfg.ChatID = int(chatID)
		cfg.MessageID = sr.MessageID
	}
	_, err = bot.Send(cfg)
	if err != nil {
		return 0, fmt.Errorf("error updating user score for scoreresult %v, err: %v", sr, err)
	}
	return sr.Score, nil
}

func (bot QuizBot) HandleSetScore(w http.ResponseWriter, r *http.Request) {
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

	score, err := bot.SetScore(sr)
	if err != nil {
		log.Printf("error setting score: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"newScore": score})
}

func (bot QuizBot) ListenUpdates() (<-chan error, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, fmt.Errorf("error getting update channel from bot: %v", err)
	}
	errs := make(chan error, MaxErrCount)

	go func() {
		for update := range updates {
			cq := update.CallbackQuery
			switch {
			case update.InlineQuery != nil:
				bot.HandleInlineQuery(update.InlineQuery)
			case update.Message != nil && update.Message.Command() != "":
				if update.Message.Chat == nil {
					errs <- fmt.Errorf("error handling command: chat is nil when message isn't")
					continue
				}
				if update.Message.From == nil {
					errs <- fmt.Errorf("error handling command; from is nil when message isn't")
					continue
				}
				bot.HandleCommand(update.Message.From.ID, update.Message.Chat.ID, update.Message)
			case cq != nil && cq.GameShortName == bot.GameShortName:
				bot.CallbackQuery(cq)
			default:
				bot.SendURL(update)
			}
		}
		close(errs)
	}()

	return errs, nil
}
