package realtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub        *Hub
	dispatcher *Dispatcher
	limiter    chan struct{}
}

func NewHandler(hub *Hub, dispatcher *Dispatcher, maxConns int) *Handler {
	var limiter chan struct{}
	if maxConns > 0 {
		limiter = make(chan struct{}, maxConns)
	}
	return &Handler{hub: hub, dispatcher: dispatcher, limiter: limiter}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate allowed origins.
		return true
	},
}

func (h *Handler) Connect(c *gin.Context) {
	userID := c.GetString("userID")

	if h.limiter != nil {
		select {
		case h.limiter <- struct{}{}:
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "websocket capacity reached"})
			return
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		if h.limiter != nil {
			<-h.limiter
		}
		return
	}

	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
		OnClose: func() {
			if h.limiter != nil {
				select {
				case <-h.limiter:
				default:
				}
			}
		},
	}

	h.hub.Register(client)
	if h.dispatcher != nil {
		h.dispatcher.DrainOffline(c.Request.Context(), userID)
	}

	go client.WritePump()
	go client.ReadPump()
}
