package kubectl

import (
	"fmt"
	"strings"

	"github.com/Azure/mcp-kubernetes/pkg/config"
	"github.com/Azure/mcp-kubernetes/pkg/security"
)

// KubectlToolExecutor handles structured kubectl command execution for grouped tools
type KubectlToolExecutor struct {
	executor *KubectlExecutor
}

// NewKubectlToolExecutor creates a new kubectl tool executor
func NewKubectlToolExecutor() *KubectlToolExecutor {
	return &KubectlToolExecutor{
		executor: NewExecutor(),
	}
}

// Execute processes structured kubectl commands with operation/resource/args parameters
func (e *KubectlToolExecutor) Execute(params map[string]interface{}, cfg *config.ConfigData) (string, error) {
	// Get the tool name from params (injected by handler)
	toolName, _ := params["_tool_name"].(string)

	// Handle call_kubectl with simplified args-only parameter
	if toolName == "call_kubectl" {
		args, ok := params["args"].(string)
		if !ok {
			return "", fmt.Errorf("args parameter is required and must be a string")
		}

		// Use args directly as the kubectl command
		fullCommand := args

		// Validate the command against security settings (includes access level and namespace checks)
		validator := security.NewValidator(cfg.SecurityConfig)
		if err := validator.ValidateCommand(fullCommand, security.CommandTypeKubectl); err != nil {
			return "", err
		}

		// Execute the command directly
		return e.executor.executeKubectlCommand(fullCommand, "", cfg)
	}

	// Handle legacy specialized tools with operation/resource/args parameters
	// Extract structured parameters
	operation, ok := params["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required and must be a string")
	}

	resource, ok := params["resource"].(string)
	if !ok {
		return "", fmt.Errorf("resource parameter is required and must be a string")
	}

	args, ok := params["args"].(string)
	if !ok {
		return "", fmt.Errorf("args parameter is required and must be a string")
	}

	// Validate the operation/resource combination
	if err := e.validateCombination(toolName, operation, resource); err != nil {
		return "", err
	}

	// Map operation to kubectl command
	kubectlCommand, err := MapOperationToCommand(toolName, operation, resource)
	if err != nil {
		return "", err
	}

	// Build the full command
	fullCommand := e.buildCommand(kubectlCommand, resource, args)

	// Validate the command against security settings (includes access level and namespace checks)
	validator := security.NewValidator(cfg.SecurityConfig)
	if err := validator.ValidateCommand(fullCommand, security.CommandTypeKubectl); err != nil {
		return "", err
	}

	// Execute the command directly
	return e.executor.executeKubectlCommand(fullCommand, "", cfg)
}

// validateCombination validates if the operation/resource combination is valid for the tool
// Note: This is only used for legacy specialized tools (kubectl_resources, kubectl_workloads, etc.)
// The unified call_kubectl tool does not use this validation
func (e *KubectlToolExecutor) validateCombination(toolName, operation, resource string) error {
	switch toolName {
	case "kubectl_resources":
		return e.validateResourcesOperation(operation)
	case "kubectl_workloads":
		return e.validateWorkloadsOperation(operation, resource)
	case "kubectl_metadata":
		return e.validateMetadataOperation(operation)
	case "kubectl_diagnostics":
		return e.validateDiagnosticsOperation(operation)
	case "kubectl_cluster":
		return e.validateClusterOperation(operation)
	case "kubectl_config":
		return e.validateConfigOperation(operation, resource)
	default:
		return fmt.Errorf("unknown tool: %s", toolName)
	}
}

// validateResourcesOperation validates operations for the resources tool
func (e *KubectlToolExecutor) validateResourcesOperation(operation string) error {
	// Always allow read-only operations
	readOnlyOps := []string{"get", "describe"}
	for _, validOp := range readOnlyOps {
		if operation == validOp {
			return nil
		}
	}

	// For write operations, they will be validated by the access level check
	writeOps := []string{"create", "delete", "apply", "patch", "replace"}
	for _, validOp := range writeOps {
		if operation == validOp {
			return nil
		}
	}

	// Node operations (admin level)
	nodeOps := []string{"cordon", "uncordon", "drain", "taint"}
	for _, validOp := range nodeOps {
		if operation == validOp {
			return nil
		}
	}

	allOps := append(readOnlyOps, writeOps...)
	allOps = append(allOps, nodeOps...)
	return fmt.Errorf("invalid operation '%s' for resources tool. Valid operations: %s",
		operation, strings.Join(allOps, ", "))
}

