package userclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) UserExists(ctx context.Context, userID string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/users/"+url.PathEscape(strings.TrimSpace(userID))+"/exists", nil)
	if err != nil {
		return false, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("user service returned %d", res.StatusCode)
	}

	var body struct {
		Exists bool `json:"exists"`
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return false, err
	}

	return body.Exists, nil
}

func (c *Client) IsBlocked(ctx context.Context, senderID, receiverID string) (bool, error) {
	url := fmt.Sprintf(
		"%s/internal/users/%s/blocked/%s",
		c.baseURL,
		url.PathEscape(strings.TrimSpace(senderID)),
		url.PathEscape(strings.TrimSpace(receiverID)),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("user service returned %d", res.StatusCode)
	}

	var body struct {
		Blocked bool `json:"blocked"`
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return false, err
	}

	return body.Blocked, nil
}
