// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.27.2
// source: attachment-service.proto

package chorus

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
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

type GetAttachmentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetAttachmentRequest) Reset() {
	*x = GetAttachmentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_attachment_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetAttachmentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAttachmentRequest) ProtoMessage() {}

func (x *GetAttachmentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_attachment_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAttachmentRequest.ProtoReflect.Descriptor instead.
func (*GetAttachmentRequest) Descriptor() ([]byte, []int) {
	return file_attachment_service_proto_rawDescGZIP(), []int{0}
}

func (x *GetAttachmentRequest) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type GetAttachmentReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result *Attachment `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *GetAttachmentReply) Reset() {
	*x = GetAttachmentReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_attachment_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetAttachmentReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAttachmentReply) ProtoMessage() {}

func (x *GetAttachmentReply) ProtoReflect() protoreflect.Message {
	mi := &file_attachment_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAttachmentReply.ProtoReflect.Descriptor instead.
func (*GetAttachmentReply) Descriptor() ([]byte, []int) {
	return file_attachment_service_proto_rawDescGZIP(), []int{1}
}

func (x *GetAttachmentReply) GetResult() *Attachment {
	if x != nil {
		return x.Result
	}
	return nil
}

type CreateAttachmentsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ResourceId   uint64                     `protobuf:"varint,1,opt,name=resourceId,proto3" json:"resourceId,omitempty"`
	ResourceType string                     `protobuf:"bytes,2,opt,name=resourceType,proto3" json:"resourceType,omitempty"`
	Attachments  []*CreateAttachmentRequest `protobuf:"bytes,3,rep,name=attachments,proto3" json:"attachments,omitempty"`
}

func (x *CreateAttachmentsRequest) Reset() {
	*x = CreateAttachmentsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_attachment_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateAttachmentsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateAttachmentsRequest) ProtoMessage() {}

func (x *CreateAttachmentsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_attachment_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateAttachmentsRequest.ProtoReflect.Descriptor instead.
func (*CreateAttachmentsRequest) Descriptor() ([]byte, []int) {
	return file_attachment_service_proto_rawDescGZIP(), []int{2}
}

func (x *CreateAttachmentsRequest) GetResourceId() uint64 {
	if x != nil {
		return x.ResourceId
	}
	return 0
}

func (x *CreateAttachmentsRequest) GetResourceType() string {
	if x != nil {
		return x.ResourceType
	}
	return ""
}

func (x *CreateAttachmentsRequest) GetAttachments() []*CreateAttachmentRequest {
	if x != nil {
		return x.Attachments
	}
	return nil
}

type CreateAttachmentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key              string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Filename         string `protobuf:"bytes,2,opt,name=filename,proto3" json:"filename,omitempty"`
	Value            string `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	ContentType      string `protobuf:"bytes,4,opt,name=contentType,proto3" json:"contentType,omitempty"`
	Location         string `protobuf:"bytes,5,opt,name=location,proto3" json:"location,omitempty"`
	DocumentCategory string `protobuf:"bytes,6,opt,name=documentCategory,proto3" json:"documentCategory,omitempty"`
}

func (x *CreateAttachmentRequest) Reset() {
	*x = CreateAttachmentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_attachment_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateAttachmentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateAttachmentRequest) ProtoMessage() {}

func (x *CreateAttachmentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_attachment_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateAttachmentRequest.ProtoReflect.Descriptor instead.
func (*CreateAttachmentRequest) Descriptor() ([]byte, []int) {
	return file_attachment_service_proto_rawDescGZIP(), []int{3}
}

func (x *CreateAttachmentRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *CreateAttachmentRequest) GetFilename() string {
	if x != nil {
		return x.Filename
	}
	return ""
}

func (x *CreateAttachmentRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *CreateAttachmentRequest) GetContentType() string {
	if x != nil {
		return x.ContentType
	}
	return ""
}

func (x *CreateAttachmentRequest) GetLocation() string {
	if x != nil {
		return x.Location
	}
	return ""
}

func (x *CreateAttachmentRequest) GetDocumentCategory() string {
	if x != nil {
		return x.DocumentCategory
	}
	return ""
}

type DeleteAttachmentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeleteAttachmentRequest) Reset() {
	*x = DeleteAttachmentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_attachment_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteAttachmentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteAttachmentRequest) ProtoMessage() {}

