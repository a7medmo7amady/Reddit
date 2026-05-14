package grpc

import (
	"context"
	"feed-service/internal/cache"
	"feed-service/internal/model"
	videopb "feed-service/pkg/proto/video"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VideoClient struct {
	client videopb.VideoServiceClient
}

func NewVideoClient(addr string) (*VideoClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &VideoClient{client: videopb.NewVideoServiceClient(conn)}, nil
}

// SyncCommunityPosts fetches posts from VideoService and writes them into PostCache and TrendingCache.
func (v *VideoClient) SyncCommunityPosts(ctx context.Context, community string, pc *cache.PostCache, tc *cache.TrendingCache) error {
	resp, err := v.client.ListPosts(ctx, &videopb.ListPostsRequest{
		Community: community,
		Limit:     50,
		Page:      1,
	})
	if err != nil {
		return err
	}

	for _, p := range resp.Posts {
		post := model.Post{
			StringID:  p.Id,
			Title:     p.Title,
			Body:      p.Body,
			Community: p.CommunityId,
			Author:    p.AuthorId,
		}
		if err := pc.Add(ctx, post); err != nil {
			log.Printf("[VideoClient] PostCache write error for post %s: %v", p.Id, err)
		}
		if err := tc.AddIfNotExists(ctx, post); err != nil {
			log.Printf("[VideoClient] TrendingCache write error for post %s: %v", p.Id, err)
		}
	}
	log.Printf("[VideoClient] synced %d posts for r/%s", len(resp.Posts), community)
	return nil
}

// GetPost fetches a single post by ID from VideoService.
func (v *VideoClient) GetPost(ctx context.Context, postID string) (*videopb.Post, error) {
	resp, err := v.client.GetPost(ctx, &videopb.GetPostRequest{PostId: postID})
	if err != nil {
		return nil, err
	}
	return resp.Post, nil
}
