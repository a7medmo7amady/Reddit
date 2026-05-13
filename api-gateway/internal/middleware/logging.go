package middleware

import (
	"time"

	"api-gateway/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Infof("%s %s → %d (%s)\n",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),	
			time.Since(start),
		)
	}
}
