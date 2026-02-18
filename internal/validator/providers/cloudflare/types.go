package cloudflare

import "time"

type CloudflareConfig struct {
	SecretKey string
	Timeout   time.Duration
}
