// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.27.2
// source: notification-service.proto

package chorus

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CountUnreadNotificationsReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result uint32 `protobuf:"varint,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *CountUnreadNotificationsReply) Reset() {
	*x = CountUnreadNotificationsReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CountUnreadNotificationsReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CountUnreadNotificationsReply) ProtoMessage() {}

func (x *CountUnreadNotificationsReply) ProtoReflect() protoreflect.Message {
	mi := &file_notification_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CountUnreadNotificationsReply.ProtoReflect.Descriptor instead.
func (*CountUnreadNotificationsReply) Descriptor() ([]byte, []int) {
	return file_notification_service_proto_rawDescGZIP(), []int{0}
}

func (x *CountUnreadNotificationsReply) GetResult() uint32 {
	if x != nil {
		return x.Result
	}
	return 0
}

type MarkNotificationsAsReadRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NotificationIds []string `protobuf:"bytes,1,rep,name=notificationIds,proto3" json:"notificationIds,omitempty"`
	MarkAll         bool     `protobuf:"varint,2,opt,name=markAll,proto3" json:"markAll,omitempty"`
}

func (x *MarkNotificationsAsReadRequest) Reset() {
	*x = MarkNotificationsAsReadRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MarkNotificationsAsReadRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MarkNotificationsAsReadRequest) ProtoMessage() {}

func (x *MarkNotificationsAsReadRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notification_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MarkNotificationsAsReadRequest.ProtoReflect.Descriptor instead.
func (*MarkNotificationsAsReadRequest) Descriptor() ([]byte, []int) {
	return file_notification_service_proto_rawDescGZIP(), []int{1}
}

func (x *MarkNotificationsAsReadRequest) GetNotificationIds() []string {
	if x != nil {
		return x.NotificationIds
	}
	return nil
}

func (x *MarkNotificationsAsReadRequest) GetMarkAll() bool {
	if x != nil {
		return x.MarkAll
	}
	return false
}

type GetNotificationsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pagination *PaginationQuery    `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
	IsRead     *wrappers.BoolValue `protobuf:"bytes,2,opt,name=isRead,proto3" json:"isRead,omitempty"`
}

func (x *GetNotificationsRequest) Reset() {
	*x = GetNotificationsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetNotificationsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNotificationsRequest) ProtoMessage() {}

func (x *GetNotificationsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notification_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNotificationsRequest.ProtoReflect.Descriptor instead.
func (*GetNotificationsRequest) Descriptor() ([]byte, []int) {
	return file_notification_service_proto_rawDescGZIP(), []int{2}
}

func (x *GetNotificationsRequest) GetPagination() *PaginationQuery {
	if x != nil {
		return x.Pagination
	}
	return nil
}

func (x *GetNotificationsRequest) GetIsRead() *wrappers.BoolValue {
	if x != nil {
		return x.IsRead
	}
	return nil
}

type GetNotificationsReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result     []*Notification `protobuf:"bytes,1,rep,name=result,proto3" json:"result,omitempty"`
	TotalItems uint32          `protobuf:"varint,2,opt,name=totalItems,proto3" json:"totalItems,omitempty"`
}

func (x *GetNotificationsReply) Reset() {
	*x = GetNotificationsReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetNotificationsReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNotificationsReply) ProtoMessage() {}