// validateWorkloadsOperation validates operations for the workloads tool
func (e *KubectlToolExecutor) validateWorkloadsOperation(operation, resource string) error {
	validOps := []string{"run", "expose", "scale", "autoscale", "rollout"}
	for _, validOp := range validOps {
		if operation == validOp {
			// Special validation for rollout subcommands
			if operation == "rollout" {
				validSubcmds := []string{"status", "history", "undo", "restart", "pause", "resume"}
				for _, subcmd := range validSubcmds {
					if resource == subcmd {
						return nil
					}
				}
				return fmt.Errorf("invalid rollout subcommand '%s'. Valid subcommands: %s",
					resource, strings.Join(validSubcmds, ", "))
			}
			return nil
		}
	}
	return fmt.Errorf("invalid operation '%s' for workloads tool. Valid operations: %s",
		operation, strings.Join(validOps, ", "))
}

// validateMetadataOperation validates operations for the metadata tool
func (e *KubectlToolExecutor) validateMetadataOperation(operation string) error {
	validOps := []string{"label", "annotate", "set"}
	for _, validOp := range validOps {
		if operation == validOp {
			return nil
		}
	}
	return fmt.Errorf("invalid operation '%s' for metadata tool. Valid operations: %s",
		operation, strings.Join(validOps, ", "))
}

// validateDiagnosticsOperation validates operations for the diagnostics tool
func (e *KubectlToolExecutor) validateDiagnosticsOperation(operation string) error {
	validOps := []string{"logs", "events", "top", "exec", "cp"}
	for _, validOp := range validOps {
		if operation == validOp {
			return nil
		}
	}
	return fmt.Errorf("invalid operation '%s' for diagnostics tool. Valid operations: %s",
		operation, strings.Join(validOps, ", "))
}

// validateClusterOperation validates operations for the cluster tool
func (e *KubectlToolExecutor) validateClusterOperation(operation string) error {
	validOps := []string{"cluster-info", "api-resources", "api-versions", "explain"}
	for _, validOp := range validOps {
		if operation == validOp {
			return nil
		}
	}
	return fmt.Errorf("invalid operation '%s' for cluster tool. Valid operations: %s",
		operation, strings.Join(validOps, ", "))
}

// validateConfigOperation validates operations for the config tool
func (e *KubectlToolExecutor) validateConfigOperation(operation, resource string) error {
	// Always allow read-only operations
	switch operation {
	case "diff":
		return nil
	case "auth":
		if resource != "can-i" {
			return fmt.Errorf("auth operation requires 'can-i' as resource")
		}
		return nil
	case "certificate":
		// Certificate operations are write operations, validated by access level check
		validSubcmds := []string{"approve", "deny"}
		for _, subcmd := range validSubcmds {
			if resource == subcmd {
				return nil
			}
		}
		return fmt.Errorf("invalid certificate subcommand '%s'. Valid subcommands: %s",
			resource, strings.Join(validSubcmds, ", "))
	case "config":
		// Config operations for context and configuration management
		validSubcmds := []string{
			"current-context", "get-contexts", "use-context",
		}
		for _, subcmd := range validSubcmds {
			if resource == subcmd {
				return nil
			}
		}
		return fmt.Errorf("invalid config subcommand '%s'. Valid subcommands: %s",
			resource, strings.Join(validSubcmds, ", "))
	default:
		return fmt.Errorf("invalid operation '%s' for config tool. Valid operations: diff, auth, certificate, config",
			operation)
	}
}

// buildCommand constructs the full kubectl command
func (e *KubectlToolExecutor) buildCommand(kubectlCommand, resource, args string) string {
	// Handle special cases where resource is part of the command
	if strings.Contains(kubectlCommand, " ") {
		// Command already includes subcommand (e.g., "rollout status", "auth can-i")
		if args != "" {
			return fmt.Sprintf("%s %s", kubectlCommand, args)
		}
		return kubectlCommand
	}

	// Standard case: command + resource + args
	parts := []string{kubectlCommand}

	// Skip the resource parameter for certain commands
	skipResourceCommands := []string{"exec", "cp", "events", "cluster-info", "api-resources", "api-versions", "diff", "run", "logs"}
	shouldSkipResource := false
	for _, cmd := range skipResourceCommands {
		if kubectlCommand == cmd {
			shouldSkipResource = true
			break
		}
	}

	// Also skip resource if it's empty (file-based operations like create -f, apply -f, etc.)
	if resource != "" && !shouldSkipResource {
		parts = append(parts, resource)
	}

	// Add args if not empty
	if args != "" {
		parts = append(parts, args)
	}

	return strings.Join(parts, " ")
}

// GetCommandForValidation returns the constructed command for security validation
func (e *KubectlToolExecutor) GetCommandForValidation(operation, resource, args string, toolName string) string {
	kubectlCommand, _ := MapOperationToCommand(toolName, operation, resource)
	return e.buildCommand(kubectlCommand, resource, args)
}
