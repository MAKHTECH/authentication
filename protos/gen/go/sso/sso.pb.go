// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v6.30.0--rc1
// source: sso/sso.proto

package ssov1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_sso_sso_proto protoreflect.FileDescriptor

var file_sso_sso_proto_rawDesc = string([]byte{
	0x0a, 0x0d, 0x73, 0x73, 0x6f, 0x2f, 0x73, 0x73, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x04, 0x61, 0x75, 0x74, 0x68, 0x1a, 0x0e, 0x73, 0x73, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0e, 0x73, 0x73, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x73, 0x73, 0x6f, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xb0, 0x02, 0x0a, 0x04, 0x41, 0x75, 0x74, 0x68,
	0x12, 0x39, 0x0a, 0x08, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x15, 0x2e, 0x61,
	0x75, 0x74, 0x68, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x05, 0x4c,
	0x6f, 0x67, 0x69, 0x6e, 0x12, 0x12, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x4c, 0x6f, 0x67, 0x69,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x45, 0x0a,
	0x0c, 0x52, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x19, 0x2e,
	0x61, 0x75, 0x74, 0x68, 0x2e, 0x52, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x52, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3f, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x73, 0x12, 0x17, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x61, 0x75,
	0x74, 0x68, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a, 0x06, 0x4c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x12,
	0x13, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x4c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x4c, 0x6f, 0x67, 0x6f,
	0x75, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x47, 0x0a, 0x04, 0x55, 0x73,
	0x65, 0x72, 0x12, 0x3f, 0x0a, 0x0a, 0x41, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x6c, 0x65,
	0x12, 0x17, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x41, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x6f,
	0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x61, 0x75, 0x74, 0x68,
	0x2e, 0x41, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x42, 0x1b, 0x8a, 0xb5, 0x18, 0x01, 0x31, 0x5a, 0x14, 0x6d, 0x61, 0x6b, 0x68,
	0x6b, 0x65, 0x74, 0x73, 0x2e, 0x67, 0x6f, 0x2e, 0x76, 0x31, 0x3b, 0x73, 0x73, 0x6f, 0x76, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var file_sso_sso_proto_goTypes = []any{
	(*RegisterRequest)(nil),      // 0: auth.RegisterRequest
	(*LoginRequest)(nil),         // 1: auth.LoginRequest
	(*RefreshTokenRequest)(nil),  // 2: auth.RefreshTokenRequest
	(*GetDevicesRequest)(nil),    // 3: auth.GetDevicesRequest
	(*LogoutRequest)(nil),        // 4: auth.LogoutRequest
	(*AssignRoleRequest)(nil),    // 5: auth.AssignRoleRequest
	(*RegisterResponse)(nil),     // 6: auth.RegisterResponse
	(*LoginResponse)(nil),        // 7: auth.LoginResponse
	(*RefreshTokenResponse)(nil), // 8: auth.RefreshTokenResponse
	(*GetDevicesResponse)(nil),   // 9: auth.GetDevicesResponse
	(*LogoutResponse)(nil),       // 10: auth.LogoutResponse
	(*AssignRoleResponse)(nil),   // 11: auth.AssignRoleResponse
}
var file_sso_sso_proto_depIdxs = []int32{
	0,  // 0: auth.Auth.Register:input_type -> auth.RegisterRequest
	1,  // 1: auth.Auth.Login:input_type -> auth.LoginRequest
	2,  // 2: auth.Auth.RefreshToken:input_type -> auth.RefreshTokenRequest
	3,  // 3: auth.Auth.GetDevices:input_type -> auth.GetDevicesRequest
	4,  // 4: auth.Auth.Logout:input_type -> auth.LogoutRequest
	5,  // 5: auth.User.AssignRole:input_type -> auth.AssignRoleRequest
	6,  // 6: auth.Auth.Register:output_type -> auth.RegisterResponse
	7,  // 7: auth.Auth.Login:output_type -> auth.LoginResponse
	8,  // 8: auth.Auth.RefreshToken:output_type -> auth.RefreshTokenResponse
	9,  // 9: auth.Auth.GetDevices:output_type -> auth.GetDevicesResponse
	10, // 10: auth.Auth.Logout:output_type -> auth.LogoutResponse
	11, // 11: auth.User.AssignRole:output_type -> auth.AssignRoleResponse
	6,  // [6:12] is the sub-list for method output_type
	0,  // [0:6] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_sso_sso_proto_init() }
func file_sso_sso_proto_init() {
	if File_sso_sso_proto != nil {
		return
	}
	file_sso_auth_proto_init()
	file_sso_user_proto_init()
	file_sso_option_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_sso_sso_proto_rawDesc), len(file_sso_sso_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_sso_sso_proto_goTypes,
		DependencyIndexes: file_sso_sso_proto_depIdxs,
	}.Build()
	File_sso_sso_proto = out.File
	file_sso_sso_proto_goTypes = nil
	file_sso_sso_proto_depIdxs = nil
}
