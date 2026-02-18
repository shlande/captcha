package validator

import (
	"captcha/internal/validator/providers"
	"context"
	"errors"
)

var (
	ErrProviderNotFound = errors.New("provider not found")
	ErrInvalidToken     = errors.New("invalid token")
)

type Validator struct {
	providers map[string]providers.Provider
}

func New() *Validator {
	return &Validator{
		providers: make(map[string]providers.Provider),
	}
}

func (v *Validator) RegisterProvider(p providers.Provider) {
	v.providers[p.Name()] = p
}

func (v *Validator) GetProvider(name string) (providers.Provider, bool) {
	p, ok := v.providers[name]
	return p, ok
}

func (v *Validator) Validate(ctx context.Context, providerName, token, remoteIP string, extra map[string]string) (*providers.ValidationResult, error) {
	p, ok := v.providers[providerName]
	if !ok {
		return nil, ErrProviderNotFound
	}

	if token == "" {
		return &providers.ValidationResult{
			Success:    false,
			ErrorCodes: []string{"missing-input-response"},
		}, nil
	}

	return p.Validate(ctx, token, remoteIP, extra)
}

func (v *Validator) ListProviders() []string {
	names := make([]string, 0, len(v.providers))
	for name := range v.providers {
		names = append(names, name)
	}
	return names
}
