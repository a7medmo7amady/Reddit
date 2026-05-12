package realtime

import (
	"encoding/json"
	"sync"
	"time"
)

type Hub struct {
	mu sync.RWMutex

	clientsByUser map[string]map[*Client]bool
	typing        map[string]map[string]time.Time
}

func NewHub() *Hub {
	return &Hub{
		clientsByUser: make(map[string]map[*Client]bool),
		typing:        make(map[string]map[string]time.Time),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clientsByUser[c.UserID] == nil {
		h.clientsByUser[c.UserID] = make(map[*Client]bool)
	}

	h.clientsByUser[c.UserID][c] = true
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clientsByUser[c.UserID] != nil {
		delete(h.clientsByUser[c.UserID], c)
	}

	close(c.Send)
}

func (h *Hub) SendToUser(userID string, payload any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for client := range h.clientsByUser[userID] {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(h.clientsByUser[userID], client)
		}
	}
}

func (h *Hub) SendRawToUser(userID string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clientsByUser[userID] {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(h.clientsByUser[userID], client)
		}
	}
}

func (h *Hub) IsUserConnected(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clientsByUser[userID]) > 0
}

func (h *Hub) SetTyping(conversationID, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.typing[conversationID] == nil {
		h.typing[conversationID] = make(map[string]time.Time)
	}

	h.typing[conversationID][userID] = time.Now().UTC().Add(5 * time.Second)
}

func (h *Hub) ClearExpiredTyping() {
	ticker := time.NewTicker(2 * time.Second)

	go func() {
		for range ticker.C {
			h.mu.Lock()

			now := time.Now().UTC()

			for conversationID, users := range h.typing {
				for userID, expiresAt := range users {
					if now.After(expiresAt) {
						delete(users, userID)
					}
				}

				if len(users) == 0 {
					delete(h.typing, conversationID)
				}
			}

			h.mu.Unlock()
		}
	}()
}
