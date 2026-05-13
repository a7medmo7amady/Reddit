package kafka

import (
	"context"
	"encoding/json"
	"feed-service/internal/cache"
	"feed-service/internal/model"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type PostEvent struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	Community    string `json:"community"`
	AuthorID     string `json:"authorId"`
	Type         string `json:"type"`
	Upvotes      int    `json:"upvotes"`
	Downvotes    int    `json:"downvotes"`
	CommentCount int    `json:"commentCount"`
	CreatedAt    string `json:"createdAt"`
}

func StartPostConsumer(ctx context.Context, brokers []string, tc *cache.TrendingCache, pc *cache.PostCache) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "post",
		GroupID:        "feed-service-v1",
		MinBytes:       1,
		MaxBytes:       1 << 20, // 1 MB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: time.Second,
	})

	go func() {
		defer r.Close()
		log.Println("[Kafka] Post consumer started")
		for {
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[Kafka] read error: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			var event PostEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[Kafka] unmarshal error: %v", err)
				continue
			}

			post := model.Post{
				StringID:     event.ID,
				Title:        event.Title,
				Body:         event.Body,
				Community:    event.Community,
				Author:       event.AuthorID,
				Score:        event.Upvotes - event.Downvotes,
				CommentCount: event.CommentCount,
				CreatedAt:    event.CreatedAt,
			}

			if err := pc.Add(ctx, post); err != nil {
				log.Printf("[Kafka] cache write error: %v", err)
			} else {
				log.Printf("[Kafka] cached post %s in r/%s", event.ID, event.Community)
			}
		}
	}()
}
