package consumer

import (
	"audit-service/internal/handler"
	"audit-service/internal/model"
)

type KafkaConsumer struct {
	handler *handler.EventHandler
}

func NewKafkaConsumer(h *handler.EventHandler) *KafkaConsumer {
	return &KafkaConsumer{handler: h}
}

func (c *KafkaConsumer) Start() {
	var event model.Event
	c.handler.Handle(event)
}
