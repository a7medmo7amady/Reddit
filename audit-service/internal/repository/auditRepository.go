package repository

import (
	"audit-service/internal/model"
	"audit-service/pkg/logger"
)

type AuditRepository interface {
	Save(log model.AuditLog) error
}

type auditRepository struct{}

func NewAuditRepository() AuditRepository {
	return &auditRepository{}
}

func (r *auditRepository) Save(log model.AuditLog) error {
	logger.Log.Info("audit log",
		"user_id", log.UserID,
		"action", log.Action,
		"service", log.Service,
		"topic", log.Topic,
		"timestamp", log.Timestamp,
	)
	return nil
}
