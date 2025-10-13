package soldier

import (
	"encoding/json"
	"mission-control/pkg/models"
    "log"
	"time"
	"github.com/streadway/amqp"
)

type QueueManager struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	ordersQ amqp.Queue
	statusQ amqp.Queue
}

func NewQueueManager(rmqURL string) (*QueueManager, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ { // retry up to 10 times
		conn, err = amqp.Dial(rmqURL)
		if err == nil {
			break
		}
		log.Printf("Waiting for RabbitMQ to be ready... retry %d/10: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if conn == nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	ordersQ, err := ch.QueueDeclare("orders_queue", true, false, false, false, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}
	statusQ, err := ch.QueueDeclare("status_queue", true, false, false, false, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &QueueManager{
		conn:    conn,
		channel: ch,
		ordersQ: ordersQ,
		statusQ: statusQ,
	}, nil
}

func (q *QueueManager) PublishStatus(missionID string, status models.MissionStatus) error {
	update := struct {
		MissionID string               `json:"mission_id"`
		Status    models.MissionStatus `json:"status"`
	}{
		MissionID: missionID,
		Status:    status,
	}
	body, err := json.Marshal(update)
	if err != nil {
		return err
	}
	return q.channel.Publish("", q.statusQ.Name, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (q *QueueManager) ConsumeOrders() (<-chan amqp.Delivery, error) {
	return q.channel.Consume(q.ordersQ.Name, "", true, false, false, false, nil)
}

func (q *QueueManager) Close() {
	q.channel.Close()
	q.conn.Close()
}

func (q *QueueManager) PublishStatusWithToken(missionID string, status models.MissionStatus, token string) error {
	update := struct {
		MissionID string               `json:"mission_id"`
		Status    models.MissionStatus `json:"status"`
		Token     string               `json:"token"`
	}{
		MissionID: missionID,
		Status:    status,
		Token:     token,
	}
	body, err := json.Marshal(update)
	if err != nil {
		return err
	}
	return q.channel.Publish("", q.statusQ.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}