func (x *GetNotificationsReply) ProtoReflect() protoreflect.Message {
	mi := &file_notification_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNotificationsReply.ProtoReflect.Descriptor instead.
func (*GetNotificationsReply) Descriptor() ([]byte, []int) {
	return file_notification_service_proto_rawDescGZIP(), []int{3}
}

func (x *GetNotificationsReply) GetResult() []*Notification {
	if x != nil {
		return x.Result
	}
	return nil
}

func (x *GetNotificationsReply) GetTotalItems() uint32 {
	if x != nil {
		return x.TotalItems
	}
	return 0
}

var File_notification_service_proto protoreflect.FileDescriptor

var file_notification_service_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2d, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x68,
	0x6f, 0x72, 0x75, 0x73, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e,
	0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x12, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x37, 0x0a, 0x1d, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x55, 0x6e, 0x72, 0x65, 0x61, 0x64,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x64, 0x0a, 0x1e, 0x4d, 0x61,
	0x72, 0x6b, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x41,
	0x73, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x28, 0x0a, 0x0f,
	0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x49, 0x64, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x61, 0x72, 0x6b, 0x41, 0x6c,
	0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x6d, 0x61, 0x72, 0x6b, 0x41, 0x6c, 0x6c,
	0x22, 0x86, 0x01, 0x0a, 0x17, 0x47, 0x65, 0x74, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x37, 0x0a, 0x0a,
	0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x32, 0x0a, 0x06, 0x69, 0x73, 0x52, 0x65, 0x61, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x52, 0x06, 0x69, 0x73, 0x52, 0x65, 0x61, 0x64, 0x22, 0x65, 0x0a, 0x15, 0x47, 0x65, 0x74,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x12, 0x2c, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x4e, 0x6f, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x12, 0x1e, 0x0a, 0x0a, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x49, 0x74, 0x65, 0x6d, 0x73,
	0x32, 0xc8, 0x05, 0x0a, 0x13, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0xf2, 0x01, 0x0a, 0x18, 0x43, 0x6f, 0x75,
	0x6e, 0x74, 0x55, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x25, 0x2e,
	0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x55, 0x6e, 0x72, 0x65,
	0x61, 0x64, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x96, 0x01, 0x92, 0x41, 0x6b, 0x0a, 0x13, 0x4e, 0x6f, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x1a, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x20, 0x75, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x20, 0x6e, 0x6f,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x38, 0x54, 0x68, 0x69,
	0x73, 0x20, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x20, 0x72, 0x65, 0x74, 0x75, 0x72,
	0x6e, 0x73, 0x20, 0x74, 0x68, 0x65, 0x20, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x20, 0x6f, 0x66,
	0x20, 0x75, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x20, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x22, 0x12, 0x20, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x72, 0x65, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0xe7, 0x01,
	0x0a, 0x17, 0x4d, 0x61, 0x72, 0x6b, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x41, 0x73, 0x52, 0x65, 0x61, 0x64, 0x12, 0x26, 0x2e, 0x63, 0x68, 0x6f, 0x72,
	0x75, 0x73, 0x2e, 0x4d, 0x61, 0x72, 0x6b, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x41, 0x73, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x8b, 0x01, 0x92, 0x41, 0x5e, 0x0a,
	0x13, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x1b, 0x4d, 0x61, 0x72, 0x6b, 0x20, 0x61, 0x20, 0x6e, 0x6f, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x61, 0x73, 0x20, 0x72, 0x65, 0x61,
	0x64, 0x1a, 0x2a, 0x54, 0x68, 0x69, 0x73, 0x20, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x20, 0x6d, 0x61, 0x72, 0x6b, 0x73, 0x20, 0x61, 0x20, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x61, 0x73, 0x20, 0x72, 0x65, 0x61, 0x64, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x24, 0x3a, 0x01, 0x2a, 0x22, 0x1f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73,
	0x74, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x72, 0x65, 0x61, 0x64, 0x12, 0xd1, 0x01, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x4e,
	0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1f, 0x2e, 0x63,
	0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e,
	0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x7d, 0x92, 0x41,
	0x58, 0x0a, 0x13, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x12, 0x4c, 0x69, 0x73, 0x74, 0x20, 0x6e, 0x6f, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x2d, 0x54, 0x68, 0x69, 0x73,
	0x20, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x20, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e,
	0x73, 0x20, 0x61, 0x20, 0x6c, 0x69, 0x73, 0x74, 0x20, 0x6f, 0x66, 0x20, 0x6e, 0x6f, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1c, 0x12,
	0x1a, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x42, 0xba, 0x01, 0x92, 0x41,
	0xac, 0x01, 0x12, 0x82, 0x01, 0x0a, 0x1b, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x20, 0x6e, 0x6f,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x22, 0x5e, 0x0a, 0x1b, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x20, 0x6e, 0x6f, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x2c, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x43, 0x48, 0x4f, 0x52, 0x55, 0x53, 0x2d, 0x54, 0x52, 0x45,
	0x2f, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2d, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x1a,
	0x11, 0x64, 0x65, 0x76, 0x40, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2d, 0x74, 0x72, 0x65, 0x2e,
	0x63, 0x68, 0x32, 0x03, 0x31, 0x2e, 0x30, 0x2a, 0x01, 0x01, 0x32, 0x10, 0x61, 0x70, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x3a, 0x10, 0x61, 0x70,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x5a, 0x08,
	0x2e, 0x3b, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_notification_service_proto_rawDescOnce sync.Once
	file_notification_service_proto_rawDescData = file_notification_service_proto_rawDesc
)

func file_notification_service_proto_rawDescGZIP() []byte {
	file_notification_service_proto_rawDescOnce.Do(func() {
		file_notification_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_notification_service_proto_rawDescData)
	})
	return file_notification_service_proto_rawDescData
}

