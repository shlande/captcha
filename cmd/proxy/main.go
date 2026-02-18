package main

import (
	"captcha/internal/proxy"
	"captcha/internal/validator"
	"captcha/internal/validator/providers/cloudflare"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type AppConfig struct {
	Server   proxy.ServerConfig   `yaml:"server"`
	Provider proxy.ProviderConfig `yaml:"providers"`
}

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	v := validator.New()

	if cfg.Provider.Cloudflare.Enabled {
		v.RegisterProvider(cloudflare.NewProvider(cloudflare.CloudflareConfig{
			SecretKey: cfg.Provider.Cloudflare.SecretKey,
			Timeout:   cfg.Provider.Cloudflare.Timeout,
		}))
	}

	handler := proxy.NewHandler(v)
	server := proxy.NewServer(cfg.Server, handler)
	httpServer := proxy.NewHTTPServer(cfg.Server)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		server.Stop()
		httpServer.Stop()
	}()

	go func() {
		if err := httpServer.Serve(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}
	}()

	fmt.Printf("Starting captcha proxy server on %s\n", cfg.Server.Addr())
	if err := server.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
