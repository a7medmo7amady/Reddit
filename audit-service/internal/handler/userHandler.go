package handler

import (
	"encoding/json"
	"fmt"

	"audit-service/internal/model"
	"audit-service/internal/service"

	"github.com/IBM/sarama"
)

type UserHandler struct {
	svc *service.AuditService
}

func NewUserHandler(svc *service.AuditService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Handle(msg *sarama.ConsumerMessage) error {
	var event model.UserEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal user event: %w", err)
	}
	return h.svc.HandleUserEvent(msg.Topic, event)
}
