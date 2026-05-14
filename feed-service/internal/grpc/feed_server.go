package grpc

import (
	"context"
	"feed-service/internal/cache"
	feedpb "feed-service/pkg/proto/feed"
	"time"
)

type FeedServer struct {
	feedpb.UnimplementedFeedServiceServer
	pc *cache.PostCache
}

func NewFeedServer(pc *cache.PostCache) *FeedServer {
	return &FeedServer{pc: pc}
}

func (s *FeedServer) GetPost(ctx context.Context, req *feedpb.GetPostRequest) (*feedpb.GetPostResponse, error) {
	// Search all community keys for the post by stringId.
	// In production this would hit a DB; for now we scan the cache.
	post, err := s.pc.GetByID(ctx, req.PostId)
	if err != nil {
		return nil, err
	}
	return &feedpb.GetPostResponse{
		PostId:    post.StringID,
		AuthorId:  post.Author,
		Title:     post.Title,
		Body:      post.Body,
		CreatedAt: time.Now().Unix(),
	}, nil
}

func (s *FeedServer) StreamFeed(req *feedpb.StreamFeedRequest, stream feedpb.FeedService_StreamFeedServer) error {
	ctx := stream.Context()
	community := req.Subreddit

	posts, err := s.pc.GetByCommunity(ctx, community, 25)
	if err != nil {
		return err
	}

	for i := range posts {
		event := &feedpb.FeedEvent{
			EventType: "new_post",
			Post: &feedpb.GetPostResponse{
				PostId:   posts[i].StringID,
				AuthorId: posts[i].Author,
				Title:    posts[i].Title,
				Body:     posts[i].Body,
			},
		}
		if err := stream.Send(event); err != nil {
			return err
		}
	}
	return nil
}
