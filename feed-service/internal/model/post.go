package model

type Image struct {
	Thumbnail    string `json:"thumbnail"`
	Preview      string `json:"preview"`
	Full         string `json:"full"`
	OriginalName string `json:"originalName,omitempty"`
}

type Video struct {
	Status      string `json:"status"`
	PlaybackUrl string `json:"playbackUrl,omitempty"`
}

type Post struct {
	ID           int     `json:"id,omitempty"`
	StringID     string  `json:"stringId,omitempty"`
	Title        string  `json:"title"`
	Body         string  `json:"body"`
	Community    string  `json:"community"`
	Author       string  `json:"author"`
	Type         string  `json:"type,omitempty"`
	Upvotes      int     `json:"upvotes"`
	Downvotes    int     `json:"downvotes"`
	Score        int     `json:"score"`
	CommentCount int     `json:"commentCount"`
	CreatedAt    string  `json:"createdAt"`
	Images       []Image `json:"images,omitempty"`
	Video        *Video  `json:"video,omitempty"`
}
