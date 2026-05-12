package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	MongoURI       string
	MongoDatabase  string
	UserServiceURL string
	RedisAddr      string
	InstanceID     string
	MaxWSConns     int
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8081"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:  getEnv("MONGO_DATABASE", "chat_service"),
		UserServiceURL: getEnv("USER_SERVICE_URL", "http://localhost:8082"),
		RedisAddr:      os.Getenv("REDIS_ADDR"),
		InstanceID:     getEnv("INSTANCE_ID", ""),
		MaxWSConns:     getEnvInt("MAX_WS_CONNS", 10000),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}
