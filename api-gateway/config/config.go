package config

import "os"

type Config struct {
	Port      string
	JWTSecret string

	GRPCPort string

	RedisAddr  string
	ConsulAddr string

	UserServiceURL         string
	FeedServiceURL         string
	SearchServiceURL       string
	VideoServiceURL        string
	NotificationServiceURL string

	UserGRPCAddr         string
	FeedGRPCAddr         string
	SearchGRPCAddr       string
	VideoGRPCAddr        string
	NotificationGRPCAddr string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8088"),
		JWTSecret: getEnv("JWT_SECRET", "changeme"),
		GRPCPort:  getEnv("GRPC_PORT", "9090"),

		RedisAddr:  getEnv("REDIS_ADDR", ""),
		ConsulAddr: getEnv("CONSUL_ADDR", ""),

		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://user-service:8080"),
		FeedServiceURL:         getEnv("FEED_SERVICE_URL", "http://feed-service:8081"),
		SearchServiceURL:       getEnv("SEARCH_SERVICE_URL", "http://search-service:8082"),
		VideoServiceURL:        getEnv("VIDEO_SERVICE_URL", "http://video-service:8083"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8084"),

		UserGRPCAddr:         getEnv("USER_GRPC_ADDR", "user-service:50051"),
		FeedGRPCAddr:         getEnv("FEED_GRPC_ADDR", "feed-service:50052"),
		SearchGRPCAddr:       getEnv("SEARCH_GRPC_ADDR", "search-service:50053"),
		VideoGRPCAddr:        getEnv("VIDEO_GRPC_ADDR", "video-service:50054"),
		NotificationGRPCAddr: getEnv("NOTIFICATION_GRPC_ADDR", "notification-service:50055"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
