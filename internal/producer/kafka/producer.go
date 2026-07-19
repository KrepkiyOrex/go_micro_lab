package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

type TaskCreatedEvent struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) PublishTaskCreated(ctx context.Context, taskID string) error {
	event := TaskCreatedEvent{
		TaskID: taskID,
		Status: "pending",
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		log.Printf("❌ Failed to send message: %v", err)
		return err
	}

	log.Printf("✅ Event sent to Kafka: topic=%s, partition=%d, offset=%d", p.topic, partition, offset)
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}