package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"feed-service/internal/cache"
	svcgrpc "feed-service/internal/grpc"
	"feed-service/internal/handler"
	"feed-service/internal/kafka"
	feedpb "feed-service/pkg/proto/feed"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	ctx := context.Background()

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{Addr: getEnv("REDIS_ADDR", "localhost:6379")})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	tc := cache.NewTrendingCache(rdb)
	if err := tc.Seed(ctx); err != nil {
		log.Fatalf("seed trending cache: %v", err)
	}
	log.Println("Trending posts seeded into Redis")

	pc := cache.NewPostCache(rdb)
	bc := cache.NewBanCache(rdb)

	// ── Kafka consumers ───────────────────────────────────────────────────────
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	kafka.StartPostConsumer(ctx, brokers, tc, pc)
	kafka.StartBanConsumer(ctx, brokers, bc)

	// ── gRPC clients ──────────────────────────────────────────────────────────
	videoClient, err := svcgrpc.NewVideoClient(getEnv("VIDEO_GRPC_ADDR", "localhost:50054"))
	if err != nil {
		log.Printf("[gRPC] VideoService client init failed: %v (continuing without it)", err)
	} else {
		log.Println("[gRPC] VideoService client connected")
		// Warm up trending communities from VideoService on startup
		go func() {
			for _, community := range []string{"programming", "golang", "python", "gaming"} {
				if err := videoClient.SyncCommunityPosts(ctx, community, pc); err != nil {
					log.Printf("[gRPC] sync r/%s: %v", community, err)
				}
			}
		}()
	}

	userClient, err := svcgrpc.NewUserClient(getEnv("USER_GRPC_ADDR", "localhost:50051"))
	if err != nil {
		log.Printf("[gRPC] UserService client init failed: %v (continuing without it)", err)
	} else {
		log.Println("[gRPC] UserService client connected")
	}

	// ── gRPC server ───────────────────────────────────────────────────────────
	grpcPort := getEnv("GRPC_PORT", "50053")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	feedpb.RegisterFeedServiceServer(grpcSrv, svcgrpc.NewFeedServer(pc))
	go func() {
		log.Printf("feed-service gRPC listening on :%s", grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	// ── HTTP server ───────────────────────────────────────────────────────────
	_ = userClient // available for handlers that need user enrichment

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/posts/trending", handler.Trending(tc))
	r.GET("/posts/community/:name", handler.CommunityFeed(pc))
	r.GET("/posts/feed", handler.UserFeed(pc))

	port := getEnv("PORT", "8081")
	log.Printf("feed-service HTTP listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
