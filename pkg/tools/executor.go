package tools

import (
	"context"
	"github.com/Azure/mcp-kubernetes/pkg/config"
)

// CommandExecutor defines the interface for executing commands
// This ensures all command executors follow the same pattern and signature
type CommandExecutor interface {
	Execute(ctx context.Context, params map[string]interface{}, cfg *config.ConfigData) (string, error)
}
