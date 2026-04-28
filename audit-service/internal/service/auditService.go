package service

import (
	"audit-service/internal/model"
	"audit-service/internal/repository"
	"fmt"
)

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) HandleEvent(event model.Event) {
	log := fmt.Sprintf("User %s did %s", event.UserID, event.EventType)
	s.repo.Save(log)
}
