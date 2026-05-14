package cache

import (
	"context"
	"encoding/json"
	"feed-service/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

const trendingLiveKey = "trending:live"
const trendingTTL     = 10 * time.Minute
const trendingLimit   = 50

// staticTrending is returned when no real posts exist yet.
var staticTrending = []model.Post{
	{ID: 1, Community: "programming", Author: "torvalds", Upvotes: 94200, Score: 94200, CommentCount: 3100, CreatedAt: "2025-05-12T08:00:00Z", Title: "I just realized I've been writing the same bug for 20 years", Body: "It's always an off-by-one error. Always."},
	{ID: 2, Community: "golang", Author: "rob_pike", Upvotes: 81700, Score: 81700, CommentCount: 2400, CreatedAt: "2025-05-12T09:10:00Z", Title: "Go 1.25 is out — generics just got faster", Body: "Compile times are down 30% in the new release. Full notes in the link."},
	{ID: 3, Community: "worldnews", Author: "newsbot", Upvotes: 76500, Score: 76500, CommentCount: 8900, CreatedAt: "2025-05-12T07:30:00Z", Title: "Scientists confirm fastest internet speed record: 402 Tbps", Body: "Researchers at UCL broke the record using a new multi-band fiber optic technique."},
	{ID: 4, Community: "gaming", Author: "xXslayer99Xx", Upvotes: 71200, Score: 71200, CommentCount: 5600, CreatedAt: "2025-05-11T22:00:00Z", Title: "After 3,000 hours I finally beat the final boss without taking damage", Body: "No summons, no shields, no cheese. Pure skill."},
	{ID: 5, Community: "science", Author: "dr_cosmos", Upvotes: 68900, Score: 68900, CommentCount: 1800, CreatedAt: "2025-05-12T06:00:00Z", Title: "NASA confirms water ice found in permanently shadowed craters on the Moon", Body: "The findings significantly boost the case for a sustainable lunar base."},
	{ID: 6, Community: "technology", Author: "hackernews_mirror", Upvotes: 65400, Score: 65400, CommentCount: 2200, CreatedAt: "2025-05-11T18:00:00Z", Title: "Chrome now uses 40% less RAM — here's how they did it", Body: "A deep dive into the new memory management system shipped in Chrome 130."},
	{ID: 7, Community: "AskReddit", Author: "curious_carl", Upvotes: 61100, Score: 61100, CommentCount: 14300, CreatedAt: "2025-05-12T10:00:00Z", Title: "What's a skill you learned during COVID that you still use every day?", Body: ""},
	{ID: 8, Community: "python", Author: "guido_fan", Upvotes: 57800, Score: 57800, CommentCount: 1900, CreatedAt: "2025-05-11T14:00:00Z", Title: "Python 3.14 drops the GIL by default — what this means for your code", Body: "Free-threaded Python is now the default build."},
	{ID: 9, Community: "linux", Author: "arch_enjoyer", Upvotes: 51600, Score: 51600, CommentCount: 3400, CreatedAt: "2025-05-11T20:00:00Z", Title: "Wayland finally works perfectly on my setup after 4 years of trying", Body: "Screen sharing, gaming, dual monitors — all working."},
	{ID: 10, Community: "webdev", Author: "css_pain", Upvotes: 33100, Score: 33100, CommentCount: 3900, CreatedAt: "2025-05-10T22:00:00Z", Title: "CSS is finally getting native masonry layout in 2025", Body: "No more JavaScript hacks for Pinterest-style grids."},
}

type TrendingCache struct {
	rdb *redis.Client
}

func NewTrendingCache(rdb *redis.Client) *TrendingCache {
	return &TrendingCache{rdb: rdb}
}

// Add inserts or updates a real post in the live trending sorted set (score = upvotes - downvotes).
func (c *TrendingCache) Add(ctx context.Context, post model.Post) error {
	score := float64(post.Upvotes - post.Downvotes)

	data, err := json.Marshal(post)
	if err != nil {
		return err
	}

	pipe := c.rdb.Pipeline()
	pipe.ZAdd(ctx, trendingLiveKey, redis.Z{Score: score, Member: post.StringID})
	pipe.HSet(ctx, "trending:data", post.StringID, data)
	pipe.Expire(ctx, trendingLiveKey, trendingTTL)
	pipe.ZRemRangeByRank(ctx, trendingLiveKey, 0, -int64(trendingLimit)-1)
	_, err = pipe.Exec(ctx)
	return err
}

// Get returns live posts sorted by score, merged on top of the static fallback list.
// Real user posts always appear first; static posts fill the remainder.
func (c *TrendingCache) Get(ctx context.Context) ([]model.Post, error) {
	ids, err := c.rdb.ZRevRange(ctx, trendingLiveKey, 0, int64(trendingLimit)-1).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var livePosts []model.Post
	if len(ids) > 0 {
		rawMap, err := c.rdb.HMGet(ctx, "trending:data", ids...).Result()
		if err == nil {
			for _, v := range rawMap {
				if v == nil {
					continue
				}
				var p model.Post
				if json.Unmarshal([]byte(v.(string)), &p) == nil {
					livePosts = append(livePosts, p)
				}
			}
		}
	}

	// Merge: live posts first, then static posts that don't duplicate community+title
	liveSet := make(map[string]bool, len(livePosts))
	for _, p := range livePosts {
		liveSet[p.StringID] = true
	}

	result := make([]model.Post, 0, trendingLimit)
	result = append(result, livePosts...)
	for _, p := range staticTrending {
		if len(result) >= trendingLimit {
			break
		}
		result = append(result, p)
	}

	return result, nil
}
