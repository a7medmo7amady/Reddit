package notification

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	NotificationService_SendNotification_FullMethodName         = "/notification.NotificationService/SendNotification"
	NotificationService_GetUserPreferences_FullMethodName         = "/notification.NotificationService/GetUserPreferences"
	NotificationService_UpdateUserPreferences_FullMethodName      = "/notification.NotificationService/UpdateUserPreferences"
)

type NotificationServiceClient interface {
	SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*SendNotificationResponse, error)
	GetUserPreferences(ctx context.Context, in *GetUserPreferencesRequest, opts ...grpc.CallOption) (*GetUserPreferencesResponse, error)
	UpdateUserPreferences(ctx context.Context, in *UpdateUserPreferencesRequest, opts ...grpc.CallOption) (*UpdateUserPreferencesResponse, error)
}

type notificationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNotificationServiceClient(cc grpc.ClientConnInterface) NotificationServiceClient {
	return &notificationServiceClient{cc}
}

func (c *notificationServiceClient) SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*SendNotificationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendNotificationResponse)
	err := c.cc.Invoke(ctx, NotificationService_SendNotification_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) GetUserPreferences(ctx context.Context, in *GetUserPreferencesRequest, opts ...grpc.CallOption) (*GetUserPreferencesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetUserPreferencesResponse)
	err := c.cc.Invoke(ctx, NotificationService_GetUserPreferences_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) UpdateUserPreferences(ctx context.Context, in *UpdateUserPreferencesRequest, opts ...grpc.CallOption) (*UpdateUserPreferencesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateUserPreferencesResponse)
	err := c.cc.Invoke(ctx, NotificationService_UpdateUserPreferences_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NotificationServiceServer is the server API for NotificationService service.
type NotificationServiceServer interface {
	SendNotification(context.Context, *SendNotificationRequest) (*SendNotificationResponse, error)
	GetUserPreferences(context.Context, *GetUserPreferencesRequest) (*GetUserPreferencesResponse, error)
	UpdateUserPreferences(context.Context, *UpdateUserPreferencesRequest) (*UpdateUserPreferencesResponse, error)
}

// UnimplementedNotificationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNotificationServiceServer struct{}

func (UnimplementedNotificationServiceServer) SendNotification(context.Context, *SendNotificationRequest) (*SendNotificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendNotification not implemented")
}
func (UnimplementedNotificationServiceServer) GetUserPreferences(context.Context, *GetUserPreferencesRequest) (*GetUserPreferencesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserPreferences not implemented")
}
func (UnimplementedNotificationServiceServer) UpdateUserPreferences(context.Context, *UpdateUserPreferencesRequest) (*UpdateUserPreferencesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUserPreferences not implemented")
}

// UnsafeNotificationServiceServer may be embedded to opt out of forward compatibility for this service.
type UnsafeNotificationServiceServer interface {
	mustEmbedUnimplementedNotificationServiceServer()
}

func RegisterNotificationServiceServer(s grpc.ServiceRegistrar, srv NotificationServiceServer) {
	s.RegisterService(&NotificationService_ServiceDesc, srv)
}

func _NotificationService_SendNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).SendNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NotificationService_SendNotification_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).SendNotification(ctx, req.(*SendNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_GetUserPreferences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserPreferencesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).GetUserPreferences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NotificationService_GetUserPreferences_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).GetUserPreferences(ctx, req.(*GetUserPreferencesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_UpdateUserPreferences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserPreferencesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).UpdateUserPreferences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NotificationService_UpdateUserPreferences_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).UpdateUserPreferences(ctx, req.(*UpdateUserPreferencesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// NotificationService_ServiceDesc is the grpc.ServiceDesc for NotificationService service.
var NotificationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "notification.NotificationService",
	HandlerType: (*NotificationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendNotification",
			Handler:    _NotificationService_SendNotification_Handler,
		},
		{
			MethodName: "GetUserPreferences",
			Handler:    _NotificationService_GetUserPreferences_Handler,
		},
		{
			MethodName: "UpdateUserPreferences",
			Handler:    _NotificationService_UpdateUserPreferences_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "notification.proto",
}
