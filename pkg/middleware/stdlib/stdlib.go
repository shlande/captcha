package stdlib

import (
	"captcha/internal/validator"
	"captcha/internal/validator/providers/cloudflare"
	"captcha/pkg/middleware"
	"context"
	"net/http"
)

type Config struct {
	Provider   string
	Mode       middleware.Mode
	GRPCClient middleware.Client
	Validator  *validator.Validator
}

func New(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provider := r.Header.Get(middleware.HeaderProvider)
			if provider == "" {
				provider = cfg.Provider
			}

			token := r.Header.Get(middleware.HeaderToken)
			remoteIP := r.Header.Get(middleware.HeaderRemoteIP)
			if remoteIP == "" {
				remoteIP = r.RemoteAddr
			}

			if token == "" {
				http.Error(w, `{"error":{"code":"missing-input-response","message":"captcha token is required"}}`, http.StatusBadRequest)
				return
			}

			switch cfg.Mode {
			case middleware.ModeDirect:
				if cfg.Validator == nil {
					http.Error(w, `{"error":{"code":"internal-error","message":"validator not configured"}}`, http.StatusInternalServerError)
					return
				}

				result, err := cfg.Validator.Validate(context.Background(), provider, token, remoteIP, nil)
				if err != nil || !result.Success {
					http.Error(w, `{"error":{"code":"validation-failed","message":"captcha validation failed"}}`, http.StatusBadRequest)
					return
				}

			case middleware.ModeGRPC:
				if cfg.GRPCClient == nil {
					http.Error(w, `{"error":{"code":"internal-error","message":"gRPC client not configured"}}`, http.StatusInternalServerError)
					return
				}

				resp, err := cfg.GRPCClient.Verify(context.Background(), provider, token, remoteIP, nil)
				if err != nil {
					http.Error(w, `{"error":{"code":"internal-error","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
					return
				}

				if !resp.Success {
					http.Error(w, `{"error":{"code":"validation-failed","message":"captcha validation failed"}}`, http.StatusBadRequest)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func Example() {
	v := validator.New()
	v.RegisterProvider(cloudflare.NewProvider(cloudflare.CloudflareConfig{
		SecretKey: "your-secret-key",
	}))

	_ = New(Config{
		Provider:  "cloudflare",
		Mode:      middleware.ModeDirect,
		Validator: v,
	})
}
