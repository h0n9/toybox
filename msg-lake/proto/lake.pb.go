// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.21.5
// source: proto/lake.proto

package proto

import (
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

type SendReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id  string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Msg *Msg   `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *SendReq) Reset() {
	*x = SendReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_lake_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendReq) ProtoMessage() {}

func (x *SendReq) ProtoReflect() protoreflect.Message {
	mi := &file_proto_lake_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendReq.ProtoReflect.Descriptor instead.
func (*SendReq) Descriptor() ([]byte, []int) {
	return file_proto_lake_proto_rawDescGZIP(), []int{0}
}

func (x *SendReq) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *SendReq) GetMsg() *Msg {
	if x != nil {
		return x.Msg
	}
	return nil
}

type SendRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (x *SendRes) Reset() {
	*x = SendRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_lake_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendRes) ProtoMessage() {}

func (x *SendRes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_lake_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendRes.ProtoReflect.Descriptor instead.
func (*SendRes) Descriptor() ([]byte, []int) {
	return file_proto_lake_proto_rawDescGZIP(), []int{1}
}

func (x *SendRes) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

type RecvReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *RecvReq) Reset() {
	*x = RecvReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_lake_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecvReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecvReq) ProtoMessage() {}

func (x *RecvReq) ProtoReflect() protoreflect.Message {
	mi := &file_proto_lake_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecvReq.ProtoReflect.Descriptor instead.
func (*RecvReq) Descriptor() ([]byte, []int) {
	return file_proto_lake_proto_rawDescGZIP(), []int{2}
}

func (x *RecvReq) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type RecvRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg *Msg `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *RecvRes) Reset() {
	*x = RecvRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_lake_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecvRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecvRes) ProtoMessage() {}

func (x *RecvRes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_lake_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecvRes.ProtoReflect.Descriptor instead.
func (*RecvRes) Descriptor() ([]byte, []int) {
	return file_proto_lake_proto_rawDescGZIP(), []int{3}
}

func (x *RecvRes) GetMsg() *Msg {
	if x != nil {
		return x.Msg
	}
	return nil
}

var File_proto_lake_proto protoreflect.FileDescriptor

var file_proto_lake_proto_rawDesc = []byte{
	0x0a, 0x10, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x61, 0x6b, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x31, 0x0a, 0x07, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16,
	0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x04, 0x2e, 0x4d, 0x73,
	0x67, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x19, 0x0a, 0x07, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65,
	0x73, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f,
	0x6b, 0x22, 0x19, 0x0a, 0x07, 0x52, 0x65, 0x63, 0x76, 0x52, 0x65, 0x71, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x21, 0x0a, 0x07,
	0x52, 0x65, 0x63, 0x76, 0x52, 0x65, 0x73, 0x12, 0x16, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x04, 0x2e, 0x4d, 0x73, 0x67, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x32,
	0x40, 0x0a, 0x04, 0x4c, 0x61, 0x6b, 0x65, 0x12, 0x1a, 0x0a, 0x04, 0x53, 0x65, 0x6e, 0x64, 0x12,
	0x08, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x08, 0x2e, 0x53, 0x65, 0x6e, 0x64,
	0x52, 0x65, 0x73, 0x12, 0x1c, 0x0a, 0x04, 0x52, 0x65, 0x63, 0x76, 0x12, 0x08, 0x2e, 0x52, 0x65,
	0x63, 0x76, 0x52, 0x65, 0x71, 0x1a, 0x08, 0x2e, 0x52, 0x65, 0x63, 0x76, 0x52, 0x65, 0x73, 0x30,
	0x01, 0x42, 0x27, 0x5a, 0x25, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x68, 0x30, 0x6e, 0x39, 0x2f, 0x74, 0x6f, 0x79, 0x62, 0x6f, 0x78, 0x2f, 0x6d, 0x73, 0x67, 0x2d,
	0x6c, 0x61, 0x6b, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_proto_lake_proto_rawDescOnce sync.Once
	file_proto_lake_proto_rawDescData = file_proto_lake_proto_rawDesc
)

func file_proto_lake_proto_rawDescGZIP() []byte {
	file_proto_lake_proto_rawDescOnce.Do(func() {
		file_proto_lake_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_lake_proto_rawDescData)
	})
	return file_proto_lake_proto_rawDescData
}

var file_proto_lake_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_lake_proto_goTypes = []interface{}{
	(*SendReq)(nil), // 0: SendReq
	(*SendRes)(nil), // 1: SendRes
	(*RecvReq)(nil), // 2: RecvReq
	(*RecvRes)(nil), // 3: RecvRes
	(*Msg)(nil),     // 4: Msg
}
var file_proto_lake_proto_depIdxs = []int32{
	4, // 0: SendReq.msg:type_name -> Msg
	4, // 1: RecvRes.msg:type_name -> Msg
	0, // 2: Lake.Send:input_type -> SendReq
	2, // 3: Lake.Recv:input_type -> RecvReq
	1, // 4: Lake.Send:output_type -> SendRes
	3, // 5: Lake.Recv:output_type -> RecvRes
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_lake_proto_init() }
func file_proto_lake_proto_init() {
	if File_proto_lake_proto != nil {
		return
	}
	file_proto_msg_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_proto_lake_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendReq); i {
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
		file_proto_lake_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendRes); i {
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
		file_proto_lake_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecvReq); i {
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
		file_proto_lake_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecvRes); i {
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
			RawDescriptor: file_proto_lake_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_lake_proto_goTypes,
		DependencyIndexes: file_proto_lake_proto_depIdxs,
		MessageInfos:      file_proto_lake_proto_msgTypes,
	}.Build()
	File_proto_lake_proto = out.File
	file_proto_lake_proto_rawDesc = nil
	file_proto_lake_proto_goTypes = nil
	file_proto_lake_proto_depIdxs = nil
}
