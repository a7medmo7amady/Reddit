package realtime

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

type RedisBroker struct {
	client *redis.Client
	cb     *gobreaker.CircuitBreaker
}

func NewRedisBroker(client *redis.Client) *RedisBroker {
	st := gobreaker.Settings{
		Name:        "redis",
		MaxRequests: 1,
		Interval:    0,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}
	return &RedisBroker{client: client, cb: gobreaker.NewCircuitBreaker(st)}
}

func (r *RedisBroker) Publish(ctx context.Context, channel string, payload []byte) error {
	op := func() error {
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
		return r.client.Publish(ctx, channel, payload).Err()
	}

	_, err := r.cb.Execute(func() (any, error) {
		b := backoff.WithMaxRetries(backoff.NewConstantBackOff(50*time.Millisecond), 2)
		err := backoff.Retry(op, backoff.WithContext(b, ctx))
		return nil, err
	})
	return err
}

func (r *RedisBroker) Subscribe(ctx context.Context, channel string, handler func(payload []byte)) (func(), error) {
	if handler == nil {
		return func() {}, errors.New("missing handler")
	}

	pubsub := r.client.Subscribe(ctx, channel)
	ch := pubsub.Channel()

	stop := func() {
		_ = pubsub.Close()
	}

	go func() {
		for msg := range ch {
			if msg == nil {
				continue
			}
			handler([]byte(msg.Payload))
		}
	}()

	return stop, nil
}

type RedisOfflineQueue struct {
	client *redis.Client
	cb     *gobreaker.CircuitBreaker
	maxLen int64
}

func NewRedisOfflineQueue(client *redis.Client, maxLen int64) *RedisOfflineQueue {
	st := gobreaker.Settings{
		Name:        "redis-offline",
		MaxRequests: 1,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}
	if maxLen <= 0 {
		maxLen = 200
	}
	return &RedisOfflineQueue{client: client, cb: gobreaker.NewCircuitBreaker(st), maxLen: maxLen}
}

func (q *RedisOfflineQueue) key(userID string) string {
	return "offline:" + userID
}

func (q *RedisOfflineQueue) Enqueue(ctx context.Context, userID string, payload []byte) error {
	op := func() error {
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		pipe := q.client.Pipeline()
		pipe.RPush(ctx, q.key(userID), payload)
		pipe.LTrim(ctx, q.key(userID), -q.maxLen, -1)
		_, err := pipe.Exec(ctx)
		return err
	}

	_, err := q.cb.Execute(func() (any, error) {
		b := backoff.WithMaxRetries(backoff.NewConstantBackOff(50*time.Millisecond), 2)
		err := backoff.Retry(op, backoff.WithContext(b, ctx))
		return nil, err
	})
	return err
}

func (q *RedisOfflineQueue) Drain(ctx context.Context, userID string, handler func(payload []byte)) error {
	if handler == nil {
		return errors.New("missing handler")
	}

	_, err := q.cb.Execute(func() (any, error) {
		for {
			ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			item, err := q.client.LPop(ctx, q.key(userID)).Bytes()
			cancel()
			if err == redis.Nil {
				return nil, nil
			}
			if err != nil {
				return nil, err
			}
			handler(item)
		}
	})
	return err
}
