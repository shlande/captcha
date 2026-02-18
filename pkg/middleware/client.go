package middleware

import (
	"captcha/api/captcha/v1"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	Verify(ctx context.Context, provider, token, remoteIP string, extra map[string]string) (*v1.VerifyResponse, error)
	ListProviders(ctx context.Context) (*v1.ListProvidersResponse, error)
}

type grpcClient struct {
	client v1.CaptchaServiceClient
}

type ClientOption func(*grpcClient)

func WithInsecure() ClientOption {
	return func(c *grpcClient) {
	}
}

func NewClient(addr string, opts ...ClientOption) (Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := &grpcClient{
		client: v1.NewCaptchaServiceClient(conn),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

func (c *grpcClient) Verify(ctx context.Context, provider, token, remoteIP string, extra map[string]string) (*v1.VerifyResponse, error) {
	return c.client.Verify(ctx, &v1.VerifyRequest{
		Provider: provider,
		Token:    token,
		RemoteIp: remoteIP,
		Extra:    extra,
	})
}

func (c *grpcClient) ListProviders(ctx context.Context) (*v1.ListProvidersResponse, error) {
	return c.client.ListProviders(ctx, &v1.ListProvidersRequest{})
}
