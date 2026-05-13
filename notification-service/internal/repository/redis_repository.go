package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"notification-service/internal/model"

	"github.com/redis/go-redis/v9"
)

type RedisRepository interface {
	EnqueueOfflineNotification(ctx context.Context, userID string, n *model.Notification) error
	GetOfflineNotifications(ctx context.Context, userID string) ([]model.Notification, error)
	DeleteOfflineNotifications(ctx context.Context, userID string) error
}

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) RedisRepository {
	return &redisRepository{client: client}
}

func (r *redisRepository) EnqueueOfflineNotification(ctx context.Context, userID string, n *model.Notification) error {
	key := fmt.Sprintf("offline_notifications:%s", userID)
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}

	// Push to list and set expiration (e.g., 30 days)
	pipe := r.client.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.Expire(ctx, key, 30*24*3600*1e9) // 30 days in nanoseconds for go-redis v9 duration? Wait, go-redis takes time.Duration.
	_, err = pipe.Exec(ctx)
	return err
}

func (r *redisRepository) GetOfflineNotifications(ctx context.Context, userID string) ([]model.Notification, error) {
	key := fmt.Sprintf("offline_notifications:%s", userID)
	
	// Get all elements from the list
	data, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	notifications := make([]model.Notification, 0, len(data))
	for _, d := range data {
		var n model.Notification
		if err := json.Unmarshal([]byte(d), &n); err != nil {
			continue
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (r *redisRepository) DeleteOfflineNotifications(ctx context.Context, userID string) error {
	key := fmt.Sprintf("offline_notifications:%s", userID)
	return r.client.Del(ctx, key).Err()
}
