package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")

		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user identity"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
