package proto

import (
	context "context"

	grpc "google.golang.org/grpc"
)

type StartExecutionRequest struct {
	TourId          string  `protobuf:"bytes,1,opt,name=tourId,proto3" json:"tourId,omitempty"`
	TouristId       int64   `protobuf:"varint,2,opt,name=touristId,proto3" json:"touristId,omitempty"`
	TouristUsername string  `protobuf:"bytes,3,opt,name=touristUsername,proto3" json:"touristUsername,omitempty"`
	StartLat        float64 `protobuf:"fixed64,4,opt,name=startLat,proto3" json:"startLat,omitempty"`
	StartLng        float64 `protobuf:"fixed64,5,opt,name=startLng,proto3" json:"startLng,omitempty"`
}

func (x *StartExecutionRequest) Reset()         {}
func (x *StartExecutionRequest) String() string { return x.TourId }
func (x *StartExecutionRequest) ProtoMessage()  {}

func (x *StartExecutionRequest) GetTourId() string {
	if x != nil {
		return x.TourId
	}
	return ""
}
func (x *StartExecutionRequest) GetTouristId() int64 {
	if x != nil {
		return x.TouristId
	}
	return 0
}
func (x *StartExecutionRequest) GetTouristUsername() string {
	if x != nil {
		return x.TouristUsername
	}
	return ""
}
func (x *StartExecutionRequest) GetStartLat() float64 {
	if x != nil {
		return x.StartLat
	}
	return 0
}
func (x *StartExecutionRequest) GetStartLng() float64 {
	if x != nil {
		return x.StartLng
	}
	return 0
}

type StartExecutionResponse struct {
	Execution *ExecutionProto `protobuf:"bytes,1,opt,name=execution,proto3" json:"execution,omitempty"`
	Error     string          `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *StartExecutionResponse) Reset()         {}
func (x *StartExecutionResponse) String() string { return "" }
func (x *StartExecutionResponse) ProtoMessage()  {}

func (x *StartExecutionResponse) GetExecution() *ExecutionProto {
	if x != nil {
		return x.Execution
	}
	return nil
}
func (x *StartExecutionResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type GetExecutionRequest struct {
	ExecId    string `protobuf:"bytes,1,opt,name=execId,proto3" json:"execId,omitempty"`
	TouristId int64  `protobuf:"varint,2,opt,name=touristId,proto3" json:"touristId,omitempty"`
}

func (x *GetExecutionRequest) Reset()         {}
func (x *GetExecutionRequest) String() string { return x.ExecId }
func (x *GetExecutionRequest) ProtoMessage()  {}

func (x *GetExecutionRequest) GetExecId() string {
	if x != nil {
		return x.ExecId
	}
	return ""
}
func (x *GetExecutionRequest) GetTouristId() int64 {
	if x != nil {
		return x.TouristId
	}
	return 0
}

type GetExecutionResponse struct {
	Execution *ExecutionProto `protobuf:"bytes,1,opt,name=execution,proto3" json:"execution,omitempty"`
	Error     string          `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *GetExecutionResponse) Reset()         {}
func (x *GetExecutionResponse) String() string { return "" }
func (x *GetExecutionResponse) ProtoMessage()  {}

func (x *GetExecutionResponse) GetExecution() *ExecutionProto {
	if x != nil {
		return x.Execution
	}
	return nil
}
func (x *GetExecutionResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type CompletedKeypointProto struct {
	KeypointId  string `protobuf:"bytes,1,opt,name=keypointId,proto3" json:"keypointId,omitempty"`
	Name        string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	CompletedAt string `protobuf:"bytes,3,opt,name=completedAt,proto3" json:"completedAt,omitempty"`
}

func (x *CompletedKeypointProto) Reset()         {}
func (x *CompletedKeypointProto) String() string { return x.Name }
func (x *CompletedKeypointProto) ProtoMessage()  {}

func (x *CompletedKeypointProto) GetKeypointId() string {
	if x != nil {
		return x.KeypointId
	}
	return ""
}
func (x *CompletedKeypointProto) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}
func (x *CompletedKeypointProto) GetCompletedAt() string {
	if x != nil {
		return x.CompletedAt
	}
	return ""
}

