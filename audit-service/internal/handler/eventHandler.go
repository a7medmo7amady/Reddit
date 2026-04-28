package handler

import (
	"audit-service/internal/model"
	"audit-service/internal/service"
)

type EventHandler struct {
	service *service.AuditService
}

func NewEventHandler(s *service.AuditService) *EventHandler {
	return &EventHandler{service: s}
}

func (h *EventHandler) Handle(event model.Event) {
	h.service.HandleEvent(event)
}
