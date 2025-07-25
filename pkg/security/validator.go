package security

import (
	"regexp"
	"strings"
)

// Command type constants
const (
	CommandTypeKubectl = "kubectl"
	CommandTypeHelm    = "helm"
	CommandTypeCilium  = "cilium"
	CommandTypeHubble  = "hubble"
)

var (
	// KubectlReadOperations defines kubectl operations that don't modify state
	KubectlReadOperations = []string{
		"get", "describe", "explain", "logs", "top", "auth", "config",
		"cluster-info", "api-resources", "api-versions", "version", "diff",
		"completion", "help", "kustomize", "options", "plugin", "proxy", "wait", "events",
	}

	// KubectlReadWriteOperations defines kubectl operations that modify state but are not admin operations
	KubectlReadWriteOperations = []string{
		"create", "delete", "apply", "expose", "run", "set", "rollout", "scale",
		"autoscale", "label", "annotate", "patch", "replace", "cp", "exec",
	}

	// KubectlAdminOperations defines kubectl operations that require admin privileges
	KubectlAdminOperations = []string{
		"cordon", "uncordon", "drain", "taint", "certificate",
	}

	// HelmReadOperations defines helm operations that don't modify state
	HelmReadOperations = []string{
		"get", "history", "list", "show", "status", "search", "repo",
		"env", "version", "verify", "completion", "help",
	}

	// CiliumReadOperations defines cilium operations that don't modify state
	CiliumReadOperations = []string{
		"status", "version", "config", "help", "context", "connectivity",
		"endpoint", "identity", "ip", "map", "metrics", "monitor", "policy",
		"hubble", "bpf", "list", "observe", "service",
	}

	// HubbleReadOperations defines hubble operations that don't modify state
	HubbleReadOperations = []string{
		"status", "version", "help", "observe", "status", "list", "config",
	}
)

// Validator handles validation of commands against security configuration
type Validator struct {
	secConfig *SecurityConfig
}

// NewValidator creates a new Validator instance with the given security configuration
func NewValidator(secConfig *SecurityConfig) *Validator {
	return &Validator{
		secConfig: secConfig,
	}
}

// ValidationError represents a security validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// getReadOperationsList returns the appropriate list of read operations based on command type
func (v *Validator) getReadOperationsList(commandType string) []string {
	switch commandType {
	case CommandTypeKubectl:
		return KubectlReadOperations
	case CommandTypeHelm:
		return HelmReadOperations
	case CommandTypeCilium:
		return CiliumReadOperations
	case CommandTypeHubble:
		return HubbleReadOperations
	default:
		return []string{}
	}
}

// getReadWriteOperationsList returns the appropriate list of read-write operations based on command type
func (v *Validator) getReadWriteOperationsList(commandType string) []string {
	switch commandType {
	case CommandTypeKubectl:
		return KubectlReadWriteOperations
	case CommandTypeHelm:
		// For now, assume helm write operations are same as read operations
		// This can be expanded when helm write operations are defined
		return []string{}
	case CommandTypeCilium:
		// For now, assume cilium write operations are same as read operations
		// This can be expanded when cilium write operations are defined
		return []string{}
	case CommandTypeHubble:
		// For now, assume hubble write operations are same as read operations
		// This can be expanded when hubble write operations are defined
		return []string{}
	default:
		return []string{}
	}
}

// getAdminOperationsList returns the appropriate list of admin operations based on command type
func (v *Validator) getAdminOperationsList(commandType string) []string {
	switch commandType {
	case CommandTypeKubectl:
		return KubectlAdminOperations
	case CommandTypeHelm:
		// For now, assume helm admin operations are not defined
		// This can be expanded when helm admin operations are defined
		return []string{}
	case CommandTypeCilium:
		// For now, assume cilium admin operations are not defined
		// This can be expanded when cilium admin operations are defined
		return []string{}
	case CommandTypeHubble:
		// For now, assume hubble admin operations are not defined
		// This can be expanded when hubble admin operations are defined
		return []string{}
	default:
		return []string{}
	}
}

