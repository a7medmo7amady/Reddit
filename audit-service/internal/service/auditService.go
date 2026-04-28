package service

import (
	"fmt"
	"time"

	"audit-service/internal/model"
	"audit-service/internal/repository"
)

type AuditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) HandleAuthEvent(topic string, event model.AuthEvent) error {
	if event.UserID == "" || event.Action == "" {
		return fmt.Errorf("invalid auth event: user_id and action are required")
	}
	return s.repo.Save(model.AuditLog{
		UserID:    event.UserID,
		Action:    event.Action,
		Service:   event.Service,
		Topic:     topic,
		Timestamp: time.Unix(event.Timestamp, 0),
		Metadata:  event.Metadata,
	})
}

func (s *AuditService) HandleUserEvent(topic string, event model.UserEvent) error {
	if event.UserID == "" || event.Action == "" {
		return fmt.Errorf("invalid user event: user_id and action are required")
	}
	return s.repo.Save(model.AuditLog{
		UserID:    event.UserID,
		Action:    event.Action,
		Service:   event.Service,
		Topic:     topic,
		Timestamp: time.Unix(event.Timestamp, 0),
		Metadata:  event.Metadata,
	})
}
