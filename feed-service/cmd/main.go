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
	"feed-service/pkg/consul"

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
	pc := cache.NewPostCache(rdb)
	bc := cache.NewBanCache(rdb)

	// ── Kafka consumers ───────────────────────────────────────────────────────
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	kafka.StartPostConsumer(ctx, brokers, tc, pc)
	kafka.StartBanConsumer(ctx, brokers, bc)

	// ── Consul service discovery ──────────────────────────────────────────────
	// Static fallbacks are used when Consul is unreachable or a service hasn't
	// registered itself yet.
	consulAddr := getEnv("CONSUL_ADDR", "consul:8500")
	resolver := consul.New(consulAddr, map[string]string{
		"video": "http://" + getEnv("VIDEO_REST_ADDR", "video-service:8083"),
		"user":  "http://" + getEnv("USER_SERVICE_ADDR", "user-service:8080"),
	})

	videoAddr := resolver.Resolve("video")
	if videoAddr == "" {
		videoAddr = "http://" + getEnv("VIDEO_REST_ADDR", "video-service:8083")
		log.Printf("[Consul] using VIDEO_REST_ADDR fallback: %s", videoAddr)
	}
	// Strip the http:// prefix — VideoClient builds its own scheme
	videoHost := strings.TrimPrefix(videoAddr, "http://")

	// ── VideoService client (HTTP/REST) ───────────────────────────────────────
	videoClient, err := svcgrpc.NewVideoClient(videoHost)
	if err != nil {
		log.Printf("[VideoClient] init failed: %v (continuing without it)", err)
	} else {
		log.Printf("[VideoClient] connected → %s (via Consul/fallback)", videoHost)
		go func() {
			seededCommunities := []string{
				"programming", "golang", "python", "gaming", "worldnews",
				"science", "technology", "AskReddit", "linux", "webdev",
			}
			for _, community := range seededCommunities {
				if err := videoClient.SyncCommunityPosts(ctx, community, pc, tc); err != nil {
					log.Printf("[VideoClient] sync r/%s: %v", community, err)
				}
			}
		}()
	}

	// ── UserService gRPC client ───────────────────────────────────────────────
	userAddr := resolver.Resolve("user")
	if userAddr == "" {
		userAddr = getEnv("USER_GRPC_ADDR", "user-service:50051")
	} else {
		// Consul returns HTTP URL; gRPC needs host:port
		userAddr = strings.TrimPrefix(userAddr, "http://")
	}
	userClient, err := svcgrpc.NewUserClient(userAddr)
	if err != nil {
		log.Printf("[gRPC] UserService client init failed: %v (continuing without it)", err)
	} else {
		log.Printf("[gRPC] UserService client connected → %s", userAddr)
	}
	_ = userClient

	// ── gRPC server ───────────────────────────────────────────────────────────
	grpcPort := getEnv("GRPC_PORT", "50052")
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
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/posts/trending", handler.Trending(tc))
	r.GET("/posts/community/:name", handler.CommunityFeed(pc, bc))
	r.GET("/posts/feed", handler.UserFeed(pc, bc))

	port := getEnv("PORT", "8081")
	log.Printf("feed-service HTTP listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
