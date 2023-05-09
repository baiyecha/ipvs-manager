// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.12
// source: grpc/proto/ipvs_service.proto

package __

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

type IpvsListRequeste struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ip string `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
}

func (x *IpvsListRequeste) Reset() {
	*x = IpvsListRequeste{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_ipvs_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IpvsListRequeste) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IpvsListRequeste) ProtoMessage() {}

func (x *IpvsListRequeste) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_ipvs_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IpvsListRequeste.ProtoReflect.Descriptor instead.
func (*IpvsListRequeste) Descriptor() ([]byte, []int) {
	return file_grpc_proto_ipvs_service_proto_rawDescGZIP(), []int{0}
}

func (x *IpvsListRequeste) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

type IpvsListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	List []*Ipvs `protobuf:"bytes,1,rep,name=list,proto3" json:"list,omitempty"`
}

func (x *IpvsListResponse) Reset() {
	*x = IpvsListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_ipvs_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IpvsListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IpvsListResponse) ProtoMessage() {}

func (x *IpvsListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_ipvs_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IpvsListResponse.ProtoReflect.Descriptor instead.
func (*IpvsListResponse) Descriptor() ([]byte, []int) {
	return file_grpc_proto_ipvs_service_proto_rawDescGZIP(), []int{1}
}

func (x *IpvsListResponse) GetList() []*Ipvs {
	if x != nil {
		return x.List
	}
	return nil
}

type Backend struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addr         string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	Weight       int64  `protobuf:"varint,2,opt,name=weight,proto3" json:"weight,omitempty"`
	Status       int64  `protobuf:"varint,3,opt,name=status,proto3" json:"status,omitempty"`
	CheckType    int64  `protobuf:"varint,4,opt,name=check_type,json=checkType,proto3" json:"check_type,omitempty"`
	CheckInfo    string `protobuf:"bytes,5,opt,name=check_info,json=checkInfo,proto3" json:"check_info,omitempty"`
	CheckResType int64  `protobuf:"varint,6,opt,name=check_res_type,json=checkResType,proto3" json:"check_res_type,omitempty"`
	CheckRes     string `protobuf:"bytes,7,opt,name=check_res,json=checkRes,proto3" json:"check_res,omitempty"`
}

func (x *Backend) Reset() {
	*x = Backend{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_ipvs_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Backend) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Backend) ProtoMessage() {}

func (x *Backend) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_ipvs_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Backend.ProtoReflect.Descriptor instead.
func (*Backend) Descriptor() ([]byte, []int) {
	return file_grpc_proto_ipvs_service_proto_rawDescGZIP(), []int{2}
}

func (x *Backend) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *Backend) GetWeight() int64 {
	if x != nil {
		return x.Weight
	}
	return 0
}

func (x *Backend) GetStatus() int64 {
	if x != nil {
		return x.Status
	}
	return 0
}

func (x *Backend) GetCheckType() int64 {
	if x != nil {
		return x.CheckType
	}
	return 0
}

func (x *Backend) GetCheckInfo() string {
	if x != nil {
		return x.CheckInfo
	}
	return ""
}

func (x *Backend) GetCheckResType() int64 {
	if x != nil {
		return x.CheckResType
	}
	return 0
}

func (x *Backend) GetCheckRes() string {
	if x != nil {
		return x.CheckRes
	}
	return ""
}

type Ipvs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vip       string     `protobuf:"bytes,1,opt,name=vip,proto3" json:"vip,omitempty"`
	Backends  []*Backend `protobuf:"bytes,2,rep,name=backends,proto3" json:"backends,omitempty"`
	Protocol  string     `protobuf:"bytes,3,opt,name=protocol,proto3" json:"protocol,omitempty"`
	SchedName string     `protobuf:"bytes,4,opt,name=sched_name,json=schedName,proto3" json:"sched_name,omitempty"`
}

func (x *Ipvs) Reset() {
	*x = Ipvs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_ipvs_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Ipvs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Ipvs) ProtoMessage() {}

func (x *Ipvs) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_ipvs_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Ipvs.ProtoReflect.Descriptor instead.
func (*Ipvs) Descriptor() ([]byte, []int) {
	return file_grpc_proto_ipvs_service_proto_rawDescGZIP(), []int{3}
}

func (x *Ipvs) GetVip() string {
	if x != nil {
		return x.Vip
	}
	return ""
}

func (x *Ipvs) GetBackends() []*Backend {
	if x != nil {
		return x.Backends
	}
	return nil
}

func (x *Ipvs) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *Ipvs) GetSchedName() string {
	if x != nil {
		return x.SchedName
	}
	return ""
}

var File_grpc_proto_ipvs_service_proto protoreflect.FileDescriptor

var file_grpc_proto_ipvs_service_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x69, 0x70, 0x76,
	0x73, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x04, 0x69, 0x70, 0x76, 0x73, 0x22, 0x22, 0x0a, 0x10, 0x49, 0x70, 0x76, 0x73, 0x4c, 0x69, 0x73,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x22, 0x32, 0x0a, 0x10, 0x49, 0x70, 0x76,
	0x73, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a,
	0x04, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x69, 0x70,
	0x76, 0x73, 0x2e, 0x49, 0x70, 0x76, 0x73, 0x52, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x22, 0xce, 0x01,
	0x0a, 0x07, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x64, 0x64,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72, 0x12, 0x16, 0x0a,
	0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x77,
	0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1d, 0x0a,
	0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x09, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1d, 0x0a, 0x0a,
	0x63, 0x68, 0x65, 0x63, 0x6b, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x24, 0x0a, 0x0e, 0x63,
	0x68, 0x65, 0x63, 0x6b, 0x5f, 0x72, 0x65, 0x73, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x0c, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x5f, 0x72, 0x65, 0x73, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x22, 0x7e,
	0x0a, 0x04, 0x49, 0x70, 0x76, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x76, 0x69, 0x70, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x76, 0x69, 0x70, 0x12, 0x29, 0x0a, 0x08, 0x62, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x69, 0x70, 0x76,
	0x73, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x52, 0x08, 0x62, 0x61, 0x63, 0x6b, 0x65,
	0x6e, 0x64, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12,
	0x1d, 0x0a, 0x0a, 0x73, 0x63, 0x68, 0x65, 0x64, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x63, 0x68, 0x65, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x32, 0x4f,
	0x0a, 0x0f, 0x49, 0x70, 0x76, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x3c, 0x0a, 0x08, 0x49, 0x70, 0x76, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x16, 0x2e,
	0x69, 0x70, 0x76, 0x73, 0x2e, 0x49, 0x70, 0x76, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x65, 0x1a, 0x16, 0x2e, 0x69, 0x70, 0x76, 0x73, 0x2e, 0x49, 0x70, 0x76,
	0x73, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42,
	0x04, 0x5a, 0x02, 0x2e, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_grpc_proto_ipvs_service_proto_rawDescOnce sync.Once
	file_grpc_proto_ipvs_service_proto_rawDescData = file_grpc_proto_ipvs_service_proto_rawDesc
)

func file_grpc_proto_ipvs_service_proto_rawDescGZIP() []byte {
	file_grpc_proto_ipvs_service_proto_rawDescOnce.Do(func() {
		file_grpc_proto_ipvs_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_grpc_proto_ipvs_service_proto_rawDescData)
	})
	return file_grpc_proto_ipvs_service_proto_rawDescData
}

var file_grpc_proto_ipvs_service_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_grpc_proto_ipvs_service_proto_goTypes = []interface{}{
	(*IpvsListRequeste)(nil), // 0: ipvs.IpvsListRequeste
	(*IpvsListResponse)(nil), // 1: ipvs.IpvsListResponse
	(*Backend)(nil),          // 2: ipvs.Backend
	(*Ipvs)(nil),             // 3: ipvs.Ipvs
}
var file_grpc_proto_ipvs_service_proto_depIdxs = []int32{
	3, // 0: ipvs.IpvsListResponse.list:type_name -> ipvs.Ipvs
	2, // 1: ipvs.Ipvs.backends:type_name -> ipvs.Backend
	0, // 2: ipvs.IpvsListService.IpvsList:input_type -> ipvs.IpvsListRequeste
	1, // 3: ipvs.IpvsListService.IpvsList:output_type -> ipvs.IpvsListResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_grpc_proto_ipvs_service_proto_init() }
func file_grpc_proto_ipvs_service_proto_init() {
	if File_grpc_proto_ipvs_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_grpc_proto_ipvs_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IpvsListRequeste); i {
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
		file_grpc_proto_ipvs_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IpvsListResponse); i {
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
		file_grpc_proto_ipvs_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Backend); i {
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
		file_grpc_proto_ipvs_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Ipvs); i {
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
			RawDescriptor: file_grpc_proto_ipvs_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_grpc_proto_ipvs_service_proto_goTypes,
		DependencyIndexes: file_grpc_proto_ipvs_service_proto_depIdxs,
		MessageInfos:      file_grpc_proto_ipvs_service_proto_msgTypes,
	}.Build()
	File_grpc_proto_ipvs_service_proto = out.File
	file_grpc_proto_ipvs_service_proto_rawDesc = nil
	file_grpc_proto_ipvs_service_proto_goTypes = nil
	file_grpc_proto_ipvs_service_proto_depIdxs = nil
}
