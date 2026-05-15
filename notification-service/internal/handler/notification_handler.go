package handler

import (
	"net/http"
	"notification-service/internal/model"
	"notification-service/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(service service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) GetRecent(c *gin.Context) {
	userID := c.Request.Header.Get("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	notifications, err := h.service.GetRecentNotifications(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := c.Request.Header.Get("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	if err := h.service.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All marked as read"})
}

func (h *NotificationHandler) UpdatePrefs(c *gin.Context) {
	// ... update preferences
	c.JSON(http.StatusOK, gin.H{"message": "Preferences updated"})
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req struct {
		UserID  string `json:"user_id" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
		Link    string `json:"link"`
		Type    string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notificationType := model.NotificationType(req.Type)
	if notificationType == "" {
		notificationType = model.TypeReply
	}

	n := &model.Notification{
		UserID:  req.UserID,
		Title:   req.Title,
		Message: req.Message,
		Link:    req.Link,
		Type:    notificationType,
	}

	if err := h.service.CreateNotification(c.Request.Context(), n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent"})
}
