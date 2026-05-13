package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type BanCache struct {
	rdb *redis.Client
}

func NewBanCache(rdb *redis.Client) *BanCache {
	return &BanCache{rdb: rdb}
}

func banKey(userID, community string) string {
	return "ban:" + userID + ":" + community
}

func (b *BanCache) Ban(ctx context.Context, userID, community string) error {
	return b.rdb.Set(ctx, banKey(userID, community), "1", 0).Err()
}

func (b *BanCache) Unban(ctx context.Context, userID, community string) error {
	return b.rdb.Del(ctx, banKey(userID, community)).Err()
}

func (b *BanCache) IsBanned(ctx context.Context, userID, community string) (bool, error) {
	n, err := b.rdb.Exists(ctx, banKey(userID, community)).Result()
	return n > 0, err
}
