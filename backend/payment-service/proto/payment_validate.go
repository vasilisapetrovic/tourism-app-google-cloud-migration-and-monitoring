package proto

type ValidateTokenRequest struct {
	TouristId int64  `protobuf:"varint,1,opt,name=touristId,proto3" json:"touristId,omitempty"`
	TourId    string `protobuf:"bytes,2,opt,name=tourId,proto3" json:"tourId,omitempty"`
}

func (x *ValidateTokenRequest) Reset()         {}
func (x *ValidateTokenRequest) String() string { return x.TourId }
func (x *ValidateTokenRequest) ProtoMessage()  {}

func (x *ValidateTokenRequest) GetTouristId() int64 {
	if x != nil {
		return x.TouristId
	}
	return 0
}
func (x *ValidateTokenRequest) GetTourId() string {
	if x != nil {
		return x.TourId
	}
	return ""
}

type ValidateTokenResponse struct {
	Valid   bool   `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
	TokenId string `protobuf:"bytes,2,opt,name=tokenId,proto3" json:"tokenId,omitempty"`
	Error   string `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ValidateTokenResponse) Reset()         {}
func (x *ValidateTokenResponse) String() string { return "" }
func (x *ValidateTokenResponse) ProtoMessage()  {}

func (x *ValidateTokenResponse) GetValid() bool {
	if x != nil {
		return x.Valid
	}
	return false
}
func (x *ValidateTokenResponse) GetTokenId() string {
	if x != nil {
		return x.TokenId
	}
	return ""
}
func (x *ValidateTokenResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}