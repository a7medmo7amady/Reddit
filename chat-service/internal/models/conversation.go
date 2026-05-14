package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conversation struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type        string             `bson:"type" json:"type"` // direct, community
	CommunityID string             `bson:"communityId,omitempty" json:"communityId,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type ConversationParticipant struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID primitive.ObjectID `bson:"conversationId" json:"conversationId"`
	UserID         string             `bson:"userId" json:"userId"`
	JoinedAt       time.Time          `bson:"joinedAt" json:"joinedAt"`
	LastReadAt     *time.Time         `bson:"lastReadAt,omitempty" json:"lastReadAt,omitempty"`
	Muted          bool               `bson:"muted" json:"muted"`
	HiddenAt       *time.Time         `bson:"hiddenAt,omitempty" json:"hiddenAt,omitempty"`
}
