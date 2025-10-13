package main

import (
    "log"
    "mission-control/internal/soldier"
    "os"
    "strconv"
    "fmt"
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

    soldierID := os.Getenv("SOLDIER_ID")
    if soldierID == "" {
        soldierID = "soldier-1" // default soldier ID
    }

    commanderURL := os.Getenv("COMMANDER_URL")
    if commanderURL == "" {
        // Adjust as per your service name; using commander for Docker Compose networking
        commanderURL = "http://commander:8080"
    }

    // Dynamically acquire initial token from Commander token issuance API
    initialToken, expiresIn, err := soldier.GetInitialToken(commanderURL, soldierID)
    if err != nil {
        log.Fatalf("Failed to acquire initial token: %v", err)
    }
    log.Printf("Obtained initial token for soldier %s", soldierID)

    renewURL := os.Getenv("TOKEN_RENEW_URL")
    if renewURL == "" {
        renewURL = fmt.Sprintf("%s/tokens/renew", commanderURL)
    }

    tokenManager := soldier.NewTokenManager(soldierID, renewURL)
    tokenManager.Start(initialToken, expiresIn)

    // Create and start the worker with token manager
    worker := soldier.NewWorker(queue, concurrency, tokenManager)
    log.Printf("Starting soldier worker with concurrency: %d", concurrency)

    err = worker.Start()
    if err != nil {
        log.Fatalf("Worker error: %v", err)
    }
}
