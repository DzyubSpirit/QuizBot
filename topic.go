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

type TopicsMap map[string]Topic

func (topics TopicsMap) TopicsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	topic := strings.Trim(r.URL.Path, "/api/topics/")
	err := json.NewEncoder(w).Encode(topics[topic])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: error marshaling topic: %v", err)
	}
}
