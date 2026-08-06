package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	TourService_GetTourById_FullMethodName    = "/TourService/GetTourById"
	TourService_PublishTour_FullMethodName    = "/TourService/PublishTour"
	TourService_StartExecution_FullMethodName = "/TourService/StartExecution"
	TourService_GetExecution_FullMethodName   = "/TourService/GetExecution"
)

type TourServiceClient interface {
	GetTourById(ctx context.Context, in *GetTourByIdRequest, opts ...grpc.CallOption) (*GetTourByIdResponse, error)
	PublishTour(ctx context.Context, in *PublishTourRequest, opts ...grpc.CallOption) (*PublishTourResponse, error)
	StartExecution(ctx context.Context, in *StartExecutionRequest, opts ...grpc.CallOption) (*StartExecutionResponse, error)
	GetExecution(ctx context.Context, in *GetExecutionRequest, opts ...grpc.CallOption) (*GetExecutionResponse, error)
}

type tourServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTourServiceClient(cc grpc.ClientConnInterface) TourServiceClient {
	return &tourServiceClient{cc}
}

func (c *tourServiceClient) GetTourById(ctx context.Context, in *GetTourByIdRequest, opts ...grpc.CallOption) (*GetTourByIdResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetTourByIdResponse)
	err := c.cc.Invoke(ctx, TourService_GetTourById_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tourServiceClient) PublishTour(ctx context.Context, in *PublishTourRequest, opts ...grpc.CallOption) (*PublishTourResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PublishTourResponse)
	err := c.cc.Invoke(ctx, TourService_PublishTour_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tourServiceClient) StartExecution(ctx context.Context, in *StartExecutionRequest, opts ...grpc.CallOption) (*StartExecutionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StartExecutionResponse)
	err := c.cc.Invoke(ctx, TourService_StartExecution_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tourServiceClient) GetExecution(ctx context.Context, in *GetExecutionRequest, opts ...grpc.CallOption) (*GetExecutionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetExecutionResponse)
	err := c.cc.Invoke(ctx, TourService_GetExecution_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type TourServiceServer interface {
	GetTourById(context.Context, *GetTourByIdRequest) (*GetTourByIdResponse, error)
	PublishTour(context.Context, *PublishTourRequest) (*PublishTourResponse, error)
	StartExecution(context.Context, *StartExecutionRequest) (*StartExecutionResponse, error)
	GetExecution(context.Context, *GetExecutionRequest) (*GetExecutionResponse, error)
	mustEmbedUnimplementedTourServiceServer()
}

type UnimplementedTourServiceServer struct{}

func (UnimplementedTourServiceServer) GetTourById(context.Context, *GetTourByIdRequest) (*GetTourByIdResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetTourById not implemented")
}
func (UnimplementedTourServiceServer) PublishTour(context.Context, *PublishTourRequest) (*PublishTourResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method PublishTour not implemented")
}
func (UnimplementedTourServiceServer) StartExecution(context.Context, *StartExecutionRequest) (*StartExecutionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method StartExecution not implemented")
}
func (UnimplementedTourServiceServer) GetExecution(context.Context, *GetExecutionRequest) (*GetExecutionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetExecution not implemented")
}
func (UnimplementedTourServiceServer) mustEmbedUnimplementedTourServiceServer() {}
func (UnimplementedTourServiceServer) testEmbeddedByValue()                     {}

type UnsafeTourServiceServer interface {
	mustEmbedUnimplementedTourServiceServer()
}

func RegisterTourServiceServer(s grpc.ServiceRegistrar, srv TourServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TourService_ServiceDesc, srv)
}

func _TourService_GetTourById_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTourByIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TourServiceServer).GetTourById(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: TourService_GetTourById_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TourServiceServer).GetTourById(ctx, req.(*GetTourByIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TourService_PublishTour_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishTourRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TourServiceServer).PublishTour(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: TourService_PublishTour_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TourServiceServer).PublishTour(ctx, req.(*PublishTourRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TourService_StartExecution_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartExecutionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TourServiceServer).StartExecution(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: TourService_StartExecution_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TourServiceServer).StartExecution(ctx, req.(*StartExecutionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TourService_GetExecution_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetExecutionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TourServiceServer).GetExecution(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: TourService_GetExecution_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TourServiceServer).GetExecution(ctx, req.(*GetExecutionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var TourService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "TourService",
	HandlerType: (*TourServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GetTourById", Handler: _TourService_GetTourById_Handler},
		{MethodName: "PublishTour", Handler: _TourService_PublishTour_Handler},
		{MethodName: "StartExecution", Handler: _TourService_StartExecution_Handler},
		{MethodName: "GetExecution", Handler: _TourService_GetExecution_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tour-service.proto",
}