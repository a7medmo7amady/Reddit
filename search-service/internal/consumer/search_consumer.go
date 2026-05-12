package consumer

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"search-service/internal/model"
	"search-service/internal/service"
	"strings"

	"github.com/IBM/sarama"
)

type SearchConsumer struct {
	searchService service.SearchService
	consumer      sarama.Consumer
}

func NewSearchConsumer(searchService service.SearchService) (*SearchConsumer, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(strings.Split(brokers, ","), config)
	if err != nil {
		return nil, err
	}

	return &SearchConsumer{
		searchService: searchService,
		consumer:      consumer,
	}, nil
}

func (c *SearchConsumer) Start(ctx context.Context) {
	topics := []string{"post.created", "post.deleted", "comment.created", "comment.deleted"}

	for _, topic := range topics {
		partitionConsumer, err := c.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Error starting consumer for topic %s: %v", topic, err)
			continue
		}

		go func(pc sarama.PartitionConsumer, t string) {
			defer pc.Close()
			for {
				select {
				case msg := <-pc.Messages():
					c.handleMessage(ctx, t, msg.Value)
				case <-ctx.Done():
					return
				}
			}
		}(partitionConsumer, topic)
	}
}

func (c *SearchConsumer) handleMessage(ctx context.Context, topic string, value []byte) {
	log.Printf("Received message on topic %s", topic)

	switch topic {
	case "post.created":
		var post model.Post
		if err := json.Unmarshal(value, &post); err == nil {
			c.searchService.IndexPost(ctx, post)
		}
	case "post.deleted":
		var msg struct{ ID string `json:"id"` }
		if err := json.Unmarshal(value, &msg); err == nil {
			c.searchService.DeletePost(ctx, msg.ID)
		}
	case "comment.created":
		var comment model.Comment
		if err := json.Unmarshal(value, &comment); err == nil {
			c.searchService.IndexComment(ctx, comment)
		}
	case "comment.deleted":
		var msg struct{ ID string `json:"id"` }
		if err := json.Unmarshal(value, &msg); err == nil {
			c.searchService.DeleteComment(ctx, msg.ID)
		}
	}
}

func (c *SearchConsumer) Close() error {
	return c.consumer.Close()
}
