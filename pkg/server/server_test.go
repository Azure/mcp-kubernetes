package server

import (
	"strings"
	"testing"

	"github.com/Azure/mcp-kubernetes/pkg/config"
)

func TestRunRejectsNonStdioTransport(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Transport = "sse"

	err := NewService(cfg).Run()
	if err == nil {
		t.Fatal("Run() error = nil, want an error for a non-stdio transport")
	}
	if !strings.Contains(err.Error(), "only stdio is supported") {
		t.Fatalf("Run() error = %q, want stdio-only validation error", err)
	}
}
