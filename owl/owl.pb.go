// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: owl/owl.proto

package owl

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type EntityEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Iri    string   `protobuf:"bytes,1,opt,name=iri,proto3" json:"iri,omitempty"`
	Parent []string `protobuf:"bytes,2,rep,name=parent,proto3" json:"parent,omitempty"`
}

func (x *EntityEntry) Reset() {
	*x = EntityEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_owl_owl_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EntityEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EntityEntry) ProtoMessage() {}

func (x *EntityEntry) ProtoReflect() protoreflect.Message {
	mi := &file_owl_owl_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EntityEntry.ProtoReflect.Descriptor instead.
func (*EntityEntry) Descriptor() ([]byte, []int) {
	return file_owl_owl_proto_rawDescGZIP(), []int{0}
}

func (x *EntityEntry) GetIri() string {
	if x != nil {
		return x.Iri
	}
	return ""
}

func (x *EntityEntry) GetParent() []string {
	if x != nil {
		return x.Parent
	}
	return nil
}

type PrefixEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Iri    string `protobuf:"bytes,2,opt,name=iri,proto3" json:"iri,omitempty"`
}

func (x *PrefixEntry) Reset() {
	*x = PrefixEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_owl_owl_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrefixEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrefixEntry) ProtoMessage() {}

func (x *PrefixEntry) ProtoReflect() protoreflect.Message {
	mi := &file_owl_owl_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrefixEntry.ProtoReflect.Descriptor instead.
func (*PrefixEntry) Descriptor() ([]byte, []int) {
	return file_owl_owl_proto_rawDescGZIP(), []int{1}
}

func (x *PrefixEntry) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *PrefixEntry) GetIri() string {
	if x != nil {
		return x.Iri
	}
	return ""
}

type Meta struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefixes []*PrefixEntry `protobuf:"bytes,1,rep,name=prefixes,proto3" json:"prefixes,omitempty"`
}

func (x *Meta) Reset() {
	*x = Meta{}
	if protoimpl.UnsafeEnabled {
		mi := &file_owl_owl_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Meta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Meta) ProtoMessage() {}

func (x *Meta) ProtoReflect() protoreflect.Message {
	mi := &file_owl_owl_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Meta.ProtoReflect.Descriptor instead.
func (*Meta) Descriptor() ([]byte, []int) {
	return file_owl_owl_proto_rawDescGZIP(), []int{2}
}

func (x *Meta) GetPrefixes() []*PrefixEntry {
	if x != nil {
		return x.Prefixes
	}
	return nil
}

var file_owl_owl_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*EntityEntry)(nil),
		Field:         50000,
		Name:          "owl.class",
		Tag:           "bytes,50000,opt,name=class",
		Filename:      "owl/owl.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*EntityEntry)(nil),
		Field:         50000,
		Name:          "owl.property",
		Tag:           "bytes,50000,opt,name=property",
		Filename:      "owl/owl.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FileOptions)(nil),
		ExtensionType: (*Meta)(nil),
		Field:         50000,
		Name:          "owl.meta",
		Tag:           "bytes,50000,opt,name=meta",
		Filename:      "owl/owl.proto",
	},
}

// Extension fields to descriptorpb.MessageOptions.
var (
	// optional owl.EntityEntry class = 50000;
	E_Class = &file_owl_owl_proto_extTypes[0]
)

// Extension fields to descriptorpb.FieldOptions.
var (
	// optional owl.EntityEntry property = 50000;
	E_Property = &file_owl_owl_proto_extTypes[1]
)

// Extension fields to descriptorpb.FileOptions.
var (
	// optional owl.Meta meta = 50000;
	E_Meta = &file_owl_owl_proto_extTypes[2]
)

var File_owl_owl_proto protoreflect.FileDescriptor

var file_owl_owl_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6f, 0x77, 0x6c, 0x2f, 0x6f, 0x77, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x03, 0x6f, 0x77, 0x6c, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x37, 0x0a, 0x0b, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x72, 0x69, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x69, 0x72, 0x69, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e,
	0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x22,
	0x37, 0x0a, 0x0b, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x16,
	0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x72, 0x69, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x69, 0x72, 0x69, 0x22, 0x34, 0x0a, 0x04, 0x4d, 0x65, 0x74, 0x61,
	0x12, 0x2c, 0x0a, 0x08, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6f, 0x77, 0x6c, 0x2e, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x3a, 0x4c,
	0x0a, 0x05, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd0, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x6f, 0x77, 0x6c, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x05, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x88, 0x01, 0x01, 0x3a, 0x50, 0x0a, 0x08,
	0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd0, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x10, 0x2e, 0x6f, 0x77, 0x6c, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x88, 0x01, 0x01, 0x3a, 0x40,
	0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x12, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd0, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x6f,
	0x77, 0x6c, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x88, 0x01, 0x01,
	0x42, 0x21, 0x5a, 0x1f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f,
	0x78, 0x69, 0x73, 0x74, 0x6f, 0x2f, 0x6f, 0x77, 0x6c, 0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x6f, 0x77, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_owl_owl_proto_rawDescOnce sync.Once
	file_owl_owl_proto_rawDescData = file_owl_owl_proto_rawDesc
)

func file_owl_owl_proto_rawDescGZIP() []byte {
	file_owl_owl_proto_rawDescOnce.Do(func() {
		file_owl_owl_proto_rawDescData = protoimpl.X.CompressGZIP(file_owl_owl_proto_rawDescData)
	})
	return file_owl_owl_proto_rawDescData
}

var file_owl_owl_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_owl_owl_proto_goTypes = []any{
	(*EntityEntry)(nil),                 // 0: owl.EntityEntry
	(*PrefixEntry)(nil),                 // 1: owl.PrefixEntry
	(*Meta)(nil),                        // 2: owl.Meta
	(*descriptorpb.MessageOptions)(nil), // 3: google.protobuf.MessageOptions
	(*descriptorpb.FieldOptions)(nil),   // 4: google.protobuf.FieldOptions
	(*descriptorpb.FileOptions)(nil),    // 5: google.protobuf.FileOptions
}
var file_owl_owl_proto_depIdxs = []int32{
	1, // 0: owl.Meta.prefixes:type_name -> owl.PrefixEntry
	3, // 1: owl.class:extendee -> google.protobuf.MessageOptions
	4, // 2: owl.property:extendee -> google.protobuf.FieldOptions
	5, // 3: owl.meta:extendee -> google.protobuf.FileOptions
	0, // 4: owl.class:type_name -> owl.EntityEntry
	0, // 5: owl.property:type_name -> owl.EntityEntry
	2, // 6: owl.meta:type_name -> owl.Meta
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	4, // [4:7] is the sub-list for extension type_name
	1, // [1:4] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_owl_owl_proto_init() }
func file_owl_owl_proto_init() {
	if File_owl_owl_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_owl_owl_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*EntityEntry); i {
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
		file_owl_owl_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*PrefixEntry); i {
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
		file_owl_owl_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*Meta); i {
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
			RawDescriptor: file_owl_owl_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 3,
			NumServices:   0,
		},
		GoTypes:           file_owl_owl_proto_goTypes,
		DependencyIndexes: file_owl_owl_proto_depIdxs,
		MessageInfos:      file_owl_owl_proto_msgTypes,
		ExtensionInfos:    file_owl_owl_proto_extTypes,
	}.Build()
	File_owl_owl_proto = out.File
	file_owl_owl_proto_rawDesc = nil
	file_owl_owl_proto_goTypes = nil
	file_owl_owl_proto_depIdxs = nil
}
