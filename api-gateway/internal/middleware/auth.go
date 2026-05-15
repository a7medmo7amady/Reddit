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

// Auth enforces a valid JWT — aborts with 401 if absent/invalid.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed token"})
			return
		}

		claims, err := jwtpkg.Verify(tokenStr, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		userID := claims.UserID
		if userID == "" {
			userID = claims.Subject
		}
		c.Request.Header.Set("X-User-Id", userID)
		c.Request.Header.Set("X-Username", claims.Username)
		c.Request.Header.Set("X-Role", claims.Role)
		c.Next()
	}
}

// OptionalAuth injects X-User-Id / X-Username when a valid JWT is present,
// but lets unauthenticated requests pass through. Use on public routes that
// still need ban/personalisation checks for logged-in users.
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr != "" {
			if claims, err := jwtpkg.Verify(tokenStr, jwtSecret); err == nil {
				userID := claims.UserID
				if userID == "" {
					userID = claims.Subject
				}
				c.Request.Header.Set("X-User-Id", userID)
				c.Request.Header.Set("X-Username", claims.Username)
				c.Request.Header.Set("X-Role", claims.Role)
			}
		}
		c.Next()
	}
}

// extractToken pulls the bearer token from Authorization header, query param, or cookie.
func extractToken(c *gin.Context) string {
	if t := bearerToken(c.GetHeader("Authorization")); t != "" {
		return t
	}
	if t := c.Query("access_token"); t != "" {
		return t
	}
	if cookie, err := c.Cookie("access_token"); err == nil {
		return cookie
	}
	return ""
}

func bearerToken(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}
