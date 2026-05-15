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
	"google.golang.org/grpc"
	"net"
	pb "notification-service/pkg/proto/notification"
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
		api.POST("/send", h.SendNotification)
	}

	// Top-level aliases for gateway compatibility
	notif := r.Group("/notifications")
	{
		notif.GET("/recent", h.GetRecent)
		notif.POST("/mark-read", h.MarkRead)
		notif.PATCH("/preferences", h.UpdatePrefs)
		notif.POST("/send", h.SendNotification)
	}

	// WebSocket endpoint (root and gateway-compatible path)
	wsHandler := func(c *gin.Context) {
		userID := c.Query("user_id") // In production, get from auth token
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}
		handleWebSocket(hub, svc, userID, c.Writer, c.Request)
	}
	r.GET("/ws", wsHandler)
	r.GET("/notifications/ws", wsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50055"
	}

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen for gRPC: %v", err)
		}

		s := grpc.NewServer()
		grpcHandler := handler.NewGrpcNotificationHandler(svc)
		pb.RegisterNotificationServiceServer(s, grpcHandler)

		log.Printf("Notification gRPC Service starting on port %s", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	log.Printf("Notification HTTP Service starting on port %s", port)
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
