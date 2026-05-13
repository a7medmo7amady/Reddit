package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chat-service/internal/dto"
	"chat-service/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chat *service.ChatService
}

func NewChatHandler(chat *service.ChatService) *ChatHandler {
	return &ChatHandler{chat: chat}
}

func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID := c.GetString("userID")

	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := h.chat.CreateConversation(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conversation)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("userID")

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, queued, err := h.chat.SendMessage(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if queued {
		c.JSON(http.StatusAccepted, msg)
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID := c.GetString("userID")
	conversationID := c.Param("conversationId")

	messages, err := h.chat.GetConversationMessages(c.Request.Context(), userID, conversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *ChatHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("userID")
	conversationID := c.Param("conversationId")

	if err := h.chat.MarkRead(c.Request.Context(), userID, conversationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ChatHandler) GetInbox(c *gin.Context) {
	userID := c.GetString("userID")

	inbox, err := h.chat.GetInbox(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inbox)
}

func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	userID := c.GetString("userID")
	messageID := c.Param("messageId")

	role := strings.ToUpper(strings.TrimSpace(c.GetString("userRole")))
	isModerator := role == "MODERATOR" || role == "ADMIN"

	if err := h.chat.DeleteMessage(c.Request.Context(), userID, messageID, isModerator); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ChatHandler) ReportMessage(c *gin.Context) {
	userID := c.GetString("userID")
	messageID := c.Param("messageId")

	var req dto.ReportMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.chat.ReportMessage(c.Request.Context(), userID, messageID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ChatHandler) GetOrCreateCommunityRoom(c *gin.Context) {
	userID := c.GetString("userID")
	communityID := c.Param("communityId")

	conv, err := h.chat.GetOrCreateCommunityRoom(c.Request.Context(), userID, communityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conv)
}

func (h *ChatHandler) GetRoomMessagesSince(c *gin.Context) {
	userID := c.GetString("userID")
	room := c.Param("room")

	sinceStr := strings.TrimSpace(c.Query("since"))
	var since *time.Time
	if sinceStr != "" {
		t, err := parseSince(sinceStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		since = &t
	}

	messages, err := h.chat.GetRoomMessagesSince(c.Request.Context(), userID, room, since)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func parseSince(v string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t.UTC(), nil
	}
	// unix millis or seconds
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return time.Time{}, errors.New("invalid since timestamp")
	}
	if n > 1_000_000_000_000 {
		return time.UnixMilli(n).UTC(), nil
	}
	return time.Unix(n, 0).UTC(), nil
}
