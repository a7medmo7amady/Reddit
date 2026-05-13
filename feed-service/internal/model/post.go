package model

type Post struct {
	ID           int    `json:"id,omitempty"`
	StringID     string `json:"stringId,omitempty"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	Community    string `json:"community"`
	Author       string `json:"author"`
	Score        int    `json:"score"`
	CommentCount int    `json:"commentCount"`
	CreatedAt    string `json:"createdAt"`
}
