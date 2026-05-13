package realtime

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 70 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 8192
)

type Client struct {
	UserID         string
	Conn           *websocket.Conn
	Send           chan []byte
	Hub            *Hub
	Context        context.Context
	OnClose        func()
	TypingNotifier TypingNotifier
}

type TypingNotifier interface {
	NotifyTyping(ctx context.Context, userID, conversationID string) error
}

type clientEvent struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversationId"`
}

func (c *Client) ReadPump() {
	defer func() {
		if c.OnClose != nil {
			c.OnClose()
		}
		c.Hub.Unregister(c)
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) handleMessage(message []byte) {
	var event clientEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("invalid ws message from user=%s: %v", c.UserID, err)
		return
	}

	switch event.Type {
	case "chat.typing", "typing":
		if c.TypingNotifier == nil {
			return
		}
		ctx := c.Context
		if ctx == nil {
			ctx = context.Background()
		}
		if err := c.TypingNotifier.NotifyTyping(ctx, c.UserID, event.ConversationID); err != nil {
			log.Printf("typing event rejected for user=%s conversation=%s: %v", c.UserID, event.ConversationID, err)
		}
	default:
		log.Printf("unknown ws message type from user=%s type=%s", c.UserID, event.Type)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
