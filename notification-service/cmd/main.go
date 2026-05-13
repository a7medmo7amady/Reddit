package main

import (
	"context"
	"log"
	"net/http"
	"notification-service/internal/handler"
	"notification-service/internal/repository"
	"notification-service/internal/service"
	"notification-service/pkg/email"
	"notification-service/pkg/websocket"
	"os"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For development
	},
}

func main() {
	ctx := context.Background()

	// Initialize Redis
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	redisRepo := repository.NewRedisRepository(rdb)

	// Initialize Email Service
	emailSvc := email.NewEmailService()

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize services and handlers
	svc := service.NewNotificationService(redisRepo, hub, emailSvc)
	h := handler.NewNotificationHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API routes
	api := r.Group("/api/v1/notifications")
	{
		api.GET("/recent", h.GetRecent)
		api.POST("/mark-read", h.MarkRead)
		api.PATCH("/preferences", h.UpdatePrefs)
	}

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		userID := c.Query("user_id") // In production, get from auth token
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}
		handleWebSocket(hub, svc, userID, c.Writer, c.Request)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	log.Printf("Notification Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func handleWebSocket(hub *websocket.Hub, svc service.NotificationService, userID string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket for user %s: %v", userID, err)
		return
	}

	client := &websocket.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	// Register client
	hub.Register <- client

	// Deliver offline notifications
	go func() {
		if err := svc.DeliverOfflineNotifications(context.Background(), userID); err != nil {
			log.Printf("Failed to deliver offline notifications for user %s: %v", userID, err)
		}
	}()

	// Start pumps
	go client.WritePump()
	go client.ReadPump(hub)
}
