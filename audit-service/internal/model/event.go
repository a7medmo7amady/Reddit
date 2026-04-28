package model

type Event struct {
	EventType string
	UserID    string
	Service   string
	Timestamp int64
	Metadata  map[string]interface{}
}
