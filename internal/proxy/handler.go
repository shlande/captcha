package proxy

import (
	"captcha/api/captcha/v1"
	"captcha/internal/validator"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	v1.UnimplementedCaptchaServiceServer
	validator *validator.Validator
}

func NewHandler(v *validator.Validator) *Handler {
	return &Handler{
		validator: v,
	}
}

func (h *Handler) Verify(ctx context.Context, req *v1.VerifyRequest) (*v1.VerifyResponse, error) {
	if req.Provider == "" {
		return &v1.VerifyResponse{
			Success:  false,
			Provider: req.Provider,
			Error: &v1.Error{
				Code:    "missing-provider",
				Message: "provider is required",
			},
		}, nil
	}

	if req.Token == "" {
		return &v1.VerifyResponse{
			Success:  false,
			Provider: req.Provider,
			Error: &v1.Error{
				Code:    "missing-input-response",
				Message: "token is required",
			},
		}, nil
	}

	result, err := h.validator.Validate(ctx, req.Provider, req.Token, req.RemoteIp, req.Extra)
	if err != nil {
		if err == validator.ErrProviderNotFound {
			return &v1.VerifyResponse{
				Success:  false,
				Provider: req.Provider,
				Error: &v1.Error{
					Code:    "provider-not-found",
					Message: "provider not found: " + req.Provider,
				},
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !result.Success {
		return &v1.VerifyResponse{
			Success:  false,
			Provider: req.Provider,
			Error: &v1.Error{
				Code:    "validation-failed",
				Message: "token validation failed",
				Details: result.ErrorCodes,
			},
		}, nil
	}

	return &v1.VerifyResponse{
		Success:  true,
		Provider: req.Provider,
		Result: &v1.ValidationResult{
			Success:     result.Success,
			ChallengeTs: result.ChallengeTS.Format("2006-01-02T15:04:05Z07:00"),
			Hostname:    result.Hostname,
			Action:      result.Action,
			CData:       result.CData,
			ErrorCodes:  result.ErrorCodes,
			Metadata:    result.Metadata,
		},
	}, nil
}

func (h *Handler) ListProviders(ctx context.Context, req *v1.ListProvidersRequest) (*v1.ListProvidersResponse, error) {
	providers := h.validator.ListProviders()
	return &v1.ListProvidersResponse{
		Providers: providers,
	}, nil
}
