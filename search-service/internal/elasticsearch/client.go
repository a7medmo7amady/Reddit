package elasticsearch

import (
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

func NewClient() (*elasticsearch.Client, error) {
	esURL := os.Getenv("ELASTICSEARCH_URL")
	if esURL == "" {
		esURL = "http://localhost:9200"
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			esURL,
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Verify connection
	res, err := client.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}

	log.Println("Successfully connected to Elasticsearch")
	return client, nil
}
