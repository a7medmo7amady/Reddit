package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"feed-service/internal/cache"
	"feed-service/internal/handler"
	"feed-service/internal/kafka"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	ctx := context.Background()

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	tc := cache.NewTrendingCache(rdb)
	if err := tc.Seed(ctx); err != nil {
		log.Fatalf("seed trending cache: %v", err)
	}
	log.Println("Trending posts seeded into Redis")

	pc := cache.NewPostCache(rdb)

	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	kafka.StartPostConsumer(ctx, brokers, tc, pc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/posts/trending", handler.Trending(tc))
	r.GET("/posts/community/:name", handler.CommunityFeed(pc))
	r.GET("/posts/feed", handler.UserFeed(pc))

	port := getEnv("PORT", "8081")
	log.Printf("feed-service listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
