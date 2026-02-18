package providers

import (
	"context"
	"time"
)

type ValidationResult struct {
	Success     bool              `json:"success"`
	ChallengeTS time.Time         `json:"challenge_ts,omitempty"`
	Hostname    string            `json:"hostname,omitempty"`
	Action      string            `json:"action,omitempty"`
	CData       string            `json:"cdata,omitempty"`
	ErrorCodes  []string          `json:"error-codes,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type Provider interface {
	Name() string
	Validate(ctx context.Context, token, remoteIP string, extra map[string]string) (*ValidationResult, error)
}
