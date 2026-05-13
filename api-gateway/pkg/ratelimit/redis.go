package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	rdb    *redis.Client
	window time.Duration
	limit  int64
}

func New(rdb *redis.Client, window time.Duration, limit int64) *Limiter {
	return &Limiter{
		rdb:    rdb,
		window: window,
		limit:  limit,
	}
}

func (l *Limiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("rl:%s", key)

	pipe := l.rdb.Pipeline()
	incrCmd := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, l.window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}

	return incrCmd.Val() <= l.limit, nil
}

func Ping(rdb *redis.Client) error {
	return rdb.Ping(context.Background()).Err()
}
