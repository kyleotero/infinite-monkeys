package main

import (
	"context"
	"encoding/json"
	"fmt"
	"infinite-monkey-theorem/src/benchmark"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

var monkeys [][]string

func Process(w http.ResponseWriter, r *http.Request) {
	var targetWord word
	bench := &benchmark.Benchmark{}

	err := json.NewDecoder(r.Body).Decode(&targetWord)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)

	data, err := os.ReadFile("./monkeys.json")
	if err != nil {
		log.Println(err)
	}

	if err := json.Unmarshal(data, &monkeys); err != nil {
		log.Fatalf("Error unmarshaling assets data: %v", err)
	}

	go func() {
		log.Println("the word is", targetWord.Target)

		content := fmt.Sprintf("A new string has been recieved. It is **%s**", targetWord.Target)
		webhook(content)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		resultChan := make(chan int64)
		workersStr := os.Getenv("WORKERS")

		workers, err := strconv.Atoi(workersStr)
		if err != nil {
			log.Fatalf("Error converting WORKERS environment variable to int: %v", err)
		}

		for i := 0; i < workers; i++ {
			go simulate(ctx, cancel, targetWord.Target, resultChan, bench)
		}

		res := <-resultChan

		content = fmt.Sprintf("The string has been found! It took %d attempts.", res)
		webhook(content)

		avgTime := bench.OutputAvg()

		content = fmt.Sprintf("average speed of %f nanoseconds per 1 million combinations, which is %f seconds!", avgTime, avgTime/1000000000)
		webhook(content)

		log.Println("string found in", res, "tries")
	}()
}

func simulate(ctx context.Context, cancel context.CancelFunc, target string, result chan<- int64, bench *benchmark.Benchmark) {
	var count int64 = 0
	var current int = 0
	var length int = -1
	targetInts := []int{}

	for i, char := range target {
		val := int(char) - 97

		if char == ' ' {
			val = 26
		}

		if i > 0 {
			val += targetInts[i-1]
		}

		targetInts = append(targetInts, val)
	}

	for length+1 != len(target) || current != targetInts[len(targetInts)-1] {
		select {
		case <-ctx.Done():
			return
		default:
			if count%1000000 == 0 {
				now := time.Now()
				bench.LogTimestamp(now.UnixNano())
			}

			if length >= len(target) || (length >= 0 && current != targetInts[length]) {
				current = 0
				length = -1
				count++
			} else if length > 7 {
				content := fmt.Sprintf("The string is almost there! We currently have: %s, and have tried %d combinatons!", target[:length], count)
				webhook(content)
			}

			char := rand.Intn(27)
			length++

			if char == 26 {
				current += 26
				continue
			}

			current += char
		}
	}

	result <- count
	cancel()
}

func webhook(message string) {
	godotenv.Load("../../.env")
	webhook := os.Getenv("WEBHOOK")
	num := rand.Intn(5)

	requestContent := request{
		Content:  message,
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
}
