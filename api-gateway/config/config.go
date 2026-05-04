package config

import "os"

type Config struct {
	Port                string
	JWTSecret           string
	UserServiceURL      string
	FeedServiceURL      string
	SearchServiceURL    string
	VideoServiceURL     string
	NotificationServiceURL string
}

func Load() *Config {
	return &Config{
		Port:                   getEnv("PORT", "8088"),
		JWTSecret:              getEnv("JWT_SECRET", "changeme"),
		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://user-service:8080"),
		FeedServiceURL:         getEnv("FEED_SERVICE_URL", "http://feed-service:8081"),
		SearchServiceURL:       getEnv("SEARCH_SERVICE_URL", "http://search-service:8082"),
		VideoServiceURL:        getEnv("VIDEO_SERVICE_URL", "http://video-service:8083"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8084"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
