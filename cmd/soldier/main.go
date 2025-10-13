package main

import (
	"log"
	"mission-control/internal/soldier"
	"os"
	"strconv"
)

const rmqURL = "amqp://guest:guest@rabbitmq:5672/"

func main() {
	// Set concurrency from env or default 5
	concurrency := 5
	if val := os.Getenv("CONCURRENCY"); val != "" {
		if c, err := strconv.Atoi(val); err == nil && c > 0 {
			concurrency = c
		}
	}

	// Connect to RabbitMQ
	queue, err := soldier.NewQueueManager(rmqURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer queue.Close()

	// Setup token manager with initial token and renewal endpoint
	// TODO: Replace with actual initial token acquisition
	initialToken := os.Getenv("INITIAL_TOKEN")
	if initialToken == "" {
		log.Fatal("INITIAL_TOKEN environment variable must be set")
	}
	soldierID := os.Getenv("SOLDIER_ID")
	if soldierID == "" {
		soldierID = "soldier-1" // default soldier ID
	}

	renewURL := os.Getenv("TOKEN_RENEW_URL")
	if renewURL == "" {
		renewURL = "http://commander-service:8080/tokens/renew"
	}

	tokenManager := soldier.NewTokenManager(soldierID, renewURL)
	tokenManager.Start(initialToken, 30) // assuming 30 seconds expiry for initial token

	// Create and start the worker with token manager
	worker := soldier.NewWorker(queue, concurrency, tokenManager)
	log.Printf("Starting soldier worker with concurrency: %d", concurrency)

	err = worker.Start()
	if err != nil {
		log.Fatalf("Worker error: %v", err)
	}
}
