package handler

import (
	"audit-service/pkg/logger"

	"github.com/IBM/sarama"
)

type TopicHandler interface {
	Handle(msg *sarama.ConsumerMessage) error
}

type Dispatcher struct {
	handlers map[string]TopicHandler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{handlers: make(map[string]TopicHandler)}
}

func (d *Dispatcher) Register(topic string, h TopicHandler) {
	d.handlers[topic] = h
}

func (d *Dispatcher) Dispatch(msg *sarama.ConsumerMessage) {
	h, ok := d.handlers[msg.Topic]
	if !ok {
		logger.Log.Warn("no handler registered for topic", "topic", msg.Topic)
		return
	}
	if err := h.Handle(msg); err != nil {
		logger.Log.Error("handler error", "topic", msg.Topic, "error", err)
	}
}
