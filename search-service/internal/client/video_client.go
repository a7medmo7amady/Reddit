package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"search-service/internal/model"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VideoClient interface {
	FetchPosts(ctx context.Context, limit, page int) ([]model.Post, int64, error)
	Close() error
}

type videoClient struct {
	conn *grpc.ClientConn
	// client videov1.VideoServiceClient
}

func NewVideoClient() (VideoClient, error) {
	addr := os.Getenv("VIDEO_GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50054"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	return &videoClient{
		conn: conn,
		// client: videov1.NewVideoServiceClient(conn),
	}, nil
}

func (c *videoClient) FetchPosts(ctx context.Context, limit, page int) ([]model.Post, int64, error) {
	log.Printf("Fetching posts from Video Service (gRPC) - limit: %d, page: %d", limit, page)
	
	// Since we don't have the generated code, we mock the response for now
	// but the structure is ready for the real gRPC call.
	
	/*
	resp, err := c.client.ListPosts(ctx, &videov1.ListPostsRequest{
		Limit: int32(limit),
		Page:  int32(page),
	})
	if err != nil {
		return nil, 0, err
	}
	
	posts := make([]model.Post, len(resp.Posts))
	for i, p := range resp.Posts {
		posts[i] = model.Post{
			ID:          p.Id,
			Title:       p.Title,
			Content:     p.Body,
			AuthorID:    p.AuthorId,
			CommunityID: p.CommunityId,
			URL:         p.Url,
			Flair:       p.Flair,
			NSFW:        p.Nsfw,
			Spoiler:     p.Spoiler,
			OC:          p.Oc,
			CreatedAt:   time.Unix(p.CreatedAt, 0),
		}
	}
	return posts, resp.Total, nil
	*/

	// Placeholder return
	return nil, 0, fmt.Errorf("gRPC generated code missing - please run protoc")
}

func (c *videoClient) Close() error {
	return c.conn.Close()
}
