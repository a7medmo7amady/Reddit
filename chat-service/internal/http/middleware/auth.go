package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Auth trusts the identity headers set by the API Gateway after JWT validation.
// The gateway sets X-User-Id and X-Role on every authenticated request.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-Id")

		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user identity"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("userRole", c.GetHeader("X-Role"))
		c.Next()
	}
}
