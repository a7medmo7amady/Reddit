package model

import "time"

type NotificationType string

const (
	TypeReply      NotificationType = "REPLY"
	TypeMention    NotificationType = "MENTION"
	TypeDirectMsg  NotificationType = "DIRECT_MESSAGE"
	TypeModAction  NotificationType = "MOD_ACTION"
)

type Notification struct {
	ID        string           `json:"id" bson:"_id"`
	UserID    string           `json:"user_id" bson:"user_id"`
	Type      NotificationType `json:"type" bson:"type"`
	Title     string           `json:"title" bson:"title"`
	Message   string           `json:"message" bson:"message"`
	IsRead    bool             `json:"is_read" bson:"is_read"`
	Link      string           `json:"link,omitempty" bson:"link,omitempty"`
	CreatedAt time.Time        `json:"created_at" bson:"created_at"`
}

type NotificationPreference struct {
	UserID       string   `json:"user_id" bson:"user_id"`
	InAppEnabled bool     `json:"in_app_enabled" bson:"in_app_enabled"`
	EmailEnabled bool     `json:"email_enabled" bson:"email_enabled"`
	PushEnabled  bool     `json:"push_enabled" bson:"push_enabled"`
	MutedTypes   []string `json:"muted_types" bson:"muted_types"`
	LastOnlineAt time.Time `json:"last_online_at" bson:"last_online_at"`
}
