package cloudflare

import (
	"captcha/internal/validator/providers"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	siteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

type Provider struct {
	secretKey string
	client    *http.Client
}

func NewProvider(config CloudflareConfig) *Provider {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	return &Provider{
		secretKey: config.SecretKey,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (p *Provider) Name() string {
	return "cloudflare"
}

type siteverifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
	ErrorCodes  []string `json:"error-codes"`
	Metadata    any      `json:"metadata"`
}

func (p *Provider) Validate(ctx context.Context, token, remoteIP string, extra map[string]string) (*providers.ValidationResult, error) {
	data := url.Values{}
	data.Set("secret", p.secretKey)
	data.Set("response", token)
	if remoteIP != "" {
		data.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, siteverifyURL, strings.NewReader(data.Encode()))
	if err != nil {
		return &providers.ValidationResult{
			Success:    false,
			ErrorCodes: []string{"internal-error"},
		}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return &providers.ValidationResult{
			Success:    false,
			ErrorCodes: []string{"internal-error"},
		}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &providers.ValidationResult{
			Success:    false,
			ErrorCodes: []string{"internal-error"},
		}, fmt.Errorf("failed to read response: %w", err)
	}

	var result siteverifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return &providers.ValidationResult{
			Success:    false,
			ErrorCodes: []string{"internal-error"},
		}, fmt.Errorf("failed to parse response: %w", err)
	}

	validationResult := &providers.ValidationResult{
		Success:    result.Success,
		Hostname:   result.Hostname,
		Action:     result.Action,
		CData:      result.CData,
		ErrorCodes: result.ErrorCodes,
	}

	if result.ChallengeTS != "" {
		if t, err := time.Parse(time.RFC3339, result.ChallengeTS); err == nil {
			validationResult.ChallengeTS = t
		}
	}

	if result.Metadata != nil {
		if metaBytes, err := json.Marshal(result.Metadata); err == nil {
			validationResult.Metadata = make(map[string]string)
			validationResult.Metadata["raw"] = string(metaBytes)
		}
	}

	return validationResult, nil
}
