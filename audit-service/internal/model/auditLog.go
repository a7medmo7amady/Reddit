package model

import "time"

type AuditLog struct {
	UserID    string
	Action    string
	Service   string
	Topic     string
	Timestamp time.Time
	Metadata  map[string]interface{}
}
