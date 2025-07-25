package hubble

import (
	"fmt"

	"github.com/Azure/mcp-kubernetes/pkg/command"
	"github.com/Azure/mcp-kubernetes/pkg/config"
	"github.com/Azure/mcp-kubernetes/pkg/security"
	"github.com/Azure/mcp-kubernetes/pkg/tools"
)

// HubbleExecutor implements the CommandExecutor interface for hubble commands
type HubbleExecutor struct{}

// This line ensures HubbleExecutor implements the CommandExecutor interface
var _ tools.CommandExecutor = (*HubbleExecutor)(nil)

// NewExecutor creates a new HubbleExecutor instance
func NewExecutor() *HubbleExecutor {
	return &HubbleExecutor{}
}

// Execute handles hubble command execution
func (e *HubbleExecutor) Execute(params map[string]interface{}, cfg *config.ConfigData) (string, error) {
	hubbleCmd, ok := params["command"].(string)
	if !ok {
		return "", fmt.Errorf("invalid command parameter")
	}

	// Validate the command against security settings
	validator := security.NewValidator(cfg.SecurityConfig)
	err := validator.ValidateCommand(hubbleCmd, security.CommandTypeHubble)
	if err != nil {
		return "", err
	}

	// Execute the command
	process := command.NewShellProcess("hubble", cfg.Timeout)
	return process.Run(hubbleCmd)
}
