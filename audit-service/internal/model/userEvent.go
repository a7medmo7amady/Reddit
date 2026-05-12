package model

const (
	UserProfileUpdated = "user.profile_updated"
	UserAccountDeleted = "user.account_deleted"
)

type UserEvent struct {
	UserID    string            
	Action    string            
	Service   string            
	Timestamp int64             
	Metadata  map[string]interface{}
}
