package handler

import (
	"context"
	"search-service/internal/model"
	"search-service/internal/service"
	pb "search-service/pkg/proto/search"
)

type GrpcSearchHandler struct {
	pb.UnimplementedSearchServiceServer
	svc service.SearchService
}

func NewGrpcSearchHandler(svc service.SearchService) *GrpcSearchHandler {
	return &GrpcSearchHandler{
		svc: svc,
	}
}

func (h *GrpcSearchHandler) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	filters := make(map[string]interface{})
	for k, v := range req.Filters {
		filters[k] = v
	}

	var results []*pb.SearchResult
	var total int64
	var err error

	switch req.Type {
	case "posts":
		posts, t, e := h.svc.SearchPosts(ctx, req.Query, filters, int(req.Limit), int(req.Page))
		total = t
		err = e
		for _, p := range posts {
			results = append(results, &pb.SearchResult{
				Id:        p.ID,
				Type:      "post",
				Title:     p.Title,
				Content:   p.Content,
				Author:    p.AuthorName,
				Community: p.CommunityName,
			})
		}
	case "communities":
		comms, t, e := h.svc.SearchCommunities(ctx, req.Query, int(req.Limit), int(req.Page))
		total = t
		err = e
		for _, c := range comms {
			results = append(results, &pb.SearchResult{
				Id:      c.ID,
				Type:    "community",
				Title:   c.Name,
				Content: c.Description,
			})
		}
	case "users":
		users, t, e := h.svc.SearchUsers(ctx, req.Query, int(req.Limit), int(req.Page))
		total = t
		err = e
		for _, u := range users {
			results = append(results, &pb.SearchResult{
				Id:    u.ID,
				Type:  "user",
				Title: u.Username,
			})
		}
	}

	if err != nil {
		return nil, err
	}

	return &pb.SearchResponse{
		Results: results,
		Total:   total,
	}, nil
}

func (h *GrpcSearchHandler) IndexPost(ctx context.Context, req *pb.IndexPostRequest) (*pb.IndexPostResponse, error) {
	post := model.Post{
		ID:          req.Id,
		Title:       req.Title,
		Content:     req.Content,
		AuthorID:    req.AuthorId,
		CommunityID: req.CommunityId,
	}

	err := h.svc.IndexPost(ctx, post)
	if err != nil {
		return &pb.IndexPostResponse{Success: false}, err
	}
	return &pb.IndexPostResponse{Success: true}, nil
}

func (h *GrpcSearchHandler) IndexCommunity(ctx context.Context, req *pb.IndexCommunityRequest) (*pb.IndexCommunityResponse, error) {
	comm := model.Community{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
	}

	err := h.svc.IndexCommunity(ctx, comm)
	if err != nil {
		return &pb.IndexCommunityResponse{Success: false}, err
	}
	return &pb.IndexCommunityResponse{Success: true}, nil
}

func (h *GrpcSearchHandler) IndexUser(ctx context.Context, req *pb.IndexUserRequest) (*pb.IndexUserResponse, error) {
	user := model.User{
		ID:       req.Id,
		Username: req.Username,
	}

	err := h.svc.IndexUser(ctx, user)
	if err != nil {
		return &pb.IndexUserResponse{Success: false}, err
	}
	return &pb.IndexUserResponse{Success: true}, nil
}

func (h *GrpcSearchHandler) IndexComment(ctx context.Context, req *pb.IndexCommentRequest) (*pb.IndexCommentResponse, error) {
	comment := model.Comment{
		ID:      req.Id,
		Content: req.Content,
		PostID:  req.PostId,
	}

	err := h.svc.IndexComment(ctx, comment)
	if err != nil {
		return &pb.IndexCommentResponse{Success: false}, err
	}
	return &pb.IndexCommentResponse{Success: true}, nil
}
