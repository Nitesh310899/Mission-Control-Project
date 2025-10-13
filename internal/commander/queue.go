package commander

import (
	"encoding/json"
	"log"
	"mission-control/pkg/models"

	"github.com/streadway/amqp"
)

type QueueManager struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	ordersQ amqp.Queue
	statusQ amqp.Queue
}

func NewQueueManager(rmqURL string) (*QueueManager, error) {
	conn, err := amqp.Dial(rmqURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	ordersQ, err := ch.QueueDeclare("orders_queue", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	statusQ, err := ch.QueueDeclare("status_queue", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &QueueManager{
		conn:    conn,
		channel: ch,
		ordersQ: ordersQ,
		statusQ: statusQ,
	}, nil
}

func (q *QueueManager) PublishOrder(mission *models.Mission) error {
	body, err := json.Marshal(mission)
	if err != nil {
		return err
	}
	return q.channel.Publish("", q.ordersQ.Name, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (q *QueueManager) SubscribeStatusUpdates(store *MissionStore) error {
	msgs, err := q.channel.Consume(
		q.statusQ.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		return err
	}
	go func() {
		for d := range msgs {
			var update struct {
				MissionID string               `json:"mission_id"`
				Status    models.MissionStatus `json:"status"`
			}
			if err := json.Unmarshal(d.Body, &update); err == nil {
				log.Printf("[Status Update] Mission %s => %s", update.MissionID, update.Status)
				store.UpdateStatus(update.MissionID, update.Status)
			}
		}
	}()
	return nil
}

func (q *QueueManager) Close() {
	q.channel.Close()
	q.conn.Close()
}
