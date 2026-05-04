package main

import (
	"api-gateway/config"
	"api-gateway/internal/middleware"
	"api-gateway/internal/proxy"
	"api-gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	middleware.SetJWTSecret(cfg.JWTSecret)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logging())
	r.Use(middleware.CORS())

	r.Any("/auth/*path", gin.WrapH(proxy.NewSingle(cfg.UserServiceURL)))

	
	protected := r.Group("/")
	protected.Use(middleware.Auth())
	{
		protected.Any("/users/*path", gin.WrapH(proxy.NewSingle(cfg.UserServiceURL)))
		protected.Any("/posts/*path", gin.WrapH(proxy.NewSingle(cfg.FeedServiceURL)))
		protected.Any("/search/*path", gin.WrapH(proxy.NewSingle(cfg.SearchServiceURL)))
		protected.Any("/video/*path", gin.WrapH(proxy.NewSingle(cfg.VideoServiceURL)))
		protected.Any("/notifications/*path", gin.WrapH(proxy.NewSingle(cfg.NotificationServiceURL)))
	}

	logger.Infof("API Gateway listening on :%s\n", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		logger.Fatalf("server error: %v\n", err)
	}
}
