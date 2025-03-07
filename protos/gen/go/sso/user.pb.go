// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v6.30.0--rc1
// source: sso/user.proto

package ssov1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Role int32

const (
	Role_USER  Role = 0
	Role_ADMIN Role = 1
)

// Enum value maps for Role.
var (
	Role_name = map[int32]string{
		0: "USER",
		1: "ADMIN",
	}
	Role_value = map[string]int32{
		"USER":  0,
		"ADMIN": 1,
	}
)

func (x Role) Enum() *Role {
	p := new(Role)
	*p = x
	return p
}

func (x Role) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Role) Descriptor() protoreflect.EnumDescriptor {
	return file_sso_user_proto_enumTypes[0].Descriptor()
}

func (Role) Type() protoreflect.EnumType {
	return &file_sso_user_proto_enumTypes[0]
}

func (x Role) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Role.Descriptor instead.
func (Role) EnumDescriptor() ([]byte, []int) {
	return file_sso_user_proto_rawDescGZIP(), []int{0}
}

type AssignRoleRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        uint32                 `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	AppId         int32                  `protobuf:"varint,2,opt,name=app_id,json=appId,proto3" json:"app_id,omitempty"`
	Role          Role                   `protobuf:"varint,3,opt,name=role,proto3,enum=auth.Role" json:"role,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AssignRoleRequest) Reset() {
	*x = AssignRoleRequest{}
	mi := &file_sso_user_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AssignRoleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssignRoleRequest) ProtoMessage() {}

func (x *AssignRoleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_sso_user_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssignRoleRequest.ProtoReflect.Descriptor instead.
func (*AssignRoleRequest) Descriptor() ([]byte, []int) {
	return file_sso_user_proto_rawDescGZIP(), []int{0}
}

func (x *AssignRoleRequest) GetUserId() uint32 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *AssignRoleRequest) GetAppId() int32 {
	if x != nil {
		return x.AppId
	}
	return 0
}

func (x *AssignRoleRequest) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_USER
}

type AssignRoleResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AssignRoleResponse) Reset() {
	*x = AssignRoleResponse{}
	mi := &file_sso_user_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AssignRoleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssignRoleResponse) ProtoMessage() {}

func (x *AssignRoleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_sso_user_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssignRoleResponse.ProtoReflect.Descriptor instead.
func (*AssignRoleResponse) Descriptor() ([]byte, []int) {
	return file_sso_user_proto_rawDescGZIP(), []int{1}
}

func (x *AssignRoleResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_sso_user_proto protoreflect.FileDescriptor

var file_sso_user_proto_rawDesc = string([]byte{
	0x0a, 0x0e, 0x73, 0x73, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x04, 0x61, 0x75, 0x74, 0x68, 0x22, 0x63, 0x0a, 0x11, 0x41, 0x73, 0x73, 0x69, 0x67, 0x6e,
	0x52, 0x6f, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x75,
	0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x75, 0x73,
	0x65, 0x72, 0x49, 0x64, 0x12, 0x15, 0x0a, 0x06, 0x61, 0x70, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x61, 0x70, 0x70, 0x49, 0x64, 0x12, 0x1e, 0x0a, 0x04, 0x72,
	0x6f, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0a, 0x2e, 0x61, 0x75, 0x74, 0x68,
	0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x22, 0x2e, 0x0a, 0x12, 0x41,
	0x73, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x2a, 0x1b, 0x0a, 0x04, 0x52,
	0x6f, 0x6c, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x55, 0x53, 0x45, 0x52, 0x10, 0x00, 0x12, 0x09, 0x0a,
	0x05, 0x41, 0x44, 0x4d, 0x49, 0x4e, 0x10, 0x01, 0x42, 0x16, 0x5a, 0x14, 0x6d, 0x61, 0x6b, 0x68,
	0x6b, 0x65, 0x74, 0x73, 0x2e, 0x67, 0x6f, 0x2e, 0x76, 0x31, 0x3b, 0x73, 0x73, 0x6f, 0x76, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_sso_user_proto_rawDescOnce sync.Once
	file_sso_user_proto_rawDescData []byte
)

func file_sso_user_proto_rawDescGZIP() []byte {
	file_sso_user_proto_rawDescOnce.Do(func() {
		file_sso_user_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_sso_user_proto_rawDesc), len(file_sso_user_proto_rawDesc)))
	})
	return file_sso_user_proto_rawDescData
}

var file_sso_user_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_sso_user_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_sso_user_proto_goTypes = []any{
	(Role)(0),                  // 0: auth.Role
	(*AssignRoleRequest)(nil),  // 1: auth.AssignRoleRequest
	(*AssignRoleResponse)(nil), // 2: auth.AssignRoleResponse
}
var file_sso_user_proto_depIdxs = []int32{
	0, // 0: auth.AssignRoleRequest.role:type_name -> auth.Role
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_sso_user_proto_init() }
func file_sso_user_proto_init() {
	if File_sso_user_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_sso_user_proto_rawDesc), len(file_sso_user_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_sso_user_proto_goTypes,
		DependencyIndexes: file_sso_user_proto_depIdxs,
		EnumInfos:         file_sso_user_proto_enumTypes,
		MessageInfos:      file_sso_user_proto_msgTypes,
	}.Build()
	File_sso_user_proto = out.File
	file_sso_user_proto_goTypes = nil
	file_sso_user_proto_depIdxs = nil
}
