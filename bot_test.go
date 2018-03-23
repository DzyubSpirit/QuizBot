package main

import (
	"testing"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/boltdb/bolt"
	"strconv"
	"log"
)

type FakeBotAPI struct {
	Updates []tgbotapi.Update
	Msg     []tgbotapi.Chattable
}

func (bot FakeBotAPI) GetUpdatesChan(tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error) {
	ch := make(chan tgbotapi.Update)
	go func() {
		for _, u := range bot.Updates {
			ch <- u
		}
		close(ch)
	}()
	return ch, nil
}

func (FakeBotAPI) AnswerInlineQuery(config tgbotapi.InlineConfig) (tgbotapi.APIResponse, error) {
	return tgbotapi.APIResponse{}, nil
}

func (FakeBotAPI) AnswerCallbackQuery(config tgbotapi.CallbackConfig) (tgbotapi.APIResponse, error) {
	return tgbotapi.APIResponse{}, nil
}

func (bot *FakeBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	bot.Msg = append(bot.Msg, c)
	return tgbotapi.Message{}, nil
}

type FakeDB struct{}

func (FakeDB) Update(fn func(*bolt.Tx) error) error {
	return nil
}

func TestChangingTopic(t *testing.T) {
	var chatID int64 = 1
	chatInst := 2
	userID := 1
	topic := "books_en"
	topic2 := "books_ru"
	messageID := 1
	newScore := 5
	newScore2 := 10
	newScore3 := 15

	expected := []tgbotapi.Chattable{
		tgbotapi.SetGameScoreConfig{
			MessageID: messageID,
			UserID:    userID,
			ChatID:    chatInst,
			Score:     newScore,
		},
		tgbotapi.NewMessage(chatID, TopicSelected(topic2)),
		tgbotapi.SetGameScoreConfig{
			MessageID: messageID,
			UserID:    userID,
			ChatID:    chatInst,
			Score:     0,
			Force:     true,
		},
		tgbotapi.SetGameScoreConfig{
			MessageID: messageID,
			UserID:    userID,
			ChatID:    chatInst,
			Score:     newScore2,
		},
		tgbotapi.NewMessage(chatID, TopicSelected(topic)),
		tgbotapi.SetGameScoreConfig{
			MessageID: messageID,
			UserID:    userID,
			ChatID:    chatInst,
			Score:     newScore,
			Force:     true,
		},
		tgbotapi.SetGameScoreConfig{
			MessageID: messageID,
			UserID:    userID,
			ChatID:    chatInst,
			Score:     newScore3,
		},
	}

	fakeBot := &FakeBotAPI{nil, nil}
	bot := QuizBot{
		BotAPI: fakeBot,
		Topics: Topics,
		DB:     FakeDB{},
		Chats:  make(map[string]*Chat),
		Users: map[int]*User{
			userID: {strconv.Itoa(chatInst) ,messageID},
		},
	}
	bot.SetScore(ScoreResult{
		Score:     newScore,
		ChatID:    strconv.Itoa(chatInst),
		UserID:    userID,
		MessageID: messageID,
		InlineID:  "",
	})
	bot.SetTopic(chatID, userID, topic2)
	bot.SetScore(ScoreResult{
		Score:     newScore2,
		ChatID:    strconv.Itoa(chatInst),
		UserID:    userID,
		MessageID: messageID,
		InlineID:  "",
	})
	bot.SetTopic(chatID, userID, topic)
	bot.SetScore(ScoreResult{
		Score:     newScore3,
		ChatID:    strconv.Itoa(chatInst),
		UserID:    userID,
		MessageID: messageID,
		InlineID:  "",
	})
	if len(fakeBot.Msg) != len(expected) {
		t.Fatalf("len(fakeBot.Msg) != len(expected), expected: %v, actual: %v", len(expected), len(fakeBot.Msg))
	}
	for i, msg := range fakeBot.Msg {
		if msg != expected[i] {
			t.Fatalf("message #%v is wrong, expected: %v, actual: %v", i, expected[i], msg)
		}
	}
}
