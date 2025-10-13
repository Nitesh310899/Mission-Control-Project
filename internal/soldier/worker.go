package soldier

import (
	"encoding/json"
	"log"
	"math/rand"
	"mission-control/pkg/models"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type Worker struct {
	queue       *QueueManager
	concurrency int
}

func NewWorker(q *QueueManager, concurrency int) *Worker {
	return &Worker{
		queue:       q,
		concurrency: concurrency,
	}
}

func (w *Worker) Start() error {
	msgs, err := w.queue.ConsumeOrders()
	if err != nil {
		return err
	}

	sem := make(chan struct{}, w.concurrency)
	var wg sync.WaitGroup

	for d := range msgs {
		sem <- struct{}{}
		wg.Add(1)
		go func(delivery amqp.Delivery) {
			defer wg.Done()
			defer func() { <-sem }()
			w.handleMission(delivery.Body)
		}(d)
	}
	wg.Wait()
	return nil
}

func (w *Worker) handleMission(data []byte) {
	var mission models.Mission
	if err := json.Unmarshal(data, &mission); err != nil {
		log.Printf("[Worker] Invalid mission data: %v", err)
		return
	}

	log.Printf("[Worker] Received mission %s, publishing IN_PROGRESS", mission.ID)
	//	err := w.queue.PublishStatus(mission.ID, models.StatusInProgress)
	err := w.queue.PublishStatusWithToken(mission.ID, models.StatusInProgress, w.tokenManager.Token())
	if err != nil {
		log.Printf("[Worker] Failed to publish IN_PROGRESS: %v", err)
		return
	}

	// Simulate mission execution delay 5-15 seconds
	rand.Seed(time.Now().UnixNano())
	delay := time.Duration(rand.Intn(11)+5) * time.Second
	time.Sleep(delay)

	// Decide outcome (90% success rate)
	success := rand.Intn(100) < 90
	finalStatus := models.StatusCompleted
	if !success {
		finalStatus = models.StatusFailed
	}
	log.Printf("[Worker] Mission %s completed with status %s", mission.ID, finalStatus)

	//err = w.queue.PublishStatus(mission.ID, finalStatus)
	err = w.queue.PublishStatusWithToken(mission.ID, finalStatus, w.tokenManager.Token())
	if err != nil {
		log.Printf("[Worker] Failed to publish final status: %v", err)
	}
}
