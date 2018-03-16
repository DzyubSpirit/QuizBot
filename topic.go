package main

import (
	"strings"
	"net/http"
	"encoding/json"
	"log"
	"fmt"
)

type Topic struct {
	Questions []Question `json:"questions"`
}

type Question struct {
	Text          string   `json:"text"`
	AnswersNumber int      `json:"answersNumber"`
	RightAnswer   string   `json:"rightAnswer"`
	WrongAnswers  []string `json:"wrongAnswers"`
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

type Questionable interface {
	ToQuestions() []Question
}

type BookAfterQuestion struct {
	AnswersNumber int
	Books         []string
	Question      string
}

func (q BookAfterQuestion) ToQuestions() []Question {
	qs := make([]Question, len(q.Books)-1)
	for i := range qs {
		wrongAnswers := make([]string, len(q.Books)-1)
		for j := range wrongAnswers {
			if j < i {
				wrongAnswers[j] = q.Books[j]
			} else {
				wrongAnswers[j] = q.Books[j+1]
			}
		}
		qs[i] = Question{
			Text:          fmt.Sprintf("%s\n'%s'?", q.Question, q.Books[i]),
			AnswersNumber: q.AnswersNumber,
			RightAnswer:   q.Books[i+1],
			WrongAnswers:  wrongAnswers,
		}
	}
	return qs
}

type BeforeAfterQuestion struct {
	Books    []string
	Question string
}

func (q BeforeAfterQuestion) ToQuestions() []Question {
	qs := make([]Question, 0, len(q.Books)*(len(q.Books)-1))
	for i, b1 := range q.Books {
		for j, b2 := range q.Books {
			if i == j {
				continue
			}
			if i < j {
				qs = append(qs, Question{
					AnswersNumber: 2,
					WrongAnswers:  []string{"After"},
					RightAnswer:   "Before",
					Text:          fmt.Sprintf(q.Question, b1, b2),
				})
			} else {
				qs = append(qs, Question{
					AnswersNumber: 2,
					WrongAnswers:  []string{"Before"},
					RightAnswer:   "After",
					Text:          fmt.Sprintf(q.Question, b1, b2),
				})
			}
		}
	}
	return qs
}
