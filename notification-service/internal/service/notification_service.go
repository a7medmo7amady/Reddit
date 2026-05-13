package service

import (
	"context"
	"encoding/json"
	"log"
	"notification-service/internal/model"
	"notification-service/internal/repository"
	"notification-service/pkg/websocket"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, n *model.Notification) error
	GetRecentNotifications(ctx context.Context, userID string, limit int) ([]model.Notification, error)
	MarkAllAsRead(ctx context.Context, userID string) error
	UpdatePreference(ctx context.Context, userID string, pref model.NotificationPreference) error
	GetPreference(ctx context.Context, userID string) (*model.NotificationPreference, error)
	DeliverOfflineNotifications(ctx context.Context, userID string) error
}

type notificationService struct {
	redisRepo repository.RedisRepository
	hub       *websocket.Hub
}

func NewNotificationService(redisRepo repository.RedisRepository, hub *websocket.Hub) NotificationService {
	return &notificationService{
		redisRepo: redisRepo,
		hub:       hub,
	}
}

func (s *notificationService) CreateNotification(ctx context.Context, n *model.Notification) error {
	// 1. Check if user is online
	if s.hub.IsUserConnected(n.UserID) {
		// 2. If online, push via WebSocket
		data, _ := json.Marshal(n)
		if s.hub.SendToUser(n.UserID, data) {
			return nil
		}
	}

	// 3. If offline or WebSocket push failed, queue in Redis
	log.Printf("User %s is offline, queuing notification in Redis", n.UserID)
	return s.redisRepo.EnqueueOfflineNotification(ctx, n.UserID, n)
}

func (s *notificationService) DeliverOfflineNotifications(ctx context.Context, userID string) error {
	notifications, err := s.redisRepo.GetOfflineNotifications(ctx, userID)
	if err != nil {
		return err
	}

	if len(notifications) == 0 {
		return nil
	}

	log.Printf("Delivering %d offline notifications to user %s", len(notifications), userID)
	for _, n := range notifications {
		data, _ := json.Marshal(n)
		s.hub.SendToUser(userID, data)
	}

	// Clear once delivered
	return s.redisRepo.DeleteOfflineNotifications(ctx, userID)
}

func (s *notificationService) GetRecentNotifications(ctx context.Context, userID string, limit int) ([]model.Notification, error) {
	return nil, nil
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return nil
}

func (s *notificationService) UpdatePreference(ctx context.Context, userID string, pref model.NotificationPreference) error {
	return nil
}

func (s *notificationService) GetPreference(ctx context.Context, userID string) (*model.NotificationPreference, error) {
	return nil, nil
}