var file_notification_service_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_notification_service_proto_goTypes = []interface{}{
	(*CountUnreadNotificationsReply)(nil),  // 0: chorus.CountUnreadNotificationsReply
	(*MarkNotificationsAsReadRequest)(nil), // 1: chorus.MarkNotificationsAsReadRequest
	(*GetNotificationsRequest)(nil),        // 2: chorus.GetNotificationsRequest
	(*GetNotificationsReply)(nil),          // 3: chorus.GetNotificationsReply
	(*PaginationQuery)(nil),                // 4: chorus.PaginationQuery
	(*wrappers.BoolValue)(nil),             // 5: google.protobuf.BoolValue
	(*Notification)(nil),                   // 6: chorus.Notification
	(*empty.Empty)(nil),                    // 7: google.protobuf.Empty
}
var file_notification_service_proto_depIdxs = []int32{
	4, // 0: chorus.GetNotificationsRequest.pagination:type_name -> chorus.PaginationQuery
	5, // 1: chorus.GetNotificationsRequest.isRead:type_name -> google.protobuf.BoolValue
	6, // 2: chorus.GetNotificationsReply.result:type_name -> chorus.Notification
	7, // 3: chorus.NotificationService.CountUnreadNotifications:input_type -> google.protobuf.Empty
	1, // 4: chorus.NotificationService.MarkNotificationsAsRead:input_type -> chorus.MarkNotificationsAsReadRequest
	2, // 5: chorus.NotificationService.GetNotifications:input_type -> chorus.GetNotificationsRequest
	0, // 6: chorus.NotificationService.CountUnreadNotifications:output_type -> chorus.CountUnreadNotificationsReply
	7, // 7: chorus.NotificationService.MarkNotificationsAsRead:output_type -> google.protobuf.Empty
	3, // 8: chorus.NotificationService.GetNotifications:output_type -> chorus.GetNotificationsReply
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_notification_service_proto_init() }
func file_notification_service_proto_init() {
	if File_notification_service_proto != nil {
		return
	}
	file_notification_proto_init()
	file_common_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_notification_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CountUnreadNotificationsReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MarkNotificationsAsReadRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetNotificationsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetNotificationsReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_notification_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_notification_service_proto_goTypes,
		DependencyIndexes: file_notification_service_proto_depIdxs,
		MessageInfos:      file_notification_service_proto_msgTypes,
	}.Build()
	File_notification_service_proto = out.File
	file_notification_service_proto_rawDesc = nil
	file_notification_service_proto_goTypes = nil
	file_notification_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// NotificationServiceClient is the client API for NotificationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NotificationServiceClient interface {
	CountUnreadNotifications(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CountUnreadNotificationsReply, error)
	MarkNotificationsAsRead(ctx context.Context, in *MarkNotificationsAsReadRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	GetNotifications(ctx context.Context, in *GetNotificationsRequest, opts ...grpc.CallOption) (*GetNotificationsReply, error)
}

type notificationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNotificationServiceClient(cc grpc.ClientConnInterface) NotificationServiceClient {
	return &notificationServiceClient{cc}
}

func (c *notificationServiceClient) CountUnreadNotifications(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CountUnreadNotificationsReply, error) {
	out := new(CountUnreadNotificationsReply)
	err := c.cc.Invoke(ctx, "/chorus.NotificationService/CountUnreadNotifications", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) MarkNotificationsAsRead(ctx context.Context, in *MarkNotificationsAsReadRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/chorus.NotificationService/MarkNotificationsAsRead", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) GetNotifications(ctx context.Context, in *GetNotificationsRequest, opts ...grpc.CallOption) (*GetNotificationsReply, error) {
	out := new(GetNotificationsReply)
	err := c.cc.Invoke(ctx, "/chorus.NotificationService/GetNotifications", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NotificationServiceServer is the server API for NotificationService service.
type NotificationServiceServer interface {
	CountUnreadNotifications(context.Context, *empty.Empty) (*CountUnreadNotificationsReply, error)
	MarkNotificationsAsRead(context.Context, *MarkNotificationsAsReadRequest) (*empty.Empty, error)
	GetNotifications(context.Context, *GetNotificationsRequest) (*GetNotificationsReply, error)
}

// UnimplementedNotificationServiceServer can be embedded to have forward compatible implementations.
type UnimplementedNotificationServiceServer struct {
}

func (*UnimplementedNotificationServiceServer) CountUnreadNotifications(context.Context, *empty.Empty) (*CountUnreadNotificationsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CountUnreadNotifications not implemented")
}
func (*UnimplementedNotificationServiceServer) MarkNotificationsAsRead(context.Context, *MarkNotificationsAsReadRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarkNotificationsAsRead not implemented")
}
func (*UnimplementedNotificationServiceServer) GetNotifications(context.Context, *GetNotificationsRequest) (*GetNotificationsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNotifications not implemented")
}

func RegisterNotificationServiceServer(s *grpc.Server, srv NotificationServiceServer) {
	s.RegisterService(&_NotificationService_serviceDesc, srv)
}

func _NotificationService_CountUnreadNotifications_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).CountUnreadNotifications(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chorus.NotificationService/CountUnreadNotifications",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).CountUnreadNotifications(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_MarkNotificationsAsRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MarkNotificationsAsReadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).MarkNotificationsAsRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chorus.NotificationService/MarkNotificationsAsRead",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).MarkNotificationsAsRead(ctx, req.(*MarkNotificationsAsReadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_GetNotifications_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNotificationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).GetNotifications(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chorus.NotificationService/GetNotifications",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).GetNotifications(ctx, req.(*GetNotificationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _NotificationService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "chorus.NotificationService",
	HandlerType: (*NotificationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CountUnreadNotifications",
			Handler:    _NotificationService_CountUnreadNotifications_Handler,
		},
		{
			MethodName: "MarkNotificationsAsRead",
			Handler:    _NotificationService_MarkNotificationsAsRead_Handler,
		},
		{
			MethodName: "GetNotifications",
			Handler:    _NotificationService_GetNotifications_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "notification-service.proto",
}