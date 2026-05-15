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

type userServiceUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Karma     int    `json:"karma"`
	CreatedAt string `json:"createdAt"`
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

	var serviceUsers []userServiceUser
	if err := json.NewDecoder(resp.Body).Decode(&serviceUsers); err != nil {
		return nil, 0, err
	}

	users := make([]model.User, 0, len(serviceUsers))
	for _, user := range serviceUsers {
		users = append(users, model.User{
			ID:        strconv.FormatInt(user.ID, 10),
			Username:  user.Username,
			AvatarURL: user.Avatar,
			Karma:     user.Karma,
			CreatedAt: parseUserCreatedAt(user.CreatedAt),
		})
	}

	return users, int64(len(users)), nil
}

func parseUserCreatedAt(value string) time.Time {
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
	} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
