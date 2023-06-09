// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.0
// source: proto/client.proto

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

type Data struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Data:
	//
	//	*Data_PasswordData
	//	*Data_TextData
	//	*Data_BinaryData
	//	*Data_CardData
	Data isData_Data `protobuf_oneof:"data"`
}

func (x *Data) Reset() {
	*x = Data{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_client_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Data) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Data) ProtoMessage() {}

func (x *Data) ProtoReflect() protoreflect.Message {
	mi := &file_proto_client_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Data.ProtoReflect.Descriptor instead.
func (*Data) Descriptor() ([]byte, []int) {
	return file_proto_client_proto_rawDescGZIP(), []int{0}
}

func (m *Data) GetData() isData_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *Data) GetPasswordData() *Data_Password {
	if x, ok := x.GetData().(*Data_PasswordData); ok {
		return x.PasswordData
	}
	return nil
}

func (x *Data) GetTextData() *Data_Text {
	if x, ok := x.GetData().(*Data_TextData); ok {
		return x.TextData
	}
	return nil
}

func (x *Data) GetBinaryData() *Data_Binary {
	if x, ok := x.GetData().(*Data_BinaryData); ok {
		return x.BinaryData
	}
	return nil
}

func (x *Data) GetCardData() *Data_Card {
	if x, ok := x.GetData().(*Data_CardData); ok {
		return x.CardData
	}
	return nil
}

type isData_Data interface {
	isData_Data()
}

type Data_PasswordData struct {
	PasswordData *Data_Password `protobuf:"bytes,1,opt,name=password_data,json=passwordData,proto3,oneof"`
}

type Data_TextData struct {
	TextData *Data_Text `protobuf:"bytes,2,opt,name=text_data,json=textData,proto3,oneof"`
}

type Data_BinaryData struct {
	BinaryData *Data_Binary `protobuf:"bytes,3,opt,name=binary_data,json=binaryData,proto3,oneof"`
}

type Data_CardData struct {
	CardData *Data_Card `protobuf:"bytes,4,opt,name=card_data,json=cardData,proto3,oneof"`
}

func (*Data_PasswordData) isData_Data() {}

func (*Data_TextData) isData_Data() {}

func (*Data_BinaryData) isData_Data() {}

func (*Data_CardData) isData_Data() {}

type Data_Password struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Login    string `protobuf:"bytes,1,opt,name=login,proto3" json:"login,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
}

func (x *Data_Password) Reset() {
	*x = Data_Password{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_client_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Data_Password) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Data_Password) ProtoMessage() {}

func (x *Data_Password) ProtoReflect() protoreflect.Message {
	mi := &file_proto_client_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Data_Password.ProtoReflect.Descriptor instead.
func (*Data_Password) Descriptor() ([]byte, []int) {
	return file_proto_client_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Data_Password) GetLogin() string {
	if x != nil {
		return x.Login
	}
	return ""
}

func (x *Data_Password) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

type Data_Text struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Text string `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
}

func (x *Data_Text) Reset() {
	*x = Data_Text{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_client_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Data_Text) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Data_Text) ProtoMessage() {}

func (x *Data_Text) ProtoReflect() protoreflect.Message {
	mi := &file_proto_client_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Data_Text.ProtoReflect.Descriptor instead.
func (*Data_Text) Descriptor() ([]byte, []int) {
	return file_proto_client_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Data_Text) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

type Data_Binary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Data_Binary) Reset() {
	*x = Data_Binary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_client_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Data_Binary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Data_Binary) ProtoMessage() {}

func (x *Data_Binary) ProtoReflect() protoreflect.Message {
	mi := &file_proto_client_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Data_Binary.ProtoReflect.Descriptor instead.
func (*Data_Binary) Descriptor() ([]byte, []int) {
	return file_proto_client_proto_rawDescGZIP(), []int{0, 2}
}

func (x *Data_Binary) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type Data_Card struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Number string `protobuf:"bytes,1,opt,name=number,proto3" json:"number,omitempty"`
	Expiry string `protobuf:"bytes,2,opt,name=expiry,proto3" json:"expiry,omitempty"`
	Cvc    string `protobuf:"bytes,3,opt,name=cvc,proto3" json:"cvc,omitempty"`
}

func (x *Data_Card) Reset() {
	*x = Data_Card{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_client_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Data_Card) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Data_Card) ProtoMessage() {}