func (x *DeleteAttachmentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_attachment_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteAttachmentRequest.ProtoReflect.Descriptor instead.
func (*DeleteAttachmentRequest) Descriptor() ([]byte, []int) {
	return file_attachment_service_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteAttachmentRequest) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type DeleteAttachmentReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result *DeleteAttachmentResult `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *DeleteAttachmentReply) Reset() {
	*x = DeleteAttachmentReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_attachment_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteAttachmentReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteAttachmentReply) ProtoMessage() {}

func (x *DeleteAttachmentReply) ProtoReflect() protoreflect.Message {
	mi := &file_attachment_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteAttachmentReply.ProtoReflect.Descriptor instead.
func (*DeleteAttachmentReply) Descriptor() ([]byte, []int) {
	return file_attachment_service_proto_rawDescGZIP(), []int{5}
}

func (x *DeleteAttachmentReply) GetResult() *DeleteAttachmentResult {
	if x != nil {
		return x.Result
	}
	return nil
}

type DeleteAttachmentResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteAttachmentResult) Reset() {
	*x = DeleteAttachmentResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_attachment_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteAttachmentResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteAttachmentResult) ProtoMessage() {}

func (x *DeleteAttachmentResult) ProtoReflect() protoreflect.Message {
	mi := &file_attachment_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteAttachmentResult.ProtoReflect.Descriptor instead.
func (*DeleteAttachmentResult) Descriptor() ([]byte, []int) {
	return file_attachment_service_proto_rawDescGZIP(), []int{6}
}

var File_attachment_service_proto protoreflect.FileDescriptor

var file_attachment_service_proto_rawDesc = []byte{
	0x0a, 0x18, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x2d, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x68, 0x6f, 0x72,
	0x75, 0x73, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61,
	0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70,
	0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x61,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x26, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x02, 0x69, 0x64, 0x22, 0x40, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x41, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x2a, 0x0a,
	0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e,
	0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e,
	0x74, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0xa1, 0x01, 0x0a, 0x18, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x49, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x41, 0x0a, 0x0b, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1f, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x52, 0x0b, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x22, 0xc7, 0x01,
	0x0a, 0x17, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65,
	0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x66,
	0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66,
	0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x20, 0x0a,
	0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2a, 0x0a, 0x10, 0x64,
	0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x43,
	0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x22, 0x29, 0x0a, 0x17, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02,
	0x69, 0x64, 0x22, 0x4f, 0x0a, 0x15, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x41, 0x74, 0x74, 0x61,
	0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x36, 0x0a, 0x06, 0x72,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x68,
	0x6f, 0x72, 0x75, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x41, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x06, 0x72, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x22, 0x18, 0x0a, 0x16, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x41, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x32, 0xd2, 0x04,
	0x0a, 0x11, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0xba, 0x01, 0x0a, 0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x20, 0x2e, 0x63, 0x68, 0x6f, 0x72,
	0x75, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d,
	0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x6b, 0x92, 0x41, 0x45, 0x0a, 0x0b, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68,
	0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x20, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x1a, 0x23, 0x54, 0x68, 0x69, 0x73, 0x20, 0x65,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x20, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x73, 0x20,
	0x61, 0x6e, 0x20, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x1d, 0x3a, 0x01, 0x2a, 0x22, 0x18, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73,
	0x74, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x73,
	0x12, 0xb8, 0x01, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65,
	0x6e, 0x74, 0x12, 0x1c, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x41,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1a, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x6d, 0x92, 0x41,
	0x45, 0x0a, 0x0b, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x11,
	0x47, 0x65, 0x74, 0x20, 0x61, 0x6e, 0x20, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e,
	0x74, 0x1a, 0x23, 0x54, 0x68, 0x69, 0x73, 0x20, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x20, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x73, 0x20, 0x61, 0x6e, 0x20, 0x61, 0x74, 0x74, 0x61,
	0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1f, 0x12, 0x1d, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0xc4, 0x01, 0x0a, 0x10,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74,
	0x12, 0x1f, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1d, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x22, 0x70, 0x92, 0x41, 0x48, 0x0a, 0x0b, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e,
	0x74, 0x73, 0x12, 0x14, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x20, 0x61, 0x6e, 0x20, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x1a, 0x23, 0x54, 0x68, 0x69, 0x73, 0x20, 0x65,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x20, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x73, 0x20,
	0x61, 0x6e, 0x20, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x1f, 0x2a, 0x1d, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x74, 0x2f, 0x76,
	0x31, 0x2f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x7b, 0x69,
	0x64, 0x7d, 0x42, 0xe4, 0x01, 0x92, 0x41, 0xd6, 0x01, 0x12, 0x7e, 0x0a, 0x19, 0x63, 0x68, 0x6f,
	0x72, 0x75, 0x73, 0x20, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x20, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x22, 0x5c, 0x0a, 0x19, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73,
	0x20, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x20, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x2c, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x43, 0x48, 0x4f, 0x52, 0x55, 0x53, 0x2d, 0x54,
	0x52, 0x45, 0x2f, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2d, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e,
	0x64, 0x1a, 0x11, 0x64, 0x65, 0x76, 0x40, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2d, 0x74, 0x72,
	0x65, 0x2e, 0x63, 0x68, 0x32, 0x03, 0x31, 0x2e, 0x30, 0x2a, 0x01, 0x01, 0x32, 0x10, 0x61, 0x70,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x3a, 0x10,
	0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e,
	0x5a, 0x1f, 0x0a, 0x1d, 0x0a, 0x06, 0x42, 0x65, 0x61, 0x72, 0x65, 0x72, 0x12, 0x13, 0x08, 0x02,
	0x1a, 0x0d, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20,
	0x02, 0x62, 0x0c, 0x0a, 0x0a, 0x0a, 0x06, 0x42, 0x65, 0x61, 0x72, 0x65, 0x72, 0x12, 0x00, 0x5a,
	0x08, 0x2e, 0x3b, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_attachment_service_proto_rawDescOnce sync.Once
	file_attachment_service_proto_rawDescData = file_attachment_service_proto_rawDesc
)

func file_attachment_service_proto_rawDescGZIP() []byte {
	file_attachment_service_proto_rawDescOnce.Do(func() {
		file_attachment_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_attachment_service_proto_rawDescData)
	})
	return file_attachment_service_proto_rawDescData
}

var file_attachment_service_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_attachment_service_proto_goTypes = []interface{}{
	(*GetAttachmentRequest)(nil),     // 0: chorus.GetAttachmentRequest
	(*GetAttachmentReply)(nil),       // 1: chorus.GetAttachmentReply
	(*CreateAttachmentsRequest)(nil), // 2: chorus.CreateAttachmentsRequest
	(*CreateAttachmentRequest)(nil),  // 3: chorus.CreateAttachmentRequest
	(*DeleteAttachmentRequest)(nil),  // 4: chorus.DeleteAttachmentRequest
	(*DeleteAttachmentReply)(nil),    // 5: chorus.DeleteAttachmentReply
	(*DeleteAttachmentResult)(nil),   // 6: chorus.DeleteAttachmentResult
	(*Attachment)(nil),               // 7: chorus.Attachment
	(*empty.Empty)(nil),              // 8: google.protobuf.Empty
}
var file_attachment_service_proto_depIdxs = []int32{
	7, // 0: chorus.GetAttachmentReply.result:type_name -> chorus.Attachment
	3, // 1: chorus.CreateAttachmentsRequest.attachments:type_name -> chorus.CreateAttachmentRequest
	6, // 2: chorus.DeleteAttachmentReply.result:type_name -> chorus.DeleteAttachmentResult
	2, // 3: chorus.AttachmentService.CreateAttachments:input_type -> chorus.CreateAttachmentsRequest
	0, // 4: chorus.AttachmentService.GetAttachment:input_type -> chorus.GetAttachmentRequest
	4, // 5: chorus.AttachmentService.DeleteAttachment:input_type -> chorus.DeleteAttachmentRequest
	8, // 6: chorus.AttachmentService.CreateAttachments:output_type -> google.protobuf.Empty
	1, // 7: chorus.AttachmentService.GetAttachment:output_type -> chorus.GetAttachmentReply
	5, // 8: chorus.AttachmentService.DeleteAttachment:output_type -> chorus.DeleteAttachmentReply
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_attachment_service_proto_init() }
func file_attachment_service_proto_init() {
	if File_attachment_service_proto != nil {
		return
	}
	file_attachment_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_attachment_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetAttachmentRequest); i {
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
		file_attachment_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetAttachmentReply); i {
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
		file_attachment_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateAttachmentsRequest); i {
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
		file_attachment_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateAttachmentRequest); i {
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
		file_attachment_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteAttachmentRequest); i {
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
		file_attachment_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteAttachmentReply); i {
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
		file_attachment_service_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteAttachmentResult); i {
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
			RawDescriptor: file_attachment_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_attachment_service_proto_goTypes,
		DependencyIndexes: file_attachment_service_proto_depIdxs,
		MessageInfos:      file_attachment_service_proto_msgTypes,
	}.Build()
	File_attachment_service_proto = out.File
	file_attachment_service_proto_rawDesc = nil
	file_attachment_service_proto_goTypes = nil
	file_attachment_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// AttachmentServiceClient is the client API for AttachmentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AttachmentServiceClient interface {
	CreateAttachments(ctx context.Context, in *CreateAttachmentsRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	GetAttachment(ctx context.Context, in *GetAttachmentRequest, opts ...grpc.CallOption) (*GetAttachmentReply, error)
	DeleteAttachment(ctx context.Context, in *DeleteAttachmentRequest, opts ...grpc.CallOption) (*DeleteAttachmentReply, error)
}

type attachmentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAttachmentServiceClient(cc grpc.ClientConnInterface) AttachmentServiceClient {
	return &attachmentServiceClient{cc}
}

func (c *attachmentServiceClient) CreateAttachments(ctx context.Context, in *CreateAttachmentsRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/chorus.AttachmentService/CreateAttachments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *attachmentServiceClient) GetAttachment(ctx context.Context, in *GetAttachmentRequest, opts ...grpc.CallOption) (*GetAttachmentReply, error) {
	out := new(GetAttachmentReply)
	err := c.cc.Invoke(ctx, "/chorus.AttachmentService/GetAttachment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *attachmentServiceClient) DeleteAttachment(ctx context.Context, in *DeleteAttachmentRequest, opts ...grpc.CallOption) (*DeleteAttachmentReply, error) {
	out := new(DeleteAttachmentReply)
	err := c.cc.Invoke(ctx, "/chorus.AttachmentService/DeleteAttachment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AttachmentServiceServer is the server API for AttachmentService service.
type AttachmentServiceServer interface {
	CreateAttachments(context.Context, *CreateAttachmentsRequest) (*empty.Empty, error)
	GetAttachment(context.Context, *GetAttachmentRequest) (*GetAttachmentReply, error)
	DeleteAttachment(context.Context, *DeleteAttachmentRequest) (*DeleteAttachmentReply, error)
}

// UnimplementedAttachmentServiceServer can be embedded to have forward compatible implementations.
type UnimplementedAttachmentServiceServer struct {
}

func (*UnimplementedAttachmentServiceServer) CreateAttachments(context.Context, *CreateAttachmentsRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAttachments not implemented")
}
func (*UnimplementedAttachmentServiceServer) GetAttachment(context.Context, *GetAttachmentRequest) (*GetAttachmentReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAttachment not implemented")
}
func (*UnimplementedAttachmentServiceServer) DeleteAttachment(context.Context, *DeleteAttachmentRequest) (*DeleteAttachmentReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAttachment not implemented")
}

func RegisterAttachmentServiceServer(s *grpc.Server, srv AttachmentServiceServer) {
	s.RegisterService(&_AttachmentService_serviceDesc, srv)
}

func _AttachmentService_CreateAttachments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAttachmentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AttachmentServiceServer).CreateAttachments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chorus.AttachmentService/CreateAttachments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AttachmentServiceServer).CreateAttachments(ctx, req.(*CreateAttachmentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AttachmentService_GetAttachment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAttachmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AttachmentServiceServer).GetAttachment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chorus.AttachmentService/GetAttachment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AttachmentServiceServer).GetAttachment(ctx, req.(*GetAttachmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AttachmentService_DeleteAttachment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAttachmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AttachmentServiceServer).DeleteAttachment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chorus.AttachmentService/DeleteAttachment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AttachmentServiceServer).DeleteAttachment(ctx, req.(*DeleteAttachmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _AttachmentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "chorus.AttachmentService",
	HandlerType: (*AttachmentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAttachments",
			Handler:    _AttachmentService_CreateAttachments_Handler,
		},
		{
			MethodName: "GetAttachment",
			Handler:    _AttachmentService_GetAttachment_Handler,
		},
		{
			MethodName: "DeleteAttachment",
			Handler:    _AttachmentService_DeleteAttachment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "attachment-service.proto",
}