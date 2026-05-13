package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"notification-service/internal/model"
	"notification-service/internal/repository"
	"notification-service/pkg/email"
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
	emailSvc  email.EmailService
}

func NewNotificationService(redisRepo repository.RedisRepository, hub *websocket.Hub, emailSvc email.EmailService) NotificationService {
	return &notificationService{
		redisRepo: redisRepo,
		hub:       hub,
		emailSvc:  emailSvc,
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
	if err := s.redisRepo.EnqueueOfflineNotification(ctx, n.UserID, n); err != nil {
		return err
	}

	// 4. Send email notification if offline
	// In a real app, fetch email from user-service. For now, using a placeholder.
	userEmail := fmt.Sprintf("%s@example.com", n.UserID) 
	subject := fmt.Sprintf("New Reddit Notification: %s", n.Title)
	body := fmt.Sprintf("Hello,\n\nYou have a new notification on Reddit:\n\n%s\n\nCheck it out here: %s", n.Message, n.Link)
	
	log.Printf("User %s is offline, sending email to %s", n.UserID, userEmail)
	go s.emailSvc.SendNotification(userEmail, subject, body)

	return nil
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
