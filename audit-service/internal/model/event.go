package model

const (
	AuthUserRegistered  = "user.registered"
	AuthUserLoggedIn    = "user.logged_in"
	AuthUserLoggedOut   = "user.logged_out"
	AuthPasswordChanged = "user.password_changed"
)

type AuthEvent struct {
	UserID    string            
	Action    string            
	Service   string            
	Timestamp int64             
	Metadata  map[string]interface{}
}
