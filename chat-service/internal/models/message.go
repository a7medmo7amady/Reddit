package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID primitive.ObjectID `bson:"conversationId" json:"conversationId"`
	SenderID       string             `bson:"senderId" json:"senderId"`
	Content        string             `bson:"content" json:"content"`
	Type           string             `bson:"type" json:"type"` // text, image, file
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	EditedAt       *time.Time         `bson:"editedAt,omitempty" json:"editedAt,omitempty"`
	DeletedAt      *time.Time         `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
	DeletedBy      *string            `bson:"deletedBy,omitempty" json:"deletedBy,omitempty"`
}
