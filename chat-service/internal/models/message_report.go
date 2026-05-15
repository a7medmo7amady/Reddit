package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageReport struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MessageID  primitive.ObjectID `bson:"messageId" json:"messageId"`
	ReporterID string             `bson:"reporterId" json:"reporterId"`
	Reason     string             `bson:"reason" json:"reason"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}
