package cache

import (
	"context"
	"encoding/json"
	"feed-service/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

const postTTL = 24 * time.Hour
const maxPostsPerCommunity = 100

type PostCache struct {
	rdb *redis.Client
}

func NewPostCache(rdb *redis.Client) *PostCache {
	return &PostCache{rdb: rdb}
}

func communityKey(community string) string {
	return "posts:community:" + community
}

func (c *PostCache) Add(ctx context.Context, post model.Post) error {
	data, err := json.Marshal(post)
	if err != nil {
		return err
	}

	score := float64(time.Now().UnixMilli())
	key := communityKey(post.Community)
	existing, err := c.rdb.ZRange(ctx, key, 0, -1).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := c.rdb.Pipeline()
	for _, value := range existing {
		var cached model.Post
		if json.Unmarshal([]byte(value), &cached) == nil && cached.StringID == post.StringID {
			pipe.ZRem(ctx, key, value)
		}
	}
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)})
	pipe.ZRemRangeByRank(ctx, key, 0, int64(-maxPostsPerCommunity-1))
	pipe.Expire(ctx, key, postTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *PostCache) GetByCommunity(ctx context.Context, community string, limit int) ([]model.Post, error) {
	key := communityKey(community)
	vals, err := c.rdb.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	return decodePosts(vals), nil
}
func (c *PostCache) GetByCommunities(ctx context.Context, communities []string, limit int) ([]model.Post, error) {
	if len(communities) == 0 {
		return nil, nil
	}

	keys := make([]string, len(communities))
	for i, c := range communities {
		keys[i] = communityKey(c)
	}
	tmpKey := "posts:union:" + communities[0]

	pipe := c.rdb.Pipeline()
	pipe.ZUnionStore(ctx, tmpKey, &redis.ZStore{Keys: keys})
	pipe.Expire(ctx, tmpKey, 30*time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	vals, err := c.rdb.ZRevRange(ctx, tmpKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	return decodePosts(vals), nil
}

func (c *PostCache) GetByID(ctx context.Context, id string) (model.Post, error) {
	var cursor uint64
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, "posts:community:*", 100).Result()
		if err != nil {
			return model.Post{}, err
		}
		for _, key := range keys {
			vals, err := c.rdb.ZRevRange(ctx, key, 0, -1).Result()
			if err != nil {
				continue
			}
			for _, v := range vals {
				var p model.Post
				if json.Unmarshal([]byte(v), &p) == nil && p.StringID == id {
					return p, nil
				}
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return model.Post{}, redis.Nil
}

func decodePosts(vals []string) []model.Post {
	posts := make([]model.Post, 0, len(vals))
	seen := make(map[string]struct{}, len(vals))
	for _, v := range vals {
		var p model.Post
		if json.Unmarshal([]byte(v), &p) == nil {
			if p.StringID != "" {
				if _, ok := seen[p.StringID]; ok {
					continue
				}
				seen[p.StringID] = struct{}{}
			}
			posts = append(posts, p)
		}
	}
	return posts
}
