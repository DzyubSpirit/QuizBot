package main

import (
	"strings"
	"net/http"
	"encoding/json"
	"log"
)

type Topic struct {
	Books    []string `json:"books"`
	Question string   `json:"question"`
}

var topics = map[string]Topic{
	"ru": {
		Books:    []string{"Бытие", "Исход", "Левит", "Числа", "Второзаконие", "Иисус Навин", "Судьей", "Руфь", "1 Царств", "2 Царств", "3 Царств", "4 Царств", "1 Паралипоменон", "2 Паралипоменон", "Ездра", "Неемия", "Есфирь", "Иов", "Псалтырь", "Притчи", "Екклесиаст", "Книга Песнь Песней", "Исаия", "Иеремия", "Плач Иеремии", "Иезекииль", "Даниил", "Осия", "Иоиль", "Амос", "Авдий", "Иона", "Михей", "Наум", "Аввакум", "Софония", "Аггей", "Захария", "Малахия", "Матфея", "Марка", "Луки", "Иоанна", "Деяния", "Иакова", "1 Петра", "2 Петра", "1 Иоанна", "2 Иоанна", "3 Иоанна", "Иуды", "Римлянам", "1 Коринфянам", "2 Коринфянам", "Галатам", "Ефесянам", "Филиппийцам", "Колоссянам", "1 Фессалоникийцам", "2 Фессалоникийцам", "1 Тимофею", "2 Тимофею", "Титу", "Филимону", "Евреям", "Откровение"},
		Question: "Какая книга идет после ",
	},
	"en": {
		Books:    []string{"Genesis", "Exodus", "Leviticus", "Numbers", "Deuteronomy", "Joshua", "Judges", "Ruth", "1 Samuel", "2 Samuel", "1 Kings", "2 Kings", "1 Chronicles", "2 Chronicles", "Ezra", "Nehemiah", "Esther", "Job", "Psalms", "Proverbs", "Ecclesiastes", "Song of Songs", "Isaiah", "Jeremiah", "Lamentations", "Ezekiel", "Daniel", "Hosea", "Joel", "Amos", "Obadiah", "Jonah", "Micah", "Nahum", "Habakkuk", "Zephaniah", "Haggai", "Zechariah", "Malachi"},
		Question: "What book is after ",
	},
}

func TopicsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	topic := strings.Trim(r.URL.Path, "/api/topics/")
	err := json.NewEncoder(w).Encode(topics[topic])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: error marshaling topic: %v", err)
	}
}
