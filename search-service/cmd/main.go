package main

import (
	"context"
	"log"
	"os"
	"search-service/internal/client"
	"search-service/internal/consumer"
	"search-service/internal/elasticsearch"
	"search-service/internal/handler"
	"search-service/internal/service"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"net"
	pb "search-service/pkg/proto/search"
)

func main() {
	// Initialize Elasticsearch client
	esClient, err := elasticsearch.NewClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// Initialize gRPC clients
	videoClient, err := client.NewVideoClient()
	if err != nil {
		log.Printf("Warning: Failed to connect to Video Service: %v", err)
	} else {
		defer videoClient.Close()
	}

	// Initialize services and handlers
	searchService := service.NewSearchService(esClient)
	searchHandler := handler.NewSearchHandler(searchService, videoClient)

	// Initialize and start Kafka Consumer for indexing
	searchConsumer, err := consumer.NewSearchConsumer(searchService)
	if err != nil {
		log.Printf("Warning: Failed to start Search Kafka Consumer: %v", err)
	} else {
		searchConsumer.Start(context.Background())
		defer searchConsumer.Close()
	}

	// Create index if it doesn't exist
	if err := searchService.CreateIndex(context.Background()); err != nil {
		log.Printf("Warning: Failed to create index: %v", err)
	}

	// Set up Gin router
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Search endpoints
	searchGroup := r.Group("/api/v1/search")
	{
		searchGroup.GET("/", searchHandler.Search) // Universal search
		searchGroup.POST("/posts", searchHandler.IndexPost)
		searchGroup.POST("/communities", searchHandler.IndexCommunity)
		searchGroup.POST("/users", searchHandler.IndexUser)
		searchGroup.POST("/comments", searchHandler.IndexComment)
		searchGroup.POST("/sync", searchHandler.Sync)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053"
	}

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen for gRPC: %v", err)
		}

		s := grpc.NewServer()
		grpcHandler := handler.NewGrpcSearchHandler(searchService)
		pb.RegisterSearchServiceServer(s, grpcHandler)

		log.Printf("Search gRPC Service starting on port %s", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	log.Printf("Search HTTP Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
