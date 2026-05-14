package kafka

import (
	"context"
	"encoding/json"
	"feed-service/internal/cache"
	"feed-service/internal/model"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// ── post consumer ─────────────────────────────────────────────────────────────

type PostEvent struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	Body         string        `json:"body"`
	Community    string        `json:"community"`
	AuthorID     string        `json:"authorId"`
	Author       string        `json:"author"`
	Type         string        `json:"type"`
	Upvotes      int           `json:"upvotes"`
	Downvotes    int           `json:"downvotes"`
	CommentCount int           `json:"commentCount"`
	CreatedAt    string        `json:"createdAt"`
	Images       []model.Image `json:"images"`
	Video        *model.Video  `json:"video"`
}

func StartPostConsumer(ctx context.Context, brokers []string, tc *cache.TrendingCache, pc *cache.PostCache) {
	r := newReader(brokers, "post", "feed-service-v1")
	go consume(ctx, r, "post", func(raw []byte) error {
		var e PostEvent
		if err := json.Unmarshal(raw, &e); err != nil {
			return err
		}

		author := e.Author
		if author == "" {
			author = e.AuthorID
		}

		post := model.Post{
			StringID:     e.ID,
			Title:        e.Title,
			Body:         e.Body,
			Community:    e.Community,
			Author:       author,
			Type:         e.Type,
			Upvotes:      e.Upvotes,
			Downvotes:    e.Downvotes,
			Score:        e.Upvotes - e.Downvotes,
			CommentCount: e.CommentCount,
			CreatedAt:    e.CreatedAt,
			Images:       e.Images,
			Video:        e.Video,
		}

		if err := pc.Add(ctx, post); err != nil {
			log.Printf("[PostConsumer] PostCache.Add error: %v", err)
		}
		if err := tc.Add(ctx, post); err != nil {
			log.Printf("[PostConsumer] TrendingCache.Add error: %v", err)
		}
		return nil
	})
}

// ── community.ban consumer ────────────────────────────────────────────────────

type BanEvent struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	Community string `json:"community"`
	Action    string `json:"action"` // "BANNED" or "UNBANNED"
	Reason    string `json:"reason"`
}

func StartBanConsumer(ctx context.Context, brokers []string, bc *cache.BanCache) {
	r := newReader(brokers, "community.ban", "feed-ban-service-v1")
	go consume(ctx, r, "community.ban", func(raw []byte) error {
		var e BanEvent
		if err := json.Unmarshal(raw, &e); err != nil {
			return err
		}
		switch e.Action {
		case "BANNED":
			err := bc.Ban(ctx, e.UserID, e.Community)
			if err == nil {
				log.Printf("[BanCache] user %s (%s) banned from r/%s", e.Username, e.UserID, e.Community)
			}
			return err
		case "UNBANNED":
			err := bc.Unban(ctx, e.UserID, e.Community)
			if err == nil {
				log.Printf("[BanCache] user %s (%s) unbanned from r/%s", e.Username, e.UserID, e.Community)
			}
			return err
		default:
			return fmt.Errorf("unknown ban action: %s", e.Action)
		}
	})
}

// ── shared helpers ────────────────────────────────────────────────────────────

func newReader(brokers []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       1 << 20,
		MaxWait:        500 * time.Millisecond,
		CommitInterval: time.Second,
	})
}

func consume(ctx context.Context, r *kafka.Reader, topic string, handle func([]byte) error) {
	defer r.Close()
	log.Printf("[Kafka] consumer started for topic: %s", topic)
	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[Kafka:%s] read error: %v", topic, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := handle(msg.Value); err != nil {
			log.Printf("[Kafka:%s] handle error: %v", topic, err)
		}
	}
}
