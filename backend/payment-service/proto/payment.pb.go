package proto

import (
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TourPurchaseToken struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	TouristId     int64                  `protobuf:"varint,2,opt,name=touristId,proto3" json:"touristId,omitempty"`
	TourId        string                 `protobuf:"bytes,3,opt,name=tourId,proto3" json:"tourId,omitempty"`
	TourName      string                 `protobuf:"bytes,4,opt,name=tourName,proto3" json:"tourName,omitempty"`
	Price         float64                `protobuf:"fixed64,5,opt,name=price,proto3" json:"price,omitempty"`
	PurchasedAt   string                 `protobuf:"bytes,6,opt,name=purchasedAt,proto3" json:"purchasedAt,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TourPurchaseToken) Reset() {
	*x = TourPurchaseToken{}
	mi := &file_proto_payment_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TourPurchaseToken) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TourPurchaseToken) ProtoMessage() {}

func (x *TourPurchaseToken) ProtoReflect() protoreflect.Message {
	mi := &file_proto_payment_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*TourPurchaseToken) Descriptor() ([]byte, []int) {
	return file_proto_payment_proto_rawDescGZIP(), []int{0}
}

func (x *TourPurchaseToken) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TourPurchaseToken) GetTouristId() int64 {
	if x != nil {
		return x.TouristId
	}
	return 0
}

func (x *TourPurchaseToken) GetTourId() string {
	if x != nil {
		return x.TourId
	}
	return ""
}

func (x *TourPurchaseToken) GetTourName() string {
	if x != nil {
		return x.TourName
	}
	return ""
}

func (x *TourPurchaseToken) GetPrice() float64 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *TourPurchaseToken) GetPurchasedAt() string {
	if x != nil {
		return x.PurchasedAt
	}
	return ""
}

type GetPurchasesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TouristId     int64                  `protobuf:"varint,1,opt,name=touristId,proto3" json:"touristId,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetPurchasesRequest) Reset() {
	*x = GetPurchasesRequest{}
	mi := &file_proto_payment_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetPurchasesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPurchasesRequest) ProtoMessage() {}

func (x *GetPurchasesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_payment_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetPurchasesRequest) Descriptor() ([]byte, []int) {
	return file_proto_payment_proto_rawDescGZIP(), []int{1}
}

func (x *GetPurchasesRequest) GetTouristId() int64 {
	if x != nil {
		return x.TouristId
	}
	return 0
}

type GetPurchasesResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tokens        []*TourPurchaseToken   `protobuf:"bytes,1,rep,name=tokens,proto3" json:"tokens,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetPurchasesResponse) Reset() {
	*x = GetPurchasesResponse{}
	mi := &file_proto_payment_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetPurchasesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPurchasesResponse) ProtoMessage() {}

func (x *GetPurchasesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_payment_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetPurchasesResponse) Descriptor() ([]byte, []int) {
	return file_proto_payment_proto_rawDescGZIP(), []int{2}
}

func (x *GetPurchasesResponse) GetTokens() []*TourPurchaseToken {
	if x != nil {
		return x.Tokens
	}
	return nil
}

type CheckoutRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TouristId     int64                  `protobuf:"varint,1,opt,name=touristId,proto3" json:"touristId,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CheckoutRequest) Reset() {
	*x = CheckoutRequest{}
	mi := &file_proto_payment_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CheckoutRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckoutRequest) ProtoMessage() {}

func (x *CheckoutRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_payment_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CheckoutRequest) Descriptor() ([]byte, []int) {
	return file_proto_payment_proto_rawDescGZIP(), []int{3}
}

func (x *CheckoutRequest) GetTouristId() int64 {
	if x != nil {
		return x.TouristId
	}
	return 0
}

type CheckoutResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tokens        []*TourPurchaseToken   `protobuf:"bytes,1,rep,name=tokens,proto3" json:"tokens,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CheckoutResponse) Reset() {
	*x = CheckoutResponse{}
	mi := &file_proto_payment_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CheckoutResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckoutResponse) ProtoMessage() {}

func (x *CheckoutResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_payment_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CheckoutResponse) Descriptor() ([]byte, []int) {
	return file_proto_payment_proto_rawDescGZIP(), []int{4}
}

func (x *CheckoutResponse) GetTokens() []*TourPurchaseToken {
	if x != nil {
		return x.Tokens
	}
	return nil
}

func (x *CheckoutResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_proto_payment_proto protoreflect.FileDescriptor

const file_proto_payment_proto_rawDesc = "" +
	"\n" +
	"\x13proto/payment.proto\"\xad\x01\n" +
	"\x11TourPurchaseToken\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x1c\n" +
	"\ttouristId\x18\x02 \x01(\x03R\ttouristId\x12\x16\n" +
	"\x06tourId\x18\x03 \x01(\tR\x06tourId\x12\x1a\n" +
	"\btourName\x18\x04 \x01(\tR\btourName\x12\x14\n" +
	"\x05price\x18\x05 \x01(\x01R\x05price\x12 \n" +
	"\vpurchasedAt\x18\x06 \x01(\tR\vpurchasedAt\"3\n" +
	"\x13GetPurchasesRequest\x12\x1c\n" +
	"\ttouristId\x18\x01 \x01(\x03R\ttouristId\"B\n" +
	"\x14GetPurchasesResponse\x12*\n" +
	"\x06tokens\x18\x01 \x03(\v2\x12.TourPurchaseTokenR\x06tokens\"/\n" +
	"\x0fCheckoutRequest\x12\x1c\n" +
	"\ttouristId\x18\x01 \x01(\x03R\ttouristId\"T\n" +
	"\x10CheckoutResponse\x12*\n" +
	"\x06tokens\x18\x01 \x03(\v2\x12.TourPurchaseTokenR\x06tokens\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error2\x82\x01\n" +
	"\x0ePaymentService\x12=\n" +
	"\fGetPurchases\x12\x14.GetPurchasesRequest\x1a\x15.GetPurchasesResponse\"\x00\x121\n" +
	"\bCheckout\x12\x10.CheckoutRequest\x1a\x11.CheckoutResponse\"\x00B\x17Z\x15payment-service/protob\x06proto3"

var (
	file_proto_payment_proto_rawDescOnce sync.Once
	file_proto_payment_proto_rawDescData []byte
)

func file_proto_payment_proto_rawDescGZIP() []byte {
	file_proto_payment_proto_rawDescOnce.Do(func() {
		file_proto_payment_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_payment_proto_rawDesc), len(file_proto_payment_proto_rawDesc)))
	})
	return file_proto_payment_proto_rawDescData
}

var file_proto_payment_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_payment_proto_goTypes = []any{
	(*TourPurchaseToken)(nil),    // 0: TourPurchaseToken
	(*GetPurchasesRequest)(nil),  // 1: GetPurchasesRequest
	(*GetPurchasesResponse)(nil), // 2: GetPurchasesResponse
	(*CheckoutRequest)(nil),      // 3: CheckoutRequest
	(*CheckoutResponse)(nil),     // 4: CheckoutResponse
}
var file_proto_payment_proto_depIdxs = []int32{
	0, // 0: GetPurchasesResponse.tokens:type_name -> TourPurchaseToken
	0, // 1: CheckoutResponse.tokens:type_name -> TourPurchaseToken
	1, // 2: PaymentService.GetPurchases:input_type -> GetPurchasesRequest
	3, // 3: PaymentService.Checkout:input_type -> CheckoutRequest
	2, // 4: PaymentService.GetPurchases:output_type -> GetPurchasesResponse
	4, // 5: PaymentService.Checkout:output_type -> CheckoutResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_payment_proto_init() }
func file_proto_payment_proto_init() {
	if File_proto_payment_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_payment_proto_rawDesc), len(file_proto_payment_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_payment_proto_goTypes,
		DependencyIndexes: file_proto_payment_proto_depIdxs,
		MessageInfos:      file_proto_payment_proto_msgTypes,
	}.Build()
	File_proto_payment_proto = out.File
	file_proto_payment_proto_goTypes = nil
	file_proto_payment_proto_depIdxs = nil
}
