package cache

import (
	"context"
	"encoding/json"
	"feed-service/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

const trendingKey = "trending:posts"
const trendingTTL = 10 * time.Minute

var staticTrending = []model.Post{
	{ID: 1, Community: "programming", Author: "torvalds", Score: 94200, CommentCount: 3100, CreatedAt: "2025-05-12T08:00:00Z", Title: "I just realized I've been writing the same bug for 20 years", Body: "It's always an off-by-one error. Always."},
	{ID: 2, Community: "golang", Author: "rob_pike", Score: 81700, CommentCount: 2400, CreatedAt: "2025-05-12T09:10:00Z", Title: "Go 1.25 is out — generics just got faster", Body: "Compile times are down 30% in the new release. Full notes in the link."},
	{ID: 3, Community: "worldnews", Author: "newsbot", Score: 76500, CommentCount: 8900, CreatedAt: "2025-05-12T07:30:00Z", Title: "Scientists confirm fastest internet speed record: 402 Tbps", Body: "Researchers at UCL broke the record using a new multi-band fiber optic technique."},
	{ID: 4, Community: "gaming", Author: "xXslayer99Xx", Score: 71200, CommentCount: 5600, CreatedAt: "2025-05-11T22:00:00Z", Title: "After 3,000 hours I finally beat the final boss without taking damage", Body: "No summons, no shields, no cheese. Pure skill. I'm finally free."},
	{ID: 5, Community: "science", Author: "dr_cosmos", Score: 68900, CommentCount: 1800, CreatedAt: "2025-05-12T06:00:00Z", Title: "NASA confirms water ice found in permanently shadowed craters on the Moon", Body: "The findings significantly boost the case for a sustainable lunar base."},
	{ID: 6, Community: "technology", Author: "hackernews_mirror", Score: 65400, CommentCount: 2200, CreatedAt: "2025-05-11T18:00:00Z", Title: "Chrome now uses 40% less RAM — here's how they did it", Body: "A deep dive into the new memory management system shipped in Chrome 130."},
	{ID: 7, Community: "AskReddit", Author: "curious_carl", Score: 61100, CommentCount: 14300, CreatedAt: "2025-05-12T10:00:00Z", Title: "What's a skill you learned during COVID that you still use every day?", Body: ""},
	{ID: 8, Community: "python", Author: "guido_fan", Score: 57800, CommentCount: 1900, CreatedAt: "2025-05-11T14:00:00Z", Title: "Python 3.14 drops the GIL by default — what this means for your code", Body: "Free-threaded Python is now the default build. Here's what breaks and what doesn't."},
	{ID: 9, Community: "aww", Author: "dogmom2025", Score: 54300, CommentCount: 890, CreatedAt: "2025-05-12T11:00:00Z", Title: "My dog figured out how to open the fridge. We are in crisis.", Body: ""},
	{ID: 10, Community: "linux", Author: "arch_enjoyer", Score: 51600, CommentCount: 3400, CreatedAt: "2025-05-11T20:00:00Z", Title: "Wayland finally works perfectly on my setup after 4 years of trying", Body: "Screen sharing, gaming, dual monitors — all working. I don't know what changed but I'm not touching anything."},
	{ID: 11, Community: "dataisbeautiful", Author: "chart_wizard", Score: 48900, CommentCount: 760, CreatedAt: "2025-05-12T05:00:00Z", Title: "Visualizing every programming language's popularity since 1990", Body: "Animated bar chart race. COBOL's brief 2020 comeback is hilarious."},
	{ID: 12, Community: "java", Author: "spring_boot_dev", Score: 44200, CommentCount: 2100, CreatedAt: "2025-05-11T16:00:00Z", Title: "Java 25 virtual threads vs Go goroutines: a real benchmark", Body: "Ran 1M concurrent tasks on identical hardware. Results are surprising."},
	{ID: 13, Community: "movies", Author: "cinephile_99", Score: 41700, CommentCount: 6700, CreatedAt: "2025-05-12T12:00:00Z", Title: "Unpopular opinion: Interstellar's third act is bad and it ruins the film", Body: "Fight me. The moment they go full time-travel nonsense it loses everything that made it great."},
	{ID: 14, Community: "learnprogramming", Author: "bootcamp_grad", Score: 38500, CommentCount: 4400, CreatedAt: "2025-05-11T12:00:00Z", Title: "I got my first dev job 8 months after starting from zero. Here's exactly what I did.", Body: "No CS degree, no bootcamp. Just a plan and a lot of LeetCode suffering."},
	{ID: 15, Community: "rust", Author: "crab_lover", Score: 35900, CommentCount: 2800, CreatedAt: "2025-05-11T10:00:00Z", Title: "Rust is now the #2 language in the Linux kernel after C", Body: "New drivers are being submitted in Rust by default. The borrow checker wins."},
	{ID: 16, Community: "webdev", Author: "css_pain", Score: 33100, CommentCount: 3900, CreatedAt: "2025-05-10T22:00:00Z", Title: "CSS is finally getting native masonry layout in 2025", Body: "No more JavaScript hacks for Pinterest-style grids. The spec just landed in Chrome Canary."},
	{ID: 17, Community: "todayilearned", Author: "til_bot", Score: 30700, CommentCount: 1200, CreatedAt: "2025-05-12T03:00:00Z", Title: "TIL the original Unix timestamp was chosen so the number would be aesthetically round in 2001", Body: ""},
	{ID: 18, Community: "devops", Author: "k8s_admin", Score: 28400, CommentCount: 1700, CreatedAt: "2025-05-11T08:00:00Z", Title: "We deleted Kubernetes and replaced it with a single bash script. Productivity went up.", Body: "Controversial, but hear me out. We had 3 services. We did not need a cluster."},
	{ID: 19, Community: "MachineLearning", Author: "grad_student", Score: 25900, CommentCount: 3300, CreatedAt: "2025-05-10T18:00:00Z", Title: "GPT-5 scores 98% on the bar exam but still can't reliably count letters in a word", Body: "Capability is deeply uneven. New paper with reproducible benchmarks."},
	{ID: 20, Community: "golang", Author: "gopher42", Score: 23100, CommentCount: 980, CreatedAt: "2025-05-10T14:00:00Z", Title: "Writing a Redis clone in Go over a weekend — what I learned", Body: "RESP protocol is simpler than I expected. Persistence is where it gets interesting."},
}

type TrendingCache struct {
	rdb *redis.Client
}

func NewTrendingCache(rdb *redis.Client) *TrendingCache {
	return &TrendingCache{rdb: rdb}
}

func (c *TrendingCache) Seed(ctx context.Context) error {
	data, err := json.Marshal(staticTrending)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, trendingKey, data, trendingTTL).Err()
}

func (c *TrendingCache) Get(ctx context.Context) ([]model.Post, error) {
	val, err := c.rdb.Get(ctx, trendingKey).Bytes()
	if err == redis.Nil {
		// cache miss — re-seed and return static data
		_ = c.Seed(ctx)
		return staticTrending, nil
	}
	if err != nil {
		return nil, err
	}
	var posts []model.Post
	if err := json.Unmarshal(val, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}
