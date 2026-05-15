package cache

import (
	"context"
	"encoding/json"
	"feed-service/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

const trendingLiveKey  = "trending:live"
const trendingDataKey  = "trending:data"
const trendingTTL      = 24 * time.Hour
const trendingDataTTL  = 7 * 24 * time.Hour
const trendingLimit    = 50

type TrendingCache struct {
	rdb *redis.Client
}

func NewTrendingCache(rdb *redis.Client) *TrendingCache {
	return &TrendingCache{rdb: rdb}
}

func (c *TrendingCache) Add(ctx context.Context, post model.Post) error {
	score := float64(post.Upvotes - post.Downvotes)

	data, err := json.Marshal(post)
	if err != nil {
		return err
	}

	pipe := c.rdb.Pipeline()
	pipe.ZAdd(ctx, trendingLiveKey, redis.Z{Score: score, Member: post.StringID})
	pipe.HSet(ctx, trendingDataKey, post.StringID, data)
	pipe.Expire(ctx, trendingLiveKey, trendingTTL)
	pipe.Expire(ctx, trendingDataKey, trendingDataTTL)
	pipe.ZRemRangeByRank(ctx, trendingLiveKey, 0, -int64(trendingLimit)-1)
	_, err = pipe.Exec(ctx)
	return err
}
// AddIfNotExists adds a post to trending only if it isn't already tracked.
// Used for startup sync so real vote scores aren't overwritten with 0.
func (c *TrendingCache) AddIfNotExists(ctx context.Context, post model.Post) error {
	score := float64(post.Upvotes - post.Downvotes)
	data, err := json.Marshal(post)
	if err != nil {
		return err
	}

	pipe := c.rdb.Pipeline()
	pipe.ZAddNX(ctx, trendingLiveKey, redis.Z{Score: score, Member: post.StringID})
	pipe.HSetNX(ctx, trendingDataKey, post.StringID, data)
	pipe.Expire(ctx, trendingLiveKey, trendingTTL)
	pipe.Expire(ctx, trendingDataKey, trendingDataTTL)
	pipe.ZRemRangeByRank(ctx, trendingLiveKey, 0, -int64(trendingLimit)-1)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *TrendingCache) Get(ctx context.Context) ([]model.Post, error) {
	ids, err := c.rdb.ZRevRange(ctx, trendingLiveKey, 0, int64(trendingLimit)-1).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var posts []model.Post
	if len(ids) > 0 {
		rawMap, err := c.rdb.HMGet(ctx, trendingDataKey, ids...).Result()
		if err == nil {
			for _, v := range rawMap {
				if v == nil {
					continue
				}
				var p model.Post
				if json.Unmarshal([]byte(v.(string)), &p) == nil {
					posts = append(posts, p)
				}
			}
		}
	}

	return posts, nil
}
