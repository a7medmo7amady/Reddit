package main

import (
	"audit-service/internal/consumer"
	"audit-service/internal/handler"
	"audit-service/internal/repository"
	"audit-service/internal/service"
)

func main() {
	repo := repository.NewAuditRepository()
	svc := service.NewAuditService(repo)
	handler := handler.NewEventHandler(svc)
	consumer := consumer.NewKafkaConsumer(handler)

	consumer.Start()
}
