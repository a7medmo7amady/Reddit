package main

import (
	"context"
	"log"
	"os"

	"audit-service/internal/consumer"
	"audit-service/internal/handler"
	"audit-service/internal/repository"
	"audit-service/internal/service"
	kafkacfg "audit-service/pkg/kafka"

	"github.com/goccy/go-yaml"
)

func main() {
	cfg := loadConfig("config/config.yaml")

	repo := repository.NewAuditRepository()
	svc := service.NewAuditService(repo)
	h := handler.NewEventHandler(svc)

	c, err := consumer.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topics, h)
	if err != nil {
		log.Fatal(err)
	}

	c.Start(context.Background())
}

func loadConfig(path string) kafkacfg.Config {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open config: %v", err)
	}
	defer f.Close()

	var cfg kafkacfg.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatalf("decode config: %v", err)
	}
	return cfg
}
