package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Health(services map[string]string) gin.HandlerFunc {
	client := &http.Client{Timeout: 3 * time.Second}

	return func(c *gin.Context) {
		results := make(map[string]string, len(services))
		degraded := false

		for name, url := range services {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url+"/health", nil)
			resp, err := client.Do(req)
			cancel()

			if err != nil || resp.StatusCode >= 500 {
				results[name] = "unavailable"
				degraded = true
			} else {
				results[name] = "ok"
			}
		}

		gatewayStatus := "ok"
		if degraded {
			gatewayStatus = "degraded"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   gatewayStatus,
			"services": results,
		})
	}
}
