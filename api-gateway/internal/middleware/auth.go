package middleware

import (
	"net/http"
	"strings"

	jwtpkg "api-gateway/pkg/jwt"

	"github.com/gin-gonic/gin"
)

var jwtSecret string

func SetJWTSecret(secret string) {
	jwtSecret = secret
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed token"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtpkg.Verify(tokenStr, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Request.Header.Set("X-User-Id", claims.UserID)
		c.Request.Header.Set("X-Role", claims.Role)
		c.Next()
	}
}
