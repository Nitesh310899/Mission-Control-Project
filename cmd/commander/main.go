package main

import (
	"log"
	"mission-control/internal/commander"
	"net/http"
	"os"
)

const rmqURL = "amqp://guest:guest@rabbitmq:5672/"

func main() {
	store := commander.NewMissionStore()
	queue, err := commander.NewQueueManager(rmqURL)
	if err != nil {
		log.Fatalf("Queue error: %v", err)
	}
	defer queue.Close()
	handler := commander.NewHandler(store, queue)
	if err := queue.SubscribeStatusUpdates(store); err != nil {
		log.Fatalf("Subscribe error: %v", err)
	}
	http.HandleFunc("/missions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handler.PostMission(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/missions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handler.GetMissionStatus(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/tokens/renew", handler.RenewToken)
	http.HandleFunc("/tokens/issue", handler.IssueToken)


	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Commander API running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
