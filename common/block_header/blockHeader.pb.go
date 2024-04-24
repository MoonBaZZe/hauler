// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.0
// source: blockHeader.proto

package block_header

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

type BlockHeaderProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version    int32  `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	PrevBlock  []byte `protobuf:"bytes,2,opt,name=prevBlock,proto3" json:"prevBlock,omitempty"`
	MerkleRoot []byte `protobuf:"bytes,3,opt,name=merkleRoot,proto3" json:"merkleRoot,omitempty"`
	Timestamp  uint32 `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Bits       uint32 `protobuf:"varint,5,opt,name=bits,proto3" json:"bits,omitempty"`
	Nonce      uint32 `protobuf:"varint,6,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Height     int32  `protobuf:"varint,7,opt,name=height,proto3" json:"height,omitempty"`
	WorkSum    []byte `protobuf:"bytes,8,opt,name=workSum,proto3" json:"workSum,omitempty"`
	Hash       []byte `protobuf:"bytes,9,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *BlockHeaderProto) Reset() {
	*x = BlockHeaderProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockHeader_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockHeaderProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockHeaderProto) ProtoMessage() {}

func (x *BlockHeaderProto) ProtoReflect() protoreflect.Message {
	mi := &file_blockHeader_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockHeaderProto.ProtoReflect.Descriptor instead.
func (*BlockHeaderProto) Descriptor() ([]byte, []int) {
	return file_blockHeader_proto_rawDescGZIP(), []int{0}
}

func (x *BlockHeaderProto) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *BlockHeaderProto) GetPrevBlock() []byte {
	if x != nil {
		return x.PrevBlock
	}
	return nil
}

func (x *BlockHeaderProto) GetMerkleRoot() []byte {
	if x != nil {
		return x.MerkleRoot
	}
	return nil
}

func (x *BlockHeaderProto) GetTimestamp() uint32 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *BlockHeaderProto) GetBits() uint32 {
	if x != nil {
		return x.Bits
	}
	return 0
}

func (x *BlockHeaderProto) GetNonce() uint32 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

func (x *BlockHeaderProto) GetHeight() int32 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *BlockHeaderProto) GetWorkSum() []byte {
	if x != nil {
		return x.WorkSum
	}
	return nil
}

func (x *BlockHeaderProto) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

var File_blockHeader_proto protoreflect.FileDescriptor

var file_blockHeader_proto_rawDesc = []byte{
	0x0a, 0x11, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x22, 0xf8, 0x01, 0x0a, 0x10, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72, 0x65, 0x76, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x72, 0x65, 0x76, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x1e,
	0x0a, 0x0a, 0x6d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x0a, 0x6d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x1c,
	0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x12, 0x0a, 0x04,
	0x62, 0x69, 0x74, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x62, 0x69, 0x74, 0x73,
	0x12, 0x14, 0x0a, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x18,
	0x0a, 0x07, 0x77, 0x6f, 0x72, 0x6b, 0x53, 0x75, 0x6d, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x07, 0x77, 0x6f, 0x72, 0x6b, 0x53, 0x75, 0x6d, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x42, 0x10, 0x5a, 0x0e,
	0x2e, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_blockHeader_proto_rawDescOnce sync.Once
	file_blockHeader_proto_rawDescData = file_blockHeader_proto_rawDesc
)

func file_blockHeader_proto_rawDescGZIP() []byte {
	file_blockHeader_proto_rawDescOnce.Do(func() {
		file_blockHeader_proto_rawDescData = protoimpl.X.CompressGZIP(file_blockHeader_proto_rawDescData)
	})
	return file_blockHeader_proto_rawDescData
}

var file_blockHeader_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_blockHeader_proto_goTypes = []interface{}{
	(*BlockHeaderProto)(nil), // 0: znn_storage.BlockHeaderProto
}
var file_blockHeader_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_blockHeader_proto_init() }
func file_blockHeader_proto_init() {
	if File_blockHeader_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_blockHeader_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockHeaderProto); i {
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
			RawDescriptor: file_blockHeader_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_blockHeader_proto_goTypes,
		DependencyIndexes: file_blockHeader_proto_depIdxs,
		MessageInfos:      file_blockHeader_proto_msgTypes,
	}.Build()
	File_blockHeader_proto = out.File
	file_blockHeader_proto_rawDesc = nil
	file_blockHeader_proto_goTypes = nil
	file_blockHeader_proto_depIdxs = nil
}
