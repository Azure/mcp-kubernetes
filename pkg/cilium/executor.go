package cilium

import (
	"context"
	"fmt"

	"github.com/Azure/mcp-kubernetes/pkg/command"
	"github.com/Azure/mcp-kubernetes/pkg/config"
	"github.com/Azure/mcp-kubernetes/pkg/security"
	"github.com/Azure/mcp-kubernetes/pkg/tools"
)

// CiliumExecutor implements the CommandExecutor interface for cilium commands
type CiliumExecutor struct{}

// This line ensures CiliumExecutor implements the CommandExecutor interface
var _ tools.CommandExecutor = (*CiliumExecutor)(nil)

// NewExecutor creates a new CiliumExecutor instance
func NewExecutor() *CiliumExecutor {
	return &CiliumExecutor{}
}

// Execute handles cilium command execution
func (e *CiliumExecutor) Execute(ctx context.Context, params map[string]interface{}, cfg *config.ConfigData) (string, error) {
	ciliumCmd, ok := params["command"].(string)
	if !ok {
		return "", fmt.Errorf("invalid command parameter")
	}

	// Validate the command against security settings
	validator := security.NewValidator(cfg.SecurityConfig)
	err := validator.ValidateCommand(ciliumCmd, security.CommandTypeCilium)
	if err != nil {
		return "", err
	}

	// Execute the command
	process := command.NewShellProcess("cilium", cfg.Timeout)
	return process.Run(ciliumCmd)
}
