package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"feed-service/internal/cache"
	"feed-service/internal/model"
)

type VideoClient struct {
	addr string
}

func NewVideoClient(addr string) (*VideoClient, error) {
	return &VideoClient{addr: addr}, nil
}

// SyncCommunityPosts fetches posts from VideoService via HTTP and writes them into PostCache and TrendingCache.
func (v *VideoClient) SyncCommunityPosts(ctx context.Context, community string, pc *cache.PostCache, tc *cache.TrendingCache) error {
	u := fmt.Sprintf("http://%s/posts?community=%s&limit=50&page=1", v.addr, url.QueryEscape(community))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("VideoService returned status %d", resp.StatusCode)
	}

	var data struct {
		Posts []struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Body      string `json:"body"`
			Community string `json:"community"`
			AuthorId  string `json:"authorId"`
		} `json:"posts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	for _, p := range data.Posts {
		post := model.Post{
			StringID:  p.ID,
			Title:     p.Title,
			Body:      p.Body,
			Community: p.Community,
			Author:    p.AuthorId,
		}
		if err := pc.Add(ctx, post); err != nil {
			log.Printf("[VideoClient] PostCache write error for post %s: %v", p.ID, err)
		}
		if err := tc.AddIfNotExists(ctx, post); err != nil {
			log.Printf("[VideoClient] TrendingCache write error for post %s: %v", p.ID, err)
		}
	}
	log.Printf("[VideoClient] HTTP synced %d posts for r/%s", len(data.Posts), community)
	return nil
}

// GetPost fetches a single post by ID from VideoService via HTTP.
func (v *VideoClient) GetPost(ctx context.Context, postID string) (*model.Post, error) {
	u := fmt.Sprintf("http://%s/posts/%s", v.addr, url.QueryEscape(postID))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VideoService returned status %d", resp.StatusCode)
	}

	var p struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		Community string `json:"community"`
		AuthorId  string `json:"authorId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}

	return &model.Post{
		StringID:  p.ID,
		Title:     p.Title,
		Body:      p.Body,
		Community: p.Community,
		Author:    p.AuthorId,
	}, nil
}
