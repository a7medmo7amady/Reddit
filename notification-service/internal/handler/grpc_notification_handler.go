package handler

import (
	"context"
	"notification-service/internal/model"
	"notification-service/internal/service"
	pb "notification-service/pkg/proto/notification"
)

type GrpcNotificationHandler struct {
	pb.UnimplementedNotificationServiceServer
	svc service.NotificationService
}

func NewGrpcNotificationHandler(svc service.NotificationService) *GrpcNotificationHandler {
	return &GrpcNotificationHandler{
		svc: svc,
	}
}

func (h *GrpcNotificationHandler) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	n := &model.Notification{
		UserID:  req.UserId,
		Title:   req.Title,
		Message: req.Message,
		Link:    req.Link,
		Type:    req.Type,
	}

	err := h.svc.CreateNotification(ctx, n)
	if err != nil {
		return &pb.SendNotificationResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SendNotificationResponse{
		Success: true,
	}, nil
}

func (h *GrpcNotificationHandler) GetUserPreferences(ctx context.Context, req *pb.GetUserPreferencesRequest) (*pb.GetUserPreferencesResponse, error) {
	prefs, err := h.svc.GetPreference(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserPreferencesResponse{
		EmailEnabled: prefs.EmailEnabled,
		PushEnabled:  prefs.PushEnabled,
		MutedTypes:   prefs.MutedTypes,
	}, nil
}

func (h *GrpcNotificationHandler) UpdateUserPreferences(ctx context.Context, req *pb.UpdateUserPreferencesRequest) (*pb.UpdateUserPreferencesResponse, error) {
	pref := model.NotificationPreference{
		EmailEnabled: req.EmailEnabled,
		PushEnabled:  req.PushEnabled,
		MutedTypes:   req.MutedTypes,
	}

	err := h.svc.UpdatePreference(ctx, req.UserId, pref)
	if err != nil {
		return &pb.UpdateUserPreferencesResponse{Success: false}, err
	}

	return &pb.UpdateUserPreferencesResponse{Success: true}, nil
}
