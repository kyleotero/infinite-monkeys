package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type word struct {
	Target string `json:"target"`
}

type request struct {
	Content  string `json:"content"`
	Username string `json:"username"`
	Image    string `json:"avatar_url"`
}

func Process(w http.ResponseWriter, r *http.Request) {
	var targetWord word

	err := json.NewDecoder(r.Body).Decode(&targetWord)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)

	go func() {
		log.Println("the word is", targetWord.Target)

		data, err := os.ReadFile("monkeys.json")
		if err != nil {
			log.Println(err)
		}

		var monkeys [][]string

		if err := json.Unmarshal(data, &monkeys); err != nil {
			log.Fatalf("Error unmarshaling assets data: %v", err)
		}

		num := rand.Intn(5)
		godotenv.Load()
		webhook := os.Getenv("WEBHOOK")
		requestContent := request{
			Content:  fmt.Sprintf("A new string has been recieved. It is **%s**", targetWord.Target),
			Username: monkeys[num][0],
			Image:    monkeys[num][1],
		}

		req, err := json.Marshal(requestContent)
		if err != nil {
			log.Println(err)
		}

		_, err = http.Post(webhook, "application/json", strings.NewReader(string(req)))

		if err != nil {
			log.Println(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		resultChan := make(chan int)
		workers := 10000

		for i := 0; i < workers; i++ {
			go simulate(ctx, cancel, targetWord.Target, resultChan)
		}

		res := <-resultChan

		requestContent = request{
			Content:  fmt.Sprintf("The string has been found! It took %d attempts.", res),
			Username: monkeys[num][0],
			Image:    monkeys[num][1],
		}

		req, err = json.Marshal(requestContent)
		if err != nil {
			log.Println(err)
		}
		_, err = http.Post(webhook, "application/json", strings.NewReader(string(req)))

		if err != nil {
			log.Println(err)
		}
	}()
}

func simulate(ctx context.Context, cancel context.CancelFunc, target string, result chan<- int) {
	count := 0
	current := ""
	for current != target {
		select {
		case <-ctx.Done():
			return
		default:
			length := len(current) - 1

			if len(current) >= len(target) || (length >= 0 && current[length] != target[length]) {
				current = ""
				count++
			}

			char := rand.Intn(27) + 97

			if char == 123 {
				current += " "
				continue
			}

			current += string(rune(char))
		}
	}

	result <- count
	cancel()
}
