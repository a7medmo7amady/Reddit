package main

import (
	"context"
	"os"

	"audit-service/internal/consumer"
	"audit-service/internal/handler"
	"audit-service/internal/repository"
	"audit-service/internal/service"
	"audit-service/pkg/kafka"
	"audit-service/pkg/logger"

	"github.com/goccy/go-yaml"
)

func main() {
	cfg := loadConfig("config/config.yaml")

	repo := repository.NewAuditRepository()
	svc := service.NewAuditService(repo)

	dispatcher := handler.NewDispatcher()
	dispatcher.Register(kafka.TopicAuth, handler.NewAuthHandler(svc))
	dispatcher.Register(kafka.TopicUser, handler.NewUserHandler(svc))

	c, err := consumer.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topics, dispatcher)
	if err != nil {
		logger.Log.Error("failed to create kafka consumer", "error", err)
		os.Exit(1)
	}

	logger.Log.Info("audit service started", "topics", cfg.Kafka.Topics)
	c.Start(context.Background())
}

func loadConfig(path string) kafka.Config {
	f, err := os.Open(path)
	if err != nil {
		logger.Log.Error("open config failed", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	var cfg kafka.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Log.Error("decode config failed", "error", err)
		os.Exit(1)
	}
	return cfg
}
