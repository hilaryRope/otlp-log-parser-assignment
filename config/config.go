package config

import (
	"flag"
	"fmt"
	"time"
)

// Config holds the application configuration
type Config struct {
	GRPCPort    int
	MetricsPort int

	// AttributeKey is the attribute key to track across Resource, Scope, and Log levels
	AttributeKey string

	// WindowDuration is the time window for aggregating and reporting counts
	WindowDuration time.Duration

	Debug bool
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	flag.IntVar(&cfg.GRPCPort, "port", 4317, "gRPC server port")
	flag.IntVar(&cfg.MetricsPort, "metrics-port", 9090, "Port for Prometheus metrics")
	flag.StringVar(&cfg.AttributeKey, "attribute-key", "service.name", "Attribute key to track")
	flag.DurationVar(&cfg.WindowDuration, "window-duration", 10*time.Second, "Window duration for reporting counts")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug logging")

	flag.Parse()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.GRPCPort <= 0 || c.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d (must be between 1 and 65535)", c.GRPCPort)
	}

	if c.MetricsPort <= 0 || c.MetricsPort > 65535 {
		return fmt.Errorf("invalid metrics port: %d (must be between 1 and 65535)", c.MetricsPort)
	}

	if c.MetricsPort == c.GRPCPort {
		return fmt.Errorf("gRPC and metrics ports cannot be the same")
	}

	if c.AttributeKey == "" {
		return fmt.Errorf("attribute-key cannot be empty")
	}

	if c.WindowDuration <= 0 {
		return fmt.Errorf("window-duration must be positive")
	}

	return nil
}