func (x *Data_Card) ProtoReflect() protoreflect.Message {
	mi := &file_proto_client_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Data_Card.ProtoReflect.Descriptor instead.
func (*Data_Card) Descriptor() ([]byte, []int) {
	return file_proto_client_proto_rawDescGZIP(), []int{0, 3}
}

func (x *Data_Card) GetNumber() string {
	if x != nil {
		return x.Number
	}
	return ""
}

func (x *Data_Card) GetExpiry() string {
	if x != nil {
		return x.Expiry
	}
	return ""
}

func (x *Data_Card) GetCvc() string {
	if x != nil {
		return x.Cvc
	}
	return ""
}

var File_proto_client_proto protoreflect.FileDescriptor

var file_proto_client_proto_rawDesc = []byte{
	0x0a, 0x12, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x22, 0xaa, 0x03, 0x0a,
	0x04, 0x44, 0x61, 0x74, 0x61, 0x12, 0x3c, 0x0a, 0x0d, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72,
	0x64, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x50, 0x61, 0x73, 0x73, 0x77,
	0x6f, 0x72, 0x64, 0x48, 0x00, 0x52, 0x0c, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x30, 0x0a, 0x09, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e,
	0x44, 0x61, 0x74, 0x61, 0x2e, 0x54, 0x65, 0x78, 0x74, 0x48, 0x00, 0x52, 0x08, 0x74, 0x65, 0x78,
	0x74, 0x44, 0x61, 0x74, 0x61, 0x12, 0x36, 0x0a, 0x0b, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x5f,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x63, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x48,
	0x00, 0x52, 0x0a, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x44, 0x61, 0x74, 0x61, 0x12, 0x30, 0x0a,
	0x09, 0x63, 0x61, 0x72, 0x64, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x11, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x43,
	0x61, 0x72, 0x64, 0x48, 0x00, 0x52, 0x08, 0x63, 0x61, 0x72, 0x64, 0x44, 0x61, 0x74, 0x61, 0x1a,
	0x3c, 0x0a, 0x08, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x6c,
	0x6f, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x6f, 0x67, 0x69,
	0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x1a, 0x1a, 0x0a,
	0x04, 0x54, 0x65, 0x78, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x1a, 0x1c, 0x0a, 0x06, 0x42, 0x69, 0x6e,
	0x61, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x48, 0x0a, 0x04, 0x43, 0x61, 0x72, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x70, 0x69, 0x72,
	0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x65, 0x78, 0x70, 0x69, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x63, 0x76, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x63, 0x76,
	0x63, 0x42, 0x06, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42, 0x0e, 0x5a, 0x0c, 0x63, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_proto_client_proto_rawDescOnce sync.Once
	file_proto_client_proto_rawDescData = file_proto_client_proto_rawDesc
)

func file_proto_client_proto_rawDescGZIP() []byte {
	file_proto_client_proto_rawDescOnce.Do(func() {
		file_proto_client_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_client_proto_rawDescData)
	})
	return file_proto_client_proto_rawDescData
}

var file_proto_client_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_client_proto_goTypes = []interface{}{
	(*Data)(nil),          // 0: client.Data
	(*Data_Password)(nil), // 1: client.Data.Password
	(*Data_Text)(nil),     // 2: client.Data.Text
	(*Data_Binary)(nil),   // 3: client.Data.Binary
	(*Data_Card)(nil),     // 4: client.Data.Card
}
var file_proto_client_proto_depIdxs = []int32{
	1, // 0: client.Data.password_data:type_name -> client.Data.Password
	2, // 1: client.Data.text_data:type_name -> client.Data.Text
	3, // 2: client.Data.binary_data:type_name -> client.Data.Binary
	4, // 3: client.Data.card_data:type_name -> client.Data.Card
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_client_proto_init() }
func file_proto_client_proto_init() {
	if File_proto_client_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_client_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Data); i {
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
		file_proto_client_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Data_Password); i {
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
		file_proto_client_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Data_Text); i {
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
		file_proto_client_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Data_Binary); i {
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
		file_proto_client_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Data_Card); i {
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
	file_proto_client_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Data_PasswordData)(nil),
		(*Data_TextData)(nil),
		(*Data_BinaryData)(nil),
		(*Data_CardData)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_client_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_client_proto_goTypes,
		DependencyIndexes: file_proto_client_proto_depIdxs,
		MessageInfos:      file_proto_client_proto_msgTypes,
	}.Build()
	File_proto_client_proto = out.File
	file_proto_client_proto_rawDesc = nil
	file_proto_client_proto_goTypes = nil
	file_proto_client_proto_depIdxs = nil
}
