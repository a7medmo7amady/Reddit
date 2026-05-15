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

type UserClient interface {
	SearchUsers(ctx context.Context, query string, limit, page int) ([]model.User, int64, error)
}

type userClient struct {
	baseURL string
	client  *http.Client
}

func NewUserClient() UserClient {
	baseURL := os.Getenv("USER_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://user-service:8080"
	}
	return &userClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *userClient) SearchUsers(ctx context.Context, query string, limit, page int) ([]model.User, int64, error) {
	url := fmt.Sprintf("%s/search/users?q=%s&limit=%s&page=%s", c.baseURL, query, strconv.Itoa(limit), strconv.Itoa(page))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("user-service returned %d: %s", resp.StatusCode, string(body))
	}

	var users []model.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, 0, err
	}

	return users, int64(len(users)), nil
}
