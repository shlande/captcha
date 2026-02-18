package v1

import (
	"context"
	"google.golang.org/grpc"
)

type VerifyRequest struct {
	Provider string
	Token    string
	RemoteIp string
	Extra    map[string]string
}

type VerifyResponse struct {
	Success  bool
	Provider string
	Result   *ValidationResult
	Error    *Error
}

type ValidationResult struct {
	Success     bool
	ChallengeTs string
	Hostname    string
	Action      string
	CData       string
	ErrorCodes  []string
	Metadata    map[string]string
}

type Error struct {
	Code    string
	Message string
	Details []string
}

type ListProvidersRequest struct{}

type ListProvidersResponse struct {
	Providers []string
}

type CaptchaServiceClient interface {
	Verify(ctx context.Context, in *VerifyRequest, opts ...grpc.CallOption) (*VerifyResponse, error)
	ListProviders(ctx context.Context, in *ListProvidersRequest, opts ...grpc.CallOption) (*ListProvidersResponse, error)
}

type CaptchaServiceServer interface {
	Verify(context.Context, *VerifyRequest) (*VerifyResponse, error)
	ListProviders(context.Context, *ListProvidersRequest) (*ListProvidersResponse, error)
}

type UnimplementedCaptchaServiceServer struct{}

func (UnimplementedCaptchaServiceServer) Verify(context.Context, *VerifyRequest) (*VerifyResponse, error) {
	return nil, nil
}

func (UnimplementedCaptchaServiceServer) ListProviders(context.Context, *ListProvidersRequest) (*ListProvidersResponse, error) {
	return nil, nil
}

func RegisterCaptchaServiceServer(s grpc.ServiceRegistrar, srv CaptchaServiceServer) {
	s.RegisterService(&_CaptchaService_serviceDesc, srv)
}

var _CaptchaService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "captcha.v1.CaptchaService",
	HandlerType: (*CaptchaServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Verify",
			Handler:    _CaptchaService_Verify_Handler,
		},
		{
			MethodName: "ListProviders",
			Handler:    _CaptchaService_ListProviders_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/captcha/v1/captcha.proto",
}

func _CaptchaService_Verify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CaptchaServiceServer).Verify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/captcha.v1.CaptchaService/Verify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CaptchaServiceServer).Verify(ctx, req.(*VerifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CaptchaService_ListProviders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListProvidersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CaptchaServiceServer).ListProviders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/captcha.v1.CaptchaService/ListProviders",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CaptchaServiceServer).ListProviders(ctx, req.(*ListProvidersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

type captchaServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCaptchaServiceClient(cc grpc.ClientConnInterface) CaptchaServiceClient {
	return &captchaServiceClient{cc}
}

func (c *captchaServiceClient) Verify(ctx context.Context, in *VerifyRequest, opts ...grpc.CallOption) (*VerifyResponse, error) {
	out := new(VerifyResponse)
	err := c.cc.Invoke(ctx, "/captcha.v1.CaptchaService/Verify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *captchaServiceClient) ListProviders(ctx context.Context, in *ListProvidersRequest, opts ...grpc.CallOption) (*ListProvidersResponse, error) {
	out := new(ListProvidersResponse)
	err := c.cc.Invoke(ctx, "/captcha.v1.CaptchaService/ListProviders", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
