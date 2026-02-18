package echo

import (
	"captcha/internal/validator"
	"captcha/internal/validator/providers/cloudflare"
	"captcha/pkg/middleware"
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Config struct {
	Provider   string
	Mode       middleware.Mode
	GRPCClient middleware.Client
	Validator  *validator.Validator
}

func New(v *validator.Validator, cfg Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			provider := c.Request().Header.Get(middleware.HeaderProvider)
			if provider == "" {
				provider = cfg.Provider
			}

			token := c.Request().Header.Get(middleware.HeaderToken)
			remoteIP := c.Request().Header.Get(middleware.HeaderRemoteIP)
			if remoteIP == "" {
				remoteIP = c.RealIP()
			}

			if token == "" {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"error": map[string]string{
						"code":    "missing-input-response",
						"message": "captcha token is required",
					},
				})
			}

			switch cfg.Mode {
			case middleware.ModeDirect:
				if cfg.Validator == nil {
					return c.JSON(http.StatusInternalServerError, map[string]interface{}{
						"error": map[string]string{
							"code":    "internal-error",
							"message": "validator not configured",
						},
					})
				}

				result, err := cfg.Validator.Validate(context.Background(), provider, token, remoteIP, nil)
				if err != nil || !result.Success {
					return c.JSON(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "validation-failed",
							"message": "captcha validation failed",
							"details": result.ErrorCodes,
						},
					})
				}

			case middleware.ModeGRPC:
				if cfg.GRPCClient == nil {
					return c.JSON(http.StatusInternalServerError, map[string]interface{}{
						"error": map[string]string{
							"code":    "internal-error",
							"message": "gRPC client not configured",
						},
					})
				}

				resp, err := cfg.GRPCClient.Verify(context.Background(), provider, token, remoteIP, nil)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]interface{}{
						"error": map[string]string{
							"code":    "internal-error",
							"message": err.Error(),
						},
					})
				}

				if !resp.Success {
					return c.JSON(http.StatusBadRequest, map[string]interface{}{
						"error": resp.Error,
					})
				}
			}

			return next(c)
		}
	}
}

func Example() {
	v := validator.New()
	v.RegisterProvider(cloudflare.NewProvider(cloudflare.CloudflareConfig{
		SecretKey: "your-secret-key",
	}))

	_ = New(v, Config{
		Provider: "cloudflare",
		Mode:     middleware.ModeDirect,
	})
}
