package handler

import (
	"net/http"
	"search-service/internal/client"
	"search-service/internal/model"
	"search-service/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService service.SearchService
	videoClient   client.VideoClient
}

func NewSearchHandler(searchService service.SearchService, videoClient client.VideoClient) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		videoClient:   videoClient,
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

	searchType := c.DefaultQuery("type", "posts")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	var results interface{}
	var total int64
	var err error

	switch searchType {
	case "posts":
		filters := map[string]interface{}{
			"type":         c.Query("content_type"), // text, image, video
			"community_id": c.Query("community_id"), // scoped search
			"date_range":   c.Query("date_range"),   // week, month
		}
		results, total, err = h.searchService.SearchPosts(c.Request.Context(), query, filters, limit, page)
	case "communities":
		results, total, err = h.searchService.SearchCommunities(c.Request.Context(), query, limit, page)
	case "users":
		results, total, err = h.searchService.SearchUsers(c.Request.Context(), query, limit, page)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   total,
		"limit":   limit,
		"page":    page,
		"type":    searchType,
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