// ValidateCommand validates a command against all security settings
func (v *Validator) ValidateCommand(command, commandType string) error {
	// Check access level restrictions
	if err := v.validateAccessLevel(command, commandType); err != nil {
		return err
	}

	// Check namespace scope restrictions
	if err := v.validateNamespaceScope(command); err != nil {
		return err
	}

	return nil
}

// validateAccessLevel validates if a command is allowed based on the configured access level
func (v *Validator) validateAccessLevel(command, commandType string) error {
	readOperations := v.getReadOperationsList(commandType)
	readWriteOperations := v.getReadWriteOperationsList(commandType)
	adminOperations := v.getAdminOperationsList(commandType)

	operation := v.extractOperationFromCommand(command, commandType)

	switch v.secConfig.AccessLevel {
	case AccessLevelReadOnly:
		if !v.isOperationInList(operation, readOperations) {
			return &ValidationError{Message: "Error: Cannot execute write or admin operations in read-only mode"}
		}
	case AccessLevelReadWrite:
		if !v.isOperationInList(operation, readOperations) && !v.isOperationInList(operation, readWriteOperations) {
			// Check if it's an admin operation to provide better error message
			if v.isOperationInList(operation, adminOperations) {
				return &ValidationError{Message: "Error: Cannot execute admin operations in read-write mode"}
			}
			return &ValidationError{Message: "Error: Operation not allowed in read-write mode"}
		}
	case AccessLevelAdmin:
		// Admin level allows all operations (read, write, and admin)
		if !v.isOperationInList(operation, readOperations) &&
			!v.isOperationInList(operation, readWriteOperations) &&
			!v.isOperationInList(operation, adminOperations) {
			return &ValidationError{Message: "Error: Unknown operation"}
		}
	default:
		return &ValidationError{Message: "Error: Invalid access level configuration"}
	}

	return nil
}

// validateNamespaceScope validates if a command's namespace scope is allowed by security settings
func (v *Validator) validateNamespaceScope(command string) error {
	// Extract namespace from command
	namespace := v.extractNamespaceFromCommand(command)

	// If command applies to all namespaces, and there are namespace restrictions
	if namespace == "*" && (len(v.secConfig.allowedNamespaces) > 0 || len(v.secConfig.allowedNamespacesRe) > 0) {
		return &ValidationError{Message: "Error: Access to all namespaces is restricted by security configuration"}
	}

	// If a namespace is specified (or default "default" is used), check if it's allowed
	if namespace != "" && namespace != "*" {
		if !v.secConfig.IsNamespaceAllowed(namespace) {
			return &ValidationError{
				Message: "Error: Access to namespace '" + namespace + "' is denied by security configuration",
			}
		}
	}

	return nil
}

// isOperationInList checks if an operation is in the given list
func (v *Validator) isOperationInList(operation string, allowedOperations []string) bool {
	for _, allowed := range allowedOperations {
		if operation == allowed {
			return true
		}
	}
	return false
}

// extractOperationFromCommand extracts the operation from a command
func (v *Validator) extractOperationFromCommand(command, commandType string) string {
	cmdParts := strings.Fields(command)
	var operation string

	for _, part := range cmdParts {
		if !strings.HasPrefix(part, "-") {
			// Skip the initial command name (kubectl, helm, cilium)
			if part != commandType {
				operation = part
				break
			}
		}
	}

	return operation
}

// extractNamespaceFromCommand extracts the namespace from a command
func (v *Validator) extractNamespaceFromCommand(command string) string {
	// Check for explicit namespace parameter
	namespacePattern := `(?:-n|--namespace)[\s=]([^\s]+)`
	re := regexp.MustCompile(namespacePattern)
	matches := re.FindStringSubmatch(command)
	if len(matches) > 1 {
		return matches[1]
	}

	// Check if there's a format like <resource>/<name>
	resourcePattern := `(\S+)/(\S+)`
	re = regexp.MustCompile(resourcePattern)
	if re.MatchString(command) {
		// If the command contains resource/name format but no explicit namespace,
		// the default namespace "default" will be used
		return "default"
	}

	// Check for --all-namespaces or -A
	if strings.Contains(command, "--all-namespaces") || strings.Contains(command, " -A") || strings.HasSuffix(command, " -A") {
		return "*" // Special marker indicating all namespaces
	}

	return "" // No namespace found, default namespace will be used
}
