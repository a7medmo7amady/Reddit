package events

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

type Producer interface {
	Publish(ctx context.Context, topic string, payload any) error
}

type LogProducer struct{}

func NewLogProducer() *LogProducer {
	return &LogProducer{}
}

func (p *LogProducer) Publish(ctx context.Context, topic string, payload any) error {
	data, _ := json.Marshal(payload)
	log.Printf("[event] topic=%s payload=%s", topic, string(data))
	return nil
}

type MessageSentEvent struct {
	EventID        string    `json:"eventId"`
	EventType      string    `json:"eventType"`
	Version        int       `json:"version"`
	OccurredAt     time.Time `json:"occurredAt"`
	MessageID      string    `json:"messageId"`
	ConversationID string    `json:"conversationId"`
	SenderID       string    `json:"senderId"`
}
