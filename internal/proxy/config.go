package proxy

import (
	"fmt"
	"time"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	HTTPHost string `yaml:"http_host"`
	HTTPPort int    `yaml:"http_port"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s ServerConfig) HTTPAddr() string {
	if s.HTTPHost == "" {
		s.HTTPHost = "0.0.0.0"
	}
	if s.HTTPPort == 0 {
		s.HTTPPort = 8080
	}
	return fmt.Sprintf("%s:%d", s.HTTPHost, s.HTTPPort)
}

type ProviderConfig struct {
	Cloudflare CloudflareProviderConfig `yaml:"cloudflare"`
}

type CloudflareProviderConfig struct {
	Enabled   bool          `yaml:"enabled"`
	SecretKey string        `yaml:"secret_key"`
	Timeout   time.Duration `yaml:"timeout"`
}
