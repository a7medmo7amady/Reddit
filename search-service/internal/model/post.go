package model

import "time"

type VideoInfo struct {
	Status      string   `json:"status"`
	S3Key       string   `json:"s3_key,omitempty"`
	PlaybackURL string   `json:"playback_url,omitempty"`
	Resolutions []string `json:"resolutions,omitempty"`
}

type ImageInfo struct {
	URL    string `json:"url"`
	S3Key  string `json:"s3_key"`
	Format string `json:"format"`
}

type Post struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	AuthorID    string    `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	CommunityID string    `json:"community_id"`
	CommunityName string  `json:"community_name"`
	Type        string    `json:"type"` // text, image, video
	URL         string    `json:"url,omitempty"`
	Flair       string    `json:"flair,omitempty"`
	NSFW        bool      `json:"nsfw"`
	Spoiler     bool      `json:"spoiler"`
	OC          bool      `json:"oc"`
	Images      []ImageInfo `json:"images,omitempty"`
	Video       *VideoInfo  `json:"video,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

type Community struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Karma       int       `json:"karma"`
	CreatedAt   time.Time `json:"created_at"`
}

type Comment struct {
	ID          string    `json:"id"`
	PostID      string    `json:"post_id"`
	AuthorID    string    `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}
