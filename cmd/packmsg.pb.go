// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.19.1
// source: packmsg.proto

//protoc --go_out=. packmsg.proto
//Imports
//protoc --js_out=library=myproto_libs,binary:./js-survive packmsg.proto
//CommonJS
//protoc --js_out=import_style=commonjs,binary:./js-survive packmsg.proto

package cmd

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

// 网关数据
type Status int32

const (
	Status_Init Status = 0
	// Connect = 1; 不再需要单独的Connect;在Live中自动判断
	Status_Close Status = 2
	Status_Live  Status = 3
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "Init",
		2: "Close",
		3: "Live",
	}
	Status_value = map[string]int32{
		"Init":  0,
		"Close": 2,
		"Live":  3,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_packmsg_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_packmsg_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_packmsg_proto_rawDescGZIP(), []int{0}
}

type PackMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cmd    int32  `protobuf:"varint,1,opt,name=cmd,proto3" json:"cmd,omitempty"`      //Cmd
	Seq    int32  `protobuf:"varint,2,opt,name=seq,proto3" json:"seq,omitempty"`      //序列号
	Ret    int32  `protobuf:"varint,3,opt,name=ret,proto3" json:"ret,omitempty"`      //返回值
	Binary []byte `protobuf:"bytes,4,opt,name=binary,proto3" json:"binary,omitempty"` //包内容
	Tip    string `protobuf:"bytes,5,opt,name=tip,proto3" json:"tip,omitempty"`       //服务器传来的提示消息内容
}

func (x *PackMsg) Reset() {
	*x = PackMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_packmsg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PackMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PackMsg) ProtoMessage() {}

func (x *PackMsg) ProtoReflect() protoreflect.Message {
	mi := &file_packmsg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PackMsg.ProtoReflect.Descriptor instead.
func (*PackMsg) Descriptor() ([]byte, []int) {
	return file_packmsg_proto_rawDescGZIP(), []int{0}
}

func (x *PackMsg) GetCmd() int32 {
	if x != nil {
		return x.Cmd
	}
	return 0
}

func (x *PackMsg) GetSeq() int32 {
	if x != nil {
		return x.Seq
	}
	return 0
}

func (x *PackMsg) GetRet() int32 {
	if x != nil {
		return x.Ret
	}
	return 0
}

func (x *PackMsg) GetBinary() []byte {
	if x != nil {
		return x.Binary
	}
	return nil
}

func (x *PackMsg) GetTip() string {
	if x != nil {
		return x.Tip
	}
	return ""
}

type AgentData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"` //服务ID
	// uint32 sid = 2;//连接的索引
	CliId  uint64 `protobuf:"varint,3,opt,name=cliId,proto3" json:"cliId,omitempty"` //客户端ID
	Status Status `protobuf:"varint,4,opt,name=status,proto3,enum=cmd.Status" json:"status,omitempty"`
	// string ip = 5;//仅连接的时候传值
	Data   []byte `protobuf:"bytes,6,opt,name=data,proto3" json:"data,omitempty"`
	Close  bool   `protobuf:"varint,7,opt,name=close,proto3" json:"close,omitempty"`   //是否关闭连接=>来自服务器
	Ws     string `protobuf:"bytes,8,opt,name=ws,proto3" json:"ws,omitempty"`          //接入地址
	Weight int32  `protobuf:"varint,9,opt,name=weight,proto3" json:"weight,omitempty"` //权重
}

func (x *AgentData) Reset() {
	*x = AgentData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_packmsg_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AgentData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AgentData) ProtoMessage() {}

func (x *AgentData) ProtoReflect() protoreflect.Message {
	mi := &file_packmsg_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AgentData.ProtoReflect.Descriptor instead.
func (*AgentData) Descriptor() ([]byte, []int) {
	return file_packmsg_proto_rawDescGZIP(), []int{1}
}

func (x *AgentData) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *AgentData) GetCliId() uint64 {
	if x != nil {
		return x.CliId
	}
	return 0
}

func (x *AgentData) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_Init
}

func (x *AgentData) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *AgentData) GetClose() bool {
	if x != nil {
		return x.Close
	}
	return false
}

func (x *AgentData) GetWs() string {
	if x != nil {
		return x.Ws
	}
	return ""
}

func (x *AgentData) GetWeight() int32 {
	if x != nil {
		return x.Weight
	}
	return 0
}

var File_packmsg_proto protoreflect.FileDescriptor

var file_packmsg_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x70, 0x61, 0x63, 0x6b, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x03, 0x63, 0x6d, 0x64, 0x22, 0x69, 0x0a, 0x07, 0x50, 0x61, 0x63, 0x6b, 0x4d, 0x73, 0x67, 0x12,
	0x10, 0x0a, 0x03, 0x63, 0x6d, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x63, 0x6d,
	0x64, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x65, 0x71, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03,
	0x73, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x03, 0x72, 0x65, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x74, 0x69, 0x70, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x69, 0x70, 0x22,
	0xa8, 0x01, 0x0a, 0x09, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x44, 0x61, 0x74, 0x61, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a,
	0x05, 0x63, 0x6c, 0x69, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x63, 0x6c,
	0x69, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x0b, 0x2e, 0x63, 0x6d, 0x64, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x14, 0x0a, 0x05,
	0x63, 0x6c, 0x6f, 0x73, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x63, 0x6c, 0x6f,
	0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x77, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x77, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x09, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x2a, 0x27, 0x0a, 0x06, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x08, 0x0a, 0x04, 0x49, 0x6e, 0x69, 0x74, 0x10, 0x00, 0x12, 0x09,
	0x0a, 0x05, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x69, 0x76,
	0x65, 0x10, 0x03, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2f, 0x63, 0x6d, 0x64, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_packmsg_proto_rawDescOnce sync.Once
	file_packmsg_proto_rawDescData = file_packmsg_proto_rawDesc
)

func file_packmsg_proto_rawDescGZIP() []byte {
	file_packmsg_proto_rawDescOnce.Do(func() {
		file_packmsg_proto_rawDescData = protoimpl.X.CompressGZIP(file_packmsg_proto_rawDescData)
	})
	return file_packmsg_proto_rawDescData
}

var file_packmsg_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_packmsg_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_packmsg_proto_goTypes = []interface{}{
	(Status)(0),       // 0: cmd.Status
	(*PackMsg)(nil),   // 1: cmd.PackMsg
	(*AgentData)(nil), // 2: cmd.AgentData
}
var file_packmsg_proto_depIdxs = []int32{
	0, // 0: cmd.AgentData.status:type_name -> cmd.Status
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_packmsg_proto_init() }
func file_packmsg_proto_init() {
	if File_packmsg_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_packmsg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PackMsg); i {
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
		file_packmsg_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AgentData); i {
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
			RawDescriptor: file_packmsg_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_packmsg_proto_goTypes,
		DependencyIndexes: file_packmsg_proto_depIdxs,
		EnumInfos:         file_packmsg_proto_enumTypes,
		MessageInfos:      file_packmsg_proto_msgTypes,
	}.Build()
	File_packmsg_proto = out.File
	file_packmsg_proto_rawDesc = nil
	file_packmsg_proto_goTypes = nil
	file_packmsg_proto_depIdxs = nil
}
