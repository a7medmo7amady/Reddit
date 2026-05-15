package notification

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type SendNotificationRequest struct {
	state         protoimpl.MessageState
	UserId        string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Title         string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Message       string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	Link          string `protobuf:"bytes,4,opt,name=link,proto3" json:"link,omitempty"`
	Type          string `protobuf:"bytes,5,opt,name=type,proto3" json:"type,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SendNotificationRequest) Reset() {
	*x = SendNotificationRequest{}
}
func (x *SendNotificationRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*SendNotificationRequest) ProtoMessage()    {}
func (x *SendNotificationRequest) ProtoReflect() protoreflect.Message {
	return nil
}
func (*SendNotificationRequest) Descriptor() ([]byte, []int) { return nil, nil }

func (x *SendNotificationRequest) GetUserId() string {
	if x != nil { return x.UserId }
	return ""
}
func (x *SendNotificationRequest) GetTitle() string {
	if x != nil { return x.Title }
	return ""
}
func (x *SendNotificationRequest) GetMessage() string {
	if x != nil { return x.Message }
	return ""
}
func (x *SendNotificationRequest) GetLink() string {
	if x != nil { return x.Link }
	return ""
}
func (x *SendNotificationRequest) GetType() string {
	if x != nil { return x.Type }
	return ""
}

type SendNotificationResponse struct {
	state         protoimpl.MessageState
	Success       bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Error         string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SendNotificationResponse) Reset() {
	*x = SendNotificationResponse{}
}
func (x *SendNotificationResponse) String() string { return protoimpl.X.MessageStringOf(x) }
func (*SendNotificationResponse) ProtoMessage()    {}
func (x *SendNotificationResponse) ProtoReflect() protoreflect.Message {
	return nil
}
func (*SendNotificationResponse) Descriptor() ([]byte, []int) { return nil, nil }

func (x *SendNotificationResponse) GetSuccess() bool {
	if x != nil { return x.Success }
	return false
}
func (x *SendNotificationResponse) GetError() string {
	if x != nil { return x.Error }
	return ""
}

type GetUserPreferencesRequest struct {
	state         protoimpl.MessageState
	UserId        string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetUserPreferencesRequest) Reset() {
	*x = GetUserPreferencesRequest{}
}
func (x *GetUserPreferencesRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*GetUserPreferencesRequest) ProtoMessage()    {}
func (x *GetUserPreferencesRequest) ProtoReflect() protoreflect.Message {
	return nil
}
func (*GetUserPreferencesRequest) Descriptor() ([]byte, []int) { return nil, nil }

func (x *GetUserPreferencesRequest) GetUserId() string {
	if x != nil { return x.UserId }
	return ""
}

type GetUserPreferencesResponse struct {
	state         protoimpl.MessageState
	EmailEnabled  bool     `protobuf:"varint,1,opt,name=email_enabled,json=emailEnabled,proto3" json:"email_enabled,omitempty"`
	PushEnabled   bool     `protobuf:"varint,2,opt,name=push_enabled,json=pushEnabled,proto3" json:"push_enabled,omitempty"`
	MutedTypes    []string `protobuf:"bytes,3,rep,name=muted_types,json=mutedTypes,proto3" json:"muted_types,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetUserPreferencesResponse) Reset() {
	*x = GetUserPreferencesResponse{}
}
func (x *GetUserPreferencesResponse) String() string { return protoimpl.X.MessageStringOf(x) }
func (*GetUserPreferencesResponse) ProtoMessage()    {}
func (x *GetUserPreferencesResponse) ProtoReflect() protoreflect.Message {
	return nil
}
func (*GetUserPreferencesResponse) Descriptor() ([]byte, []int) { return nil, nil }

func (x *GetUserPreferencesResponse) GetEmailEnabled() bool {
	if x != nil { return x.EmailEnabled }
	return false
}
func (x *GetUserPreferencesResponse) GetPushEnabled() bool {
	if x != nil { return x.PushEnabled }
	return false
}
func (x *GetUserPreferencesResponse) GetMutedTypes() []string {
	if x != nil { return x.MutedTypes }
	return nil
}

type UpdateUserPreferencesRequest struct {
	state         protoimpl.MessageState
	UserId        string   `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	EmailEnabled  bool     `protobuf:"varint,2,opt,name=email_enabled,json=emailEnabled,proto3" json:"email_enabled,omitempty"`
	PushEnabled   bool     `protobuf:"varint,3,opt,name=push_enabled,json=pushEnabled,proto3" json:"push_enabled,omitempty"`
	MutedTypes    []string `protobuf:"bytes,4,rep,name=muted_types,json=mutedTypes,proto3" json:"muted_types,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateUserPreferencesRequest) Reset() {
	*x = UpdateUserPreferencesRequest{}
}
func (x *UpdateUserPreferencesRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*UpdateUserPreferencesRequest) ProtoMessage()    {}
func (x *UpdateUserPreferencesRequest) ProtoReflect() protoreflect.Message {
	return nil
}
func (*UpdateUserPreferencesRequest) Descriptor() ([]byte, []int) { return nil, nil }

func (x *UpdateUserPreferencesRequest) GetUserId() string {
	if x != nil { return x.UserId }
	return ""
}
func (x *UpdateUserPreferencesRequest) GetEmailEnabled() bool {
	if x != nil { return x.EmailEnabled }
	return false
}
func (x *UpdateUserPreferencesRequest) GetPushEnabled() bool {
	if x != nil { return x.PushEnabled }
	return false
}
func (x *UpdateUserPreferencesRequest) GetMutedTypes() []string {
	if x != nil { return x.MutedTypes }
	return nil
}

type UpdateUserPreferencesResponse struct {
	state         protoimpl.MessageState
	Success       bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateUserPreferencesResponse) Reset() {
	*x = UpdateUserPreferencesResponse{}
}
func (x *UpdateUserPreferencesResponse) String() string { return protoimpl.X.MessageStringOf(x) }
func (*UpdateUserPreferencesResponse) ProtoMessage()    {}
func (x *UpdateUserPreferencesResponse) ProtoReflect() protoreflect.Message {
	return nil
}
func (*UpdateUserPreferencesResponse) Descriptor() ([]byte, []int) { return nil, nil }

func (x *UpdateUserPreferencesResponse) GetSuccess() bool {
	if x != nil { return x.Success }
	return false
}
