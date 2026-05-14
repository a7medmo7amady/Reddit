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
	// Health check (no auth).
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Routes reached via the API Gateway: gateway proxies /chat/* → chat-service.
	// A request to gateway /chat/inbox arrives here as /chat/inbox.
	chat := r.Group("/chat")
	chat.Use(middleware.Auth())

	chat.GET("/inbox", chatHandler.GetInbox)

	chat.POST("/conversations", chatHandler.CreateConversation)
	chat.POST("/messages", chatHandler.SendMessage)
	chat.DELETE("/messages/:messageId", chatHandler.DeleteMessage)
	chat.POST("/messages/:messageId/report", chatHandler.ReportMessage)

	chat.GET("/communities/:communityId/room", chatHandler.GetOrCreateCommunityRoom)

	chat.GET("/conversations/:conversationId/messages", chatHandler.GetMessages)
	chat.POST("/conversations/:conversationId/read", chatHandler.MarkRead)
	chat.DELETE("/conversations/:conversationId", chatHandler.HideConversation)
	chat.POST("/conversations/:conversationId/hide", chatHandler.HideConversation)
	chat.POST("/conversations/:conversationId/mute", chatHandler.MuteConversation)
	chat.POST("/conversations/:conversationId/unmute", chatHandler.UnmuteConversation)
	chat.POST("/conversations/:conversationId/muted", chatHandler.SetConversationMuted)
	chat.GET("/:room/messages", chatHandler.GetRoomMessagesSince)

	chat.GET("/ws", wsHandler.Connect)

	// Legacy /api prefix (for direct access without the gateway, e.g. local dev).
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
	api.DELETE("/conversations/:conversationId", chatHandler.HideConversation)
	api.POST("/conversations/:conversationId/hide", chatHandler.HideConversation)
	api.POST("/conversations/:conversationId/mute", chatHandler.MuteConversation)
	api.POST("/conversations/:conversationId/unmute", chatHandler.UnmuteConversation)
	api.POST("/conversations/:conversationId/muted", chatHandler.SetConversationMuted)
	api.GET("/chat/:room/messages", chatHandler.GetRoomMessagesSince)
	api.GET("/ws", wsHandler.Connect)
}
