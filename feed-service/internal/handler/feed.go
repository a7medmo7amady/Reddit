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

func CommunityFeed(pc *cache.PostCache, bc *cache.BanCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		community := c.Param("name")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
		if limit <= 0 || limit > 100 {
			limit = 25
		}
		userID := c.GetHeader("X-User-Id")
		if userID != "" && bc != nil {
			banned, err := bc.IsBanned(c.Request.Context(), userID, community)
			if err == nil && banned {
				c.JSON(http.StatusForbidden, gin.H{"error": "you are banned from this community"})
				return
			}
		}

		posts, err := pc.GetByCommunity(c.Request.Context(), community, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch posts"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"community": community, "posts": posts, "total": len(posts)})
	}
}

func UserFeed(pc *cache.PostCache, bc *cache.BanCache) gin.HandlerFunc {
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

		userID := c.GetHeader("X-User-Id")
		if userID != "" && bc != nil {
			var allowed []string
			for _, comm := range communities {
				banned, err := bc.IsBanned(c.Request.Context(), userID, comm)
				if err != nil || !banned {
					allowed = append(allowed, comm)
				}
			}
			communities = allowed
		}

		if len(communities) == 0 {
			c.JSON(http.StatusOK, gin.H{"posts": []interface{}{}, "total": 0})
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

