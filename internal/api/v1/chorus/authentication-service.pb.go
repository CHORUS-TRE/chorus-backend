// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.27.2
// source: authentication-service.proto

package chorus

import (
	context "context"
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

type AuthenticationReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result *AuthenticationResult `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *AuthenticationReply) Reset() {
	*x = AuthenticationReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticationReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticationReply) ProtoMessage() {}

func (x *AuthenticationReply) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticationReply.ProtoReflect.Descriptor instead.
func (*AuthenticationReply) Descriptor() ([]byte, []int) {
	return file_authentication_service_proto_rawDescGZIP(), []int{0}
}

func (x *AuthenticationReply) GetResult() *AuthenticationResult {
	if x != nil {
		return x.Result
	}
	return nil
}

type AuthenticationResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *AuthenticationResult) Reset() {
	*x = AuthenticationResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticationResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticationResult) ProtoMessage() {}

func (x *AuthenticationResult) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticationResult.ProtoReflect.Descriptor instead.
func (*AuthenticationResult) Descriptor() ([]byte, []int) {
	return file_authentication_service_proto_rawDescGZIP(), []int{1}
}

func (x *AuthenticationResult) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

var File_authentication_service_proto protoreflect.FileDescriptor

var file_authentication_service_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2d, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e,
	0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f,
	0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4b,
	0x0a, 0x13, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x34, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x41,
	0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x2c, 0x0a, 0x14, 0x41,
	0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x32, 0xda, 0x01, 0x0a, 0x15, 0x41, 0x75,
	0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0xc0, 0x01, 0x0a, 0x0c, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69,
	0x63, 0x61, 0x74, 0x65, 0x12, 0x13, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2e, 0x43, 0x72,
	0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x1a, 0x1b, 0x2e, 0x63, 0x68, 0x6f, 0x72,
	0x75, 0x73, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x7e, 0x92, 0x41, 0x4b, 0x0a, 0x15, 0x41, 0x75, 0x74,
	0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x0c, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65,
	0x1a, 0x22, 0x54, 0x68, 0x69, 0x73, 0x20, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x20,
	0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x20, 0x61, 0x20,
	0x75, 0x73, 0x65, 0x72, 0x62, 0x00, 0x88, 0xb5, 0x18, 0x01, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x26,
	0x3a, 0x01, 0x2a, 0x22, 0x21, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x74, 0x2f, 0x76,
	0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x42, 0xbe, 0x01, 0x92, 0x41, 0xb0, 0x01, 0x12, 0x86, 0x01,
	0x0a, 0x1d, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x20, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x22,
	0x60, 0x0a, 0x1d, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x20, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e,
	0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x2c, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x43, 0x48, 0x4f, 0x52, 0x55, 0x53, 0x2d, 0x54, 0x52, 0x45, 0x2f,
	0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2d, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x1a, 0x11,
	0x64, 0x65, 0x76, 0x40, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x2d, 0x74, 0x72, 0x65, 0x2e, 0x63,
	0x68, 0x32, 0x03, 0x31, 0x2e, 0x30, 0x2a, 0x01, 0x01, 0x32, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x3a, 0x10, 0x61, 0x70, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x5a, 0x08, 0x2e,
	0x3b, 0x63, 0x68, 0x6f, 0x72, 0x75, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_authentication_service_proto_rawDescOnce sync.Once
	file_authentication_service_proto_rawDescData = file_authentication_service_proto_rawDesc
)

func file_authentication_service_proto_rawDescGZIP() []byte {
	file_authentication_service_proto_rawDescOnce.Do(func() {
		file_authentication_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_authentication_service_proto_rawDescData)
	})
	return file_authentication_service_proto_rawDescData
}

var file_authentication_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_authentication_service_proto_goTypes = []interface{}{
	(*AuthenticationReply)(nil),  // 0: chorus.AuthenticationReply
	(*AuthenticationResult)(nil), // 1: chorus.AuthenticationResult
	(*Credentials)(nil),          // 2: chorus.Credentials
}
var file_authentication_service_proto_depIdxs = []int32{
	1, // 0: chorus.AuthenticationReply.result:type_name -> chorus.AuthenticationResult
	2, // 1: chorus.AuthenticationService.Authenticate:input_type -> chorus.Credentials
	0, // 2: chorus.AuthenticationService.Authenticate:output_type -> chorus.AuthenticationReply
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_authentication_service_proto_init() }
func file_authentication_service_proto_init() {
	if File_authentication_service_proto != nil {
		return
	}
	file_credentials_proto_init()
	file_public_endpoint_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_authentication_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthenticationReply); i {
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
		file_authentication_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthenticationResult); i {
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
			RawDescriptor: file_authentication_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_authentication_service_proto_goTypes,
		DependencyIndexes: file_authentication_service_proto_depIdxs,
		MessageInfos:      file_authentication_service_proto_msgTypes,
	}.Build()
	File_authentication_service_proto = out.File
	file_authentication_service_proto_rawDesc = nil
	file_authentication_service_proto_goTypes = nil
	file_authentication_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// AuthenticationServiceClient is the client API for AuthenticationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AuthenticationServiceClient interface {
	Authenticate(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*AuthenticationReply, error)
}

type authenticationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthenticationServiceClient(cc grpc.ClientConnInterface) AuthenticationServiceClient {
	return &authenticationServiceClient{cc}
}

func (c *authenticationServiceClient) Authenticate(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*AuthenticationReply, error) {
	out := new(AuthenticationReply)
	err := c.cc.Invoke(ctx, "/chorus.AuthenticationService/Authenticate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthenticationServiceServer is the server API for AuthenticationService service.
type AuthenticationServiceServer interface {
	Authenticate(context.Context, *Credentials) (*AuthenticationReply, error)
}

// UnimplementedAuthenticationServiceServer can be embedded to have forward compatible implementations.
type UnimplementedAuthenticationServiceServer struct {
}

func (*UnimplementedAuthenticationServiceServer) Authenticate(context.Context, *Credentials) (*AuthenticationReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authenticate not implemented")
}

func RegisterAuthenticationServiceServer(s *grpc.Server, srv AuthenticationServiceServer) {
	s.RegisterService(&_AuthenticationService_serviceDesc, srv)
}

func _AuthenticationService_Authenticate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Credentials)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthenticationServiceServer).Authenticate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chorus.AuthenticationService/Authenticate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthenticationServiceServer).Authenticate(ctx, req.(*Credentials))
	}
	return interceptor(ctx, in, info, handler)
}

var _AuthenticationService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "chorus.AuthenticationService",
	HandlerType: (*AuthenticationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Authenticate",
			Handler:    _AuthenticationService_Authenticate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "authentication-service.proto",
}
