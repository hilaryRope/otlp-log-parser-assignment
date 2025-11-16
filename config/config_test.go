package config

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				GRPCPort:       4317,
				MetricsPort:    9090,
				AttributeKey:   "service.name",
				WindowDuration: 10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid port - zero",
			config: Config{
				GRPCPort:       0,
				MetricsPort:    9090,
				AttributeKey:   "service.name",
				WindowDuration: 10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port - negative",
			config: Config{
				GRPCPort:       -1,
				MetricsPort:    9090,
				AttributeKey:   "service.name",
				WindowDuration: 10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too large",
			config: Config{
				GRPCPort:       70000,
				MetricsPort:    9090,
				AttributeKey:   "service.name",
				WindowDuration: 10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "empty attribute key",
			config: Config{
				GRPCPort:       4317,
				MetricsPort:    9090,
				AttributeKey:   "",
				WindowDuration: 10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid window duration - zero",
			config: Config{
				GRPCPort:       4317,
				MetricsPort:    9090,
				AttributeKey:   "service.name",
				WindowDuration: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid window duration - negative",
			config: Config{
				GRPCPort:       4317,
				MetricsPort:    9090,
				AttributeKey:   "service.name",
				WindowDuration: -1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
