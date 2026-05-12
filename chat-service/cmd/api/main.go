package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat-service/internal/config"
	"chat-service/internal/db"
	"chat-service/internal/events"
	chathttp "chat-service/internal/http"
	"chat-service/internal/http/handlers"
	"chat-service/internal/realtime"
	"chat-service/internal/service"
	"chat-service/internal/userclient"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mongoClient, err := db.ConnectMongo(cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}

	database := mongoClient.Database(cfg.MongoDatabase)

	userClient := userclient.New(cfg.UserServiceURL)
	producer := events.NewLogProducer()

	var fanout realtime.Fanout
	var offline realtime.OfflineQueue
	if cfg.RedisAddr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
		broker := realtime.NewRedisBroker(rdb)
		fanout = broker
		offline = realtime.NewRedisOfflineQueue(rdb, 200)
	}

	hub := realtime.NewHub()
	hub.ClearExpiredTyping()

	dispatcher := realtime.NewDispatcher(hub, cfg.InstanceID, fanout, offline)
	dispatcher.Start(ctx)

	chatService := service.NewChatService(database, userClient, producer, dispatcher)
	chatService.Start(ctx)
	chatHandler := handlers.NewChatHandler(chatService)

	wsHandler := realtime.NewHandler(hub, dispatcher, cfg.MaxWSConns)

	r := gin.Default()
	chathttp.RegisterRoutes(r, chatHandler, wsHandler)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("chat service running on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