type ExecutionProto struct {
	Id                 string                    `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	TourId             string                    `protobuf:"bytes,2,opt,name=tourId,proto3" json:"tourId,omitempty"`
	TouristId          int64                     `protobuf:"varint,3,opt,name=touristId,proto3" json:"touristId,omitempty"`
	TouristUsername    string                    `protobuf:"bytes,4,opt,name=touristUsername,proto3" json:"touristUsername,omitempty"`
	Status             string                    `protobuf:"bytes,5,opt,name=status,proto3" json:"status,omitempty"`
	StartedAt          string                    `protobuf:"bytes,6,opt,name=startedAt,proto3" json:"startedAt,omitempty"`
	EndedAt            string                    `protobuf:"bytes,7,opt,name=endedAt,proto3" json:"endedAt,omitempty"`
	LastActivity       string                    `protobuf:"bytes,8,opt,name=lastActivity,proto3" json:"lastActivity,omitempty"`
	CurrentLat         float64                   `protobuf:"fixed64,9,opt,name=currentLat,proto3" json:"currentLat,omitempty"`
	CurrentLng         float64                   `protobuf:"fixed64,10,opt,name=currentLng,proto3" json:"currentLng,omitempty"`
	CompletedKeypoints []*CompletedKeypointProto `protobuf:"bytes,11,rep,name=completedKeypoints,proto3" json:"completedKeypoints,omitempty"`
}

func (x *ExecutionProto) Reset()         {}
func (x *ExecutionProto) String() string { return x.Id }
func (x *ExecutionProto) ProtoMessage()  {}

func (x *ExecutionProto) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}
func (x *ExecutionProto) GetTourId() string {
	if x != nil {
		return x.TourId
	}
	return ""
}
func (x *ExecutionProto) GetTouristId() int64 {
	if x != nil {
		return x.TouristId
	}
	return 0
}
func (x *ExecutionProto) GetTouristUsername() string {
	if x != nil {
		return x.TouristUsername
	}
	return ""
}
func (x *ExecutionProto) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}
func (x *ExecutionProto) GetStartedAt() string {
	if x != nil {
		return x.StartedAt
	}
	return ""
}
func (x *ExecutionProto) GetEndedAt() string {
	if x != nil {
		return x.EndedAt
	}
	return ""
}
func (x *ExecutionProto) GetLastActivity() string {
	if x != nil {
		return x.LastActivity
	}
	return ""
}
func (x *ExecutionProto) GetCurrentLat() float64 {
	if x != nil {
		return x.CurrentLat
	}
	return 0
}
func (x *ExecutionProto) GetCurrentLng() float64 {
	if x != nil {
		return x.CurrentLng
	}
	return 0
}
func (x *ExecutionProto) GetCompletedKeypoints() []*CompletedKeypointProto {
	if x != nil {
		return x.CompletedKeypoints
	}
	return nil
}

// Dodaj StartExecution i GetExecution u TourServiceClient
type TourServiceExecutionClient interface {
	StartExecution(ctx context.Context, in *StartExecutionRequest, opts ...grpc.CallOption) (*StartExecutionResponse, error)
	GetExecution(ctx context.Context, in *GetExecutionRequest, opts ...grpc.CallOption) (*GetExecutionResponse, error)
}

type tourServiceExecutionClient struct {
	cc grpc.ClientConnInterface
}

func NewTourServiceExecutionClient(cc grpc.ClientConnInterface) TourServiceExecutionClient {
	return &tourServiceExecutionClient{cc}
}

func (c *tourServiceExecutionClient) StartExecution(ctx context.Context, in *StartExecutionRequest, opts ...grpc.CallOption) (*StartExecutionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StartExecutionResponse)
	err := c.cc.Invoke(ctx, TourService_StartExecution_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tourServiceExecutionClient) GetExecution(ctx context.Context, in *GetExecutionRequest, opts ...grpc.CallOption) (*GetExecutionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetExecutionResponse)
	err := c.cc.Invoke(ctx, TourService_GetExecution_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}