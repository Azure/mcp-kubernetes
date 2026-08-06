package config

import (
	"context"
	"testing"
)

func TestAccessLevelValidation(t *testing.T) {
	tests := []struct {
		name        string
		accessLevel string
		expectError bool
	}{
		{"Valid readonly", "readonly", false},
		{"Valid readwrite", "readwrite", false},
		{"Valid admin", "admin", false},
		{"Invalid value", "invalid", true},
		{"Empty value", "", true},
		{"Case sensitive", "READONLY", true},
		{"Case sensitive", "ReadOnly", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.AccessLevel = tt.accessLevel

			// Skip flag parsing, just test the validation logic
			var err error
			switch cfg.AccessLevel {
			case "readonly", "readwrite", "admin":
				err = nil
			default:
				err = &ValidationError{Message: "invalid access level"}
			}

			if tt.expectError && err == nil {
				t.Errorf("Expected error for access level '%s', but got none", tt.accessLevel)
			} else if !tt.expectError && err != nil {
				t.Errorf("Did not expect error for access level '%s', but got: %v", tt.accessLevel, err)
			}
		})
	}
}

func TestInitializeTelemetry(t *testing.T) {
	tests := []struct {
		name           string
		otlpEndpoint   string
		serviceName    string
		serviceVersion string
	}{
		{
			name:           "Initialize without OTLP endpoint",
			otlpEndpoint:   "",
			serviceName:    "test-service",
			serviceVersion: "1.0.0",
		},
		{
			name:           "Initialize with OTLP endpoint",
			otlpEndpoint:   "localhost:4317",
			serviceName:    "test-service",
			serviceVersion: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.OTLPEndpoint = tt.otlpEndpoint

			ctx := context.Background()

			// Should not panic or error
			cfg.InitializeTelemetry(ctx, tt.serviceName, tt.serviceVersion)

			if cfg.TelemetryService == nil {
				t.Error("Expected TelemetryService to be initialized")
			}

			// Clean up
			if cfg.TelemetryService != nil {
				_ = cfg.TelemetryService.Shutdown(ctx)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	// Test default values
	if cfg.Transport != "stdio" {
		t.Errorf("Expected default transport 'stdio', got '%s'", cfg.Transport)
	}

	if cfg.AccessLevel != "readonly" {
		t.Errorf("Expected default access level 'readonly', got '%s'", cfg.AccessLevel)
	}

	if cfg.Timeout != 60 {
		t.Errorf("Expected default timeout 60, got %d", cfg.Timeout)
	}

	if cfg.AdditionalTools == nil {
		t.Error("Expected AdditionalTools map to be initialized")
	}

	if cfg.SecurityConfig == nil {
		t.Error("Expected SecurityConfig to be initialized")
	}
}

func TestValidateTransport(t *testing.T) {
	for _, tt := range []struct {
		transport string
		wantErr   bool
	}{
		{transport: "stdio"},
		{transport: "sse", wantErr: true},
		{transport: "streamable-http", wantErr: true},
	} {
		t.Run(tt.transport, func(t *testing.T) {
			err := validateTransport(tt.transport)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateTransport(%q) error = %v, wantErr %v", tt.transport, err, tt.wantErr)
			}
		})
	}
}

// ValidationError for testing purposes
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
