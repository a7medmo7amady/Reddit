package dto

import "time"

type InboxItem struct {
	ConversationID string        `json:"conversationId"`
	Type           string        `json:"type"`
	CommunityID    string        `json:"communityId,omitempty"`
	LastMessage    *InboxMessage `json:"lastMessage,omitempty"`
	UnreadCount    int           `json:"unreadCount"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type InboxMessage struct {
	ID        string    `json:"id"`
	SenderID  string    `json:"senderId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
