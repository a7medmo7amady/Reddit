package grpc

import (
	"context"
	userpb "feed-service/pkg/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	client userpb.UserServiceClient
}

func NewUserClient(addr string) (*UserClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &UserClient{client: userpb.NewUserServiceClient(conn)}, nil
}

func (u *UserClient) GetUser(ctx context.Context, userID string) (*userpb.GetUserResponse, error) {
	return u.client.GetUser(ctx, &userpb.GetUserRequest{UserId: userID})
}
func (u *UserClient) GetUserBatch(ctx context.Context, userIDs []string) ([]*userpb.GetUserResponse, error) {
	resp, err := u.client.GetUserBatch(ctx, &userpb.GetUserBatchRequest{UserIds: userIDs})
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}
