package main

import (
	"context"
	"log"
	"os"
	"search-service/internal/client"
	"search-service/internal/elasticsearch"
	"search-service/internal/handler"
	"search-service/internal/service"

	"github.com/gin-gonic/gin"
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

	log.Printf("Search service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
