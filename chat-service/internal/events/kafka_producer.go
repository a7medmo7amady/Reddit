package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer publishes domain events to Kafka topics.
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a producer that writes JSON-encoded events.
// brokers example: []string{"kafka:9092"}
func NewKafkaProducer(brokers []string) *KafkaProducer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}
	return &KafkaProducer{writer: w}
}

func (p *KafkaProducer) Publish(ctx context.Context, topic string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: topic,
		Value: data,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("[kafka] publish error topic=%s: %v", topic, err)
		return err
	}

	log.Printf("[kafka] published topic=%s size=%d", topic, len(data))
	return nil
}

// Close flushes buffered messages and releases resources.
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
