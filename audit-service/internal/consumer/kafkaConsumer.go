package consumer

import (
	"context"

	"audit-service/internal/handler"
	"audit-service/pkg/logger"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	group      sarama.ConsumerGroup
	topics     []string
	dispatcher *handler.Dispatcher
}

func NewKafkaConsumer(brokers []string, groupID string, topics []string, d *handler.Dispatcher) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		group:      group,
		topics:     topics,
		dispatcher: d,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	for {
		if err := c.group.Consume(ctx, c.topics, &consumerGroupHandler{dispatcher: c.dispatcher}); err != nil {
			logger.Log.Error("consumer group error", "error", err)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

type consumerGroupHandler struct {
	dispatcher *handler.Dispatcher
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.dispatcher.Dispatch(msg)
		session.MarkMessage(msg, "")
	}
	return nil
}
