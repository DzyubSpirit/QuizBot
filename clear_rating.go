package main

import (
	"flag"
	"log"
	"gopkg.in/telegram-bot-api.v4"
)

func main() {
	botToken := flag.String("bot_token", "", "token of the Telegram bot")
	gameShortName := flag.String("game_short_name", "", "short name of telegram game")
	if *botToken == "" {
		log.Fatalf("should be bot_token arg for command")
	}
	if *gameShortName == "" {
		log.Fatalf("should be bot_token arg for command")
	}
	bot, err := tgbotapi.NewBotAPI(*botToken)
	if err != nil {
		log.Fatalf("error creating bot api: %v", err)
	}
	bot.Debug = true
	bot.Send(tgbotapi.GetGameHighScoresConfig{
		InlineMessageID:"AgAAAEWOAAA6sZwHRPat_fxRCwc",
		ChatID: -3941326202853168911,
	})
}
