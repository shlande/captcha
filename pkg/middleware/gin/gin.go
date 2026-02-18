package gin

import (
	"captcha/internal/validator"
	"captcha/internal/validator/providers"
	"captcha/internal/validator/providers/cloudflare"
	"captcha/pkg/middleware"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Config struct {
	Provider   string
	Mode       middleware.Mode
	GRPCClient middleware.Client
	Validator  *validator.Validator
	ProxyURL   string
}

func New(v *validator.Validator, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.GetHeader(middleware.HeaderProvider)
		if provider == "" {
			provider = cfg.Provider
		}

		token := c.GetHeader(middleware.HeaderToken)
		remoteIP := c.GetHeader(middleware.HeaderRemoteIP)
		if remoteIP == "" {
			remoteIP = c.ClientIP()
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "missing-input-response",
					"message": "captcha token is required",
				},
			})
			return
		}

		switch cfg.Mode {
		case middleware.ModeDirect:
			if cfg.Validator == nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "internal-error",
						"message": "validator not configured",
					},
				})
				return
			}

			result, err := cfg.Validator.Validate(context.Background(), provider, token, remoteIP, nil)
			handleResult(c, result, err)

		case middleware.ModeGRPC:
			if cfg.GRPCClient == nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "internal-error",
						"message": "gRPC client not configured",
					},
				})
				return
			}

			resp, err := cfg.GRPCClient.Verify(context.Background(), provider, token, remoteIP, nil)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "internal-error",
						"message": err.Error(),
					},
				})
				return
			}

			if !resp.Success {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": resp.Error,
				})
				return
			}

			c.Next()
		}
	}
}

func handleResult(c *gin.Context, result *providers.ValidationResult, err error) {
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "internal-error",
				"message": err.Error(),
			},
		})
		return
	}

	if !result.Success {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "validation-failed",
				"message": "captcha validation failed",
				"details": result.ErrorCodes,
			},
		})
		return
	}

	c.Next()
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
