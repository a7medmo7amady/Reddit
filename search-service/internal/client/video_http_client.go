package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"search-service/internal/model"
	"strconv"
	"time"
)

type VideoHTTPClient interface {
	Search(ctx context.Context, query string, limit, page int) (*VideoSearchResult, error)
}

type VideoSearchResult struct {
	Posts struct {
		Items []model.Post `json:"items"`
		Total int64        `json:"total"`
		Page  int          `json:"page"`
		Limit int          `json:"limit"`
		Pages int          `json:"pages"`
	} `json:"posts"`
	Comments struct {
		Items []model.Comment `json:"items"`
		Total int64           `json:"total"`
		Page  int             `json:"page"`
		Limit int             `json:"limit"`
		Pages int             `json:"pages"`
	} `json:"comments"`
	Query string `json:"query"`
}

type videoHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewVideoHTTPClient() VideoHTTPClient {
	baseURL := os.Getenv("VIDEO_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://video-service:8083"
	}
	return &videoHTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *videoHTTPClient) Search(ctx context.Context, query string, limit, page int) (*VideoSearchResult, error) {
	url := fmt.Sprintf("%s/search?q=%s&limit=%s&page=%s", c.baseURL, query, strconv.Itoa(limit), strconv.Itoa(page))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("video-service returned %d: %s", resp.StatusCode, string(body))
	}

	var result VideoSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
