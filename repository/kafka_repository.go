package repository

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaRepository struct {
	Writer *kafka.Writer
}

func NewKafkaRepository(brokers []string, topic string) *KafkaRepository {
	return &KafkaRepository{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (r *KafkaRepository) PublishPurchase(userID, ticketName string) error {
	return r.Writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(userID),
			Value: []byte(ticketName),
		},
	)
}
