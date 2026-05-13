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
		tokenStr := bearerToken(c.GetHeader("Authorization"))
		if tokenStr == "" {
			tokenStr = c.Query("access_token")
		}
		if tokenStr == "" {
			if cookie, err := c.Cookie("access_token"); err == nil {
				tokenStr = cookie
			}
		}
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed token"})
			return
		}

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

func bearerToken(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}
