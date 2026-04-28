package handler

import (
	"encoding/json"
	"fmt"

	"audit-service/internal/model"
	"audit-service/internal/service"

	"github.com/IBM/sarama"
)

type AuthHandler struct {
	svc *service.AuditService
}

func NewAuthHandler(svc *service.AuditService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Handle(msg *sarama.ConsumerMessage) error {
	var event model.AuthEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal auth event: %w", err)
	}
	return h.svc.HandleAuthEvent(msg.Topic, event)
}
