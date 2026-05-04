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

	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	cfg := config.Load()
	middleware.SetJWTSecret(cfg.JWTSecret)

	resolve := staticResolver(cfg)
	if cfg.ConsulAddr != "" {
		r, err := consulpkg.New(cfg.ConsulAddr)
		if err != nil {
			logger.Fatalf("consul init: %v\n", err)
		}
		resolve = consulResolver(r)
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

	// Health check — hits each downstream service and reports aggregate status.
	r.GET("/health", handler.Health(map[string]string{
		"user":         resolve("user"),
		"feed":         resolve("feed"),
		"search":       resolve("search"),
		"video":        resolve("video"),
		"notification": resolve("notification"),
	}))

	// Public
	r.Any("/auth/*path", gin.WrapH(proxy.NewSingle(resolve("user"))))

	// Protected
	protected := r.Group("/")
	protected.Use(middleware.Auth())
	{
		protected.Any("/users/*path", gin.WrapH(proxy.NewSingle(resolve("user"))))
		protected.Any("/posts/*path", gin.WrapH(proxy.NewSingle(resolve("feed"))))
		protected.Any("/search/*path", gin.WrapH(proxy.NewSingle(resolve("search"))))
		protected.Any("/video/*path", gin.WrapH(proxy.NewSingle(resolve("video"))))
		protected.Any("/notifications/*path", gin.WrapH(proxy.NewSingle(resolve("notification"))))
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
	}
	return func(name string) string { return m[name] }
}

func consulResolver(r *consulpkg.Resolver) func(string) string {
	names := []string{"user", "feed", "search", "video", "notification"}
	m := make(map[string]string, len(names))
	for _, name := range names {
		url, err := r.Resolve(name)
		if err != nil {
			logger.Warnf("consul: could not resolve %q (%v), will skip\n", name, err)
			continue
		}
		m[name] = url
		logger.Infof("consul: %s → %s\n", name, url)
	}
	return func(name string) string { return m[name] }
}
