package consumer

import (
	"context"
	"encoding/json"
	"log"

	"audit-service/internal/handler"
	"audit-service/internal/model"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	group   sarama.ConsumerGroup
	topics  []string
	handler *handler.EventHandler
}

func NewKafkaConsumer(brokers []string, groupID string, topics []string, h *handler.EventHandler) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		group:   group,
		topics:  topics,
		handler: h,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	for {
		err := c.group.Consume(ctx, c.topics, &consumerGroupHandler{handler: c.handler})
		if err != nil {
			log.Println("Error consuming:", err)
		}
	}
}

type consumerGroupHandler struct {
	handler *handler.EventHandler
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for msg := range claim.Messages() {

		var event model.Event

		err := json.Unmarshal(msg.Value, &event)
		if err != nil {
			log.Println("JSON error:", err)
			continue
		}

		// call your handler
		h.handler.Handle(event)

		// commit offset
		session.MarkMessage(msg, "")
	}

	return nil
}
