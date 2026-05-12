package http

import (
	"chat-service/internal/http/handlers"
	"chat-service/internal/http/middleware"
	"chat-service/internal/realtime"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	chatHandler *handlers.ChatHandler,
	wsHandler *realtime.Handler,
) {
	api := r.Group("/api")
	api.Use(middleware.Auth())

	api.GET("/inbox", chatHandler.GetInbox)

	api.POST("/conversations", chatHandler.CreateConversation)
	api.POST("/messages", chatHandler.SendMessage)
	api.DELETE("/messages/:messageId", chatHandler.DeleteMessage)
	api.POST("/messages/:messageId/report", chatHandler.ReportMessage)

	api.GET("/communities/:communityId/room", chatHandler.GetOrCreateCommunityRoom)

	api.GET("/conversations/:conversationId/messages", chatHandler.GetMessages)
	api.POST("/conversations/:conversationId/read", chatHandler.MarkRead)
	api.GET("/chat/:room/messages", chatHandler.GetRoomMessagesSince)

	api.GET("/ws", wsHandler.Connect)

	// Spec/ADR path used by reconnect flows: GET /chat/{room}/messages?since={timestamp}
	chat := r.Group("/chat")
	chat.Use(middleware.Auth())
	chat.GET("/:room/messages", chatHandler.GetRoomMessagesSince)
}
