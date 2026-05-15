package handler

import (
	"net/http"
	"search-service/internal/client"
	"search-service/internal/model"
	"search-service/internal/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService   service.SearchService
	videoClient     client.VideoClient
	userClient      client.UserClient
	videoHTTPClient client.VideoHTTPClient
}

func NewSearchHandler(
	searchService service.SearchService,
	videoClient client.VideoClient,
	userClient client.UserClient,
	videoHTTPClient client.VideoHTTPClient,
) *SearchHandler {
	return &SearchHandler{
		searchService:   searchService,
		videoClient:     videoClient,
		userClient:      userClient,
		videoHTTPClient: videoHTTPClient,
	}
}

// Index endpoints
func (h *SearchHandler) IndexPost(c *gin.Context) {
	var post model.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.searchService.IndexPost(c.Request.Context(), post)
	c.JSON(http.StatusOK, gin.H{"message": "Post indexed"})
}

func (h *SearchHandler) IndexCommunity(c *gin.Context) {
	var community model.Community
	if err := c.ShouldBindJSON(&community); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.searchService.IndexCommunity(c.Request.Context(), community)
	c.JSON(http.StatusOK, gin.H{"message": "Community indexed"})
}

func (h *SearchHandler) IndexUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.searchService.IndexUser(c.Request.Context(), user)
	c.JSON(http.StatusOK, gin.H{"message": "User indexed"})
}

func (h *SearchHandler) IndexComment(c *gin.Context) {
	var comment model.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.searchService.IndexComment(c.Request.Context(), comment)
	c.JSON(http.StatusOK, gin.H{"message": "Comment indexed"})
}

// Search endpoint (Universal)
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query 'q' is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	// If user typed "u/" do the searching inside the postgres db (user-service)
	// else search in MongoDB (video-service)
	if strings.HasPrefix(query, "u/") {
		userQuery := strings.TrimPrefix(query, "u/")
		users, total, err := h.userClient.SearchUsers(c.Request.Context(), userQuery, limit, page)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"results": users,
			"total":   total,
			"limit":   limit,
			"page":    page,
			"type":    "users",
			"source":  "postgres",
		})
		return
	}

	// Search in MongoDB via video-service
	result, err := h.videoHTTPClient.Search(c.Request.Context(), query, limit, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": gin.H{
			"posts":    result.Posts.Items,
			"comments": result.Comments.Items,
		},
		"posts_total":    result.Posts.Total,
		"comments_total": result.Comments.Total,
		"limit":          limit,
		"page":           page,
		"type":           "posts",
		"source":         "mongodb",
	})
}

func (h *SearchHandler) Sync(c *gin.Context) {
	posts, total, err := h.videoClient.FetchPosts(c.Request.Context(), 100, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.searchService.SyncPosts(c.Request.Context(), posts)
	c.JSON(http.StatusOK, gin.H{"message": "Sync completed", "synced": len(posts), "total": total})
}
