package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
)

const PaymentService_ValidateToken_FullMethodName = "/PaymentService/ValidateToken"

type PaymentValidateTokenRequest struct {
	TouristId int64  `protobuf:"varint,1,opt,name=touristId,proto3" json:"touristId,omitempty"`
	TourId    string `protobuf:"bytes,2,opt,name=tourId,proto3" json:"tourId,omitempty"`
}

func (x *PaymentValidateTokenRequest) Reset()         {}
func (x *PaymentValidateTokenRequest) String() string { return x.TourId }
func (x *PaymentValidateTokenRequest) ProtoMessage()  {}

func (x *PaymentValidateTokenRequest) GetTouristId() int64 {
	if x != nil {
		return x.TouristId
	}
	return 0
}
func (x *PaymentValidateTokenRequest) GetTourId() string {
	if x != nil {
		return x.TourId
	}
	return ""
}

type PaymentValidateTokenResponse struct {
	Valid   bool   `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
	TokenId string `protobuf:"bytes,2,opt,name=tokenId,proto3" json:"tokenId,omitempty"`
	Error   string `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *PaymentValidateTokenResponse) Reset()         {}
func (x *PaymentValidateTokenResponse) String() string { return "" }
func (x *PaymentValidateTokenResponse) ProtoMessage()  {}

func (x *PaymentValidateTokenResponse) GetValid() bool {
	if x != nil {
		return x.Valid
	}
	return false
}
func (x *PaymentValidateTokenResponse) GetTokenId() string {
	if x != nil {
		return x.TokenId
	}
	return ""
}
func (x *PaymentValidateTokenResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type PaymentServiceGrpcClient interface {
	ValidateToken(ctx context.Context, in *PaymentValidateTokenRequest, opts ...grpc.CallOption) (*PaymentValidateTokenResponse, error)
}

type paymentServiceGrpcClient struct {
	cc grpc.ClientConnInterface
}

func NewPaymentServiceGrpcClient(cc grpc.ClientConnInterface) PaymentServiceGrpcClient {
	return &paymentServiceGrpcClient{cc}
}

func (c *paymentServiceGrpcClient) ValidateToken(ctx context.Context, in *PaymentValidateTokenRequest, opts ...grpc.CallOption) (*PaymentValidateTokenResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PaymentValidateTokenResponse)
	err := c.cc.Invoke(ctx, PaymentService_ValidateToken_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}