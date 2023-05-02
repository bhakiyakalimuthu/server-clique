package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bhakiyakalimuthu/server-clique/types"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

//const (
//	queueName = "clique"
//)

type queue struct {
	logger    *zap.Logger
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string
	appID     string
}

func New(connStr, queueName, appID string) (*queue, error) {
	conn, err := amqp.Dial(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel %v", err)
	}
	_, err = ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %v", err)
	}
	return &queue{
		conn:      conn,
		ch:        ch,
		queueName: queueName,
		appID:     appID,
	}, nil
}

func (q *queue) Publish(message *types.Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return q.ch.Publish("", q.queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Timestamp:   time.Now(),
		MessageId:   uuid.New().String(),
		AppId:       q.appID,
		Body:        body,
	})
}

func (q *queue) Consume(ctx context.Context) (<-chan *types.Message, error) {
	deliveryChan, err := q.ch.Consume(q.queueName, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	msgChan := make(chan *types.Message)
	go func() {
		defer close(msgChan)
		for {
			select {
			case msg := <-deliveryChan:
				m := new(types.Message)
				if err := json.Unmarshal(msg.Body, &m); err != nil {
					q.logger.Error("failed to unmarshal message body", zap.Error(err))
					continue
				}
				select {
				// make sure that none of the msg get into msgChan  after context gets cancelled
				case <-ctx.Done():
					return
				default:
				}
				msgChan <- m
			case <-ctx.Done():
				return
			}
		}
	}()
	return msgChan, nil
}
