package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/middleware"
	"api-gateway/internal/proxy"
	consulpkg "api-gateway/pkg/consul"
	"api-gateway/pkg/logger"
	rlpkg "api-gateway/pkg/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	cfg := config.Load()
	middleware.SetJWTSecret(cfg.JWTSecret)
	middleware.SetAllowedOrigin(cfg.AllowedOrigin)

	if cfg.RedisAddr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
		if err := rlpkg.Ping(rdb); err != nil {
			logger.Warnf("Redis unreachable (%v), falling back to in-memory rate limiter\n", err)
		} else {
			middleware.SetRedisLimiter(rlpkg.New(rdb, time.Second, 30))
			logger.Infof("Redis rate limiter enabled at %s\n", cfg.RedisAddr)
		}
	}

	resolve := staticResolver(cfg)
	if cfg.ConsulAddr != "" {
		r, err := consulpkg.New(cfg.ConsulAddr)
		if err != nil {
			logger.Fatalf("consul init: %v\n", err)
		}
		resolve = consulResolver(cfg, r)
		logger.Infof("Consul service discovery enabled at %s\n", cfg.ConsulAddr)
	}

	httpSrv := buildHTTPServer(cfg, resolve)

	grpcSrv := buildGRPCServer(cfg)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Infof("HTTP gateway listening on :%s\n", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server: %v\n", err)
		}
	}()
	go func() {
		logger.Infof("gRPC gateway listening on :%s\n", cfg.GRPCPort)
		if err := grpcSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("gRPC server: %v\n", err)
		}
	}()

	<-quit
	logger.Infof("Shutting down...\n")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpSrv.Shutdown(ctx)
	grpcSrv.Shutdown(ctx)
}
func buildHTTPServer(cfg *config.Config, resolve func(string) string) *http.Server {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logging())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found: " + c.Request.Method + " " + c.Request.URL.Path})
	})

	r.GET("/health", handler.Health(map[string]string{
		"user":         resolve("user"),
		"feed":         resolve("feed"),
		"search":       resolve("search"),
		"video":        resolve("video"),
		"notification": resolve("notification"),
		"chat":         resolve("chat"),
	}))

	r.Any("/auth/*path", gin.WrapH(proxy.NewSingle(resolve("user"))))
	r.Any("/oauth2/*path", gin.WrapH(proxy.NewSingle(resolve("user"))))
	r.Any("/login/oauth2/*path", gin.WrapH(proxy.NewSingle(resolve("user"))))
	r.Any("/users/*path", gin.WrapH(proxy.NewSingle(resolve("user"))))
	r.GET("/posts/trending", middleware.OptionalAuth(), gin.WrapH(proxy.NewSingle(resolve("feed"))))
	r.GET("/posts/feed", middleware.OptionalAuth(), gin.WrapH(proxy.NewSingle(resolve("feed"))))
	r.GET("/posts/community/*path", middleware.OptionalAuth(), gin.WrapH(proxy.NewSingle(resolve("feed"))))
	r.GET("/posts", middleware.OptionalAuth(), gin.WrapH(proxy.NewSingle(resolve("video"))))
	r.GET("/posts/:id", middleware.OptionalAuth(), gin.WrapH(proxy.NewSingle(resolve("video"))))
	r.GET("/posts/:id/status", gin.WrapH(proxy.NewSingle(resolve("video"))))
	r.GET("/posts/:id/history", gin.WrapH(proxy.NewSingle(resolve("video"))))
	r.GET("/posts/:id/comments", gin.WrapH(proxy.NewSingle(resolve("video"))))
	r.GET("/comments", gin.WrapH(proxy.NewSingle(resolve("video"))))
	r.GET("/search", middleware.OptionalAuth(), gin.WrapH(proxy.NewSingle(resolve("search"))))

	r.GET("/communities/:name", middleware.OptionalAuth(), gin.WrapH(proxy.NewSingle(resolve("user"))))

	r.GET("/assets/*path", gin.WrapH(proxy.NewSingle(resolve("video"))))

	protected := r.Group("/")
	protected.Use(middleware.Auth())
	{
		protected.POST("/posts", gin.WrapH(proxy.NewSingle(resolve("video"))))

		protected.POST("/posts/:id/vote", gin.WrapH(proxy.NewSingle(resolve("video"))))
		protected.PATCH("/posts/:id", gin.WrapH(proxy.NewSingle(resolve("video"))))
		protected.DELETE("/posts/:id", gin.WrapH(proxy.NewSingle(resolve("video"))))

		protected.POST("/posts/:id/comments", gin.WrapH(proxy.NewSingle(resolve("video"))))
		protected.POST("/comments/:id/vote", gin.WrapH(proxy.NewSingle(resolve("video"))))

		protected.POST("/communities", gin.WrapH(proxy.NewSingle(resolve("user"))))
		protected.GET("/communities/me", gin.WrapH(proxy.NewSingle(resolve("user"))))
		protected.POST("/communities/:name/join", gin.WrapH(proxy.NewSingle(resolve("user"))))
		protected.POST("/communities/:name/leave", gin.WrapH(proxy.NewSingle(resolve("user"))))
		protected.GET("/communities/:name/membership", gin.WrapH(proxy.NewSingle(resolve("user"))))
		protected.Any("/search/*path", gin.WrapH(proxy.NewSingle(resolve("search"))))
		protected.Any("/video/*path", gin.WrapH(proxy.NewSingle(resolve("video"))))
		protected.Any("/notifications/*path", gin.WrapH(proxy.NewSingle(resolve("notification"))))
		protected.Any("/chat/*path", gin.WrapH(proxy.NewSingle(resolve("chat"))))
	}

	return &http.Server{Addr: ":" + cfg.Port, Handler: r}
}

func buildGRPCServer(cfg *config.Config) *http.Server {
	mux := http.NewServeMux()

	grpcRoute := func(prefix, target string) {
		mux.Handle(prefix, middleware.GRPCAuth(proxy.NewGRPC("http://"+target)))
	}

	grpcRoute("/user.", cfg.UserGRPCAddr)
	grpcRoute("/feed.", cfg.FeedGRPCAddr)
	grpcRoute("/search.", cfg.SearchGRPCAddr)
	grpcRoute("/video.", cfg.VideoGRPCAddr)
	grpcRoute("/notification.", cfg.NotificationGRPCAddr)

	return &http.Server{
		Addr:    ":" + cfg.GRPCPort,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
}

func staticResolver(cfg *config.Config) func(string) string {
	m := map[string]string{
		"user":         cfg.UserServiceURL,
		"feed":         cfg.FeedServiceURL,
		"search":       cfg.SearchServiceURL,
		"video":        cfg.VideoServiceURL,
		"notification": cfg.NotificationServiceURL,
		"chat":         cfg.ChatServiceURL,
	}
	return func(name string) string { return m[name] }
}

func consulResolver(cfg *config.Config, r *consulpkg.Resolver) func(string) string {
	static := staticResolver(cfg)
	names := []string{"user", "feed", "search", "video", "notification", "chat"}
	m := make(map[string]string, len(names))
	for _, name := range names {
		url, err := r.Resolve(name)
		if err != nil {
			fallback := static(name)
			logger.Warnf("consul: could not resolve %q (%v), falling back to %s\n", name, err, fallback)
			m[name] = fallback
			continue
		}
		m[name] = url
		logger.Infof("consul: %s → %s\n", name, url)
	}
	return func(name string) string { return m[name] }
}
