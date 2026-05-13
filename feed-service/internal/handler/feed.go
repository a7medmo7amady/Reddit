package handler

import (
	"feed-service/internal/cache"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func Trending(tc *cache.TrendingCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, err := tc.Get(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch trending posts"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"posts": posts, "total": len(posts)})
	}
}

func CommunityFeed(pc *cache.PostCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		community := c.Param("name")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
		if limit <= 0 || limit > 100 {
			limit = 25
		}

		posts, err := pc.GetByCommunity(c.Request.Context(), community, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch posts"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"community": community, "posts": posts, "total": len(posts)})
	}
}

func UserFeed(pc *cache.PostCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.Query("communities")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if limit <= 0 || limit > 100 {
			limit = 50
		}

		var communities []string
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				communities = append(communities, s)
			}
		}

		if len(communities) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "communities query param required"})
			return
		}

		posts, err := pc.GetByCommunities(c.Request.Context(), communities, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch feed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"posts": posts, "total": len(posts)})
	}
}
