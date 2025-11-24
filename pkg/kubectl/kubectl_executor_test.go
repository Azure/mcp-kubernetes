package kubectl

import (
	"strings"
	"testing"

	"github.com/Azure/mcp-kubernetes/pkg/config"
	"github.com/Azure/mcp-kubernetes/pkg/security"
)

func TestKubectlToolExecutor_ValidateCombination(t *testing.T) {
	executor := NewKubectlToolExecutor()

	tests := []struct {
		name      string
		toolName  string
		operation string
		resource  string
		wantErr   bool
		errMsg    string
	}{
		// Resources tool tests
		{
			name:      "valid resources get",
			toolName:  "kubectl_resources",
			operation: "get",
			resource:  "pods",
			wantErr:   false,
		},
		{
			name:      "invalid resources operation",
			toolName:  "kubectl_resources",
			operation: "rollout",
			resource:  "pods",
			wantErr:   true,
			errMsg:    "invalid operation 'rollout' for resources tool",
		},
		// Workloads tool tests
		{
			name:      "valid workloads scale",
			toolName:  "kubectl_workloads",
			operation: "scale",
			resource:  "deployment",
			wantErr:   false,
		},
		{
			name:      "valid rollout status",
			toolName:  "kubectl_workloads",
			operation: "rollout",
			resource:  "status",
			wantErr:   false,
		},
		{
			name:      "invalid rollout subcommand",
			toolName:  "kubectl_workloads",
			operation: "rollout",
			resource:  "invalid",
			wantErr:   true,
			errMsg:    "invalid rollout subcommand 'invalid'",
		},
		// Metadata tool tests
		{
			name:      "valid metadata label",
			toolName:  "kubectl_metadata",
			operation: "label",
			resource:  "pods",
			wantErr:   false,
		},
		// Diagnostics tool tests
		{
			name:      "valid diagnostics logs",
			toolName:  "kubectl_diagnostics",
			operation: "logs",
			resource:  "pod",
			wantErr:   false,
		},
		// Cluster tool tests
		{
			name:      "valid cluster info",
			toolName:  "kubectl_cluster",
			operation: "cluster-info",
			resource:  "",
			wantErr:   false,
		},
		// Node operations in resources tool
		{
			name:      "valid resources cordon",
			toolName:  "kubectl_resources",
			operation: "cordon",
			resource:  "node",
			wantErr:   false,
		},
		{
			name:      "valid resources drain",
			toolName:  "kubectl_resources",
			operation: "drain",
			resource:  "node",
			wantErr:   false,
		},
		// Config tool tests
		{
			name:      "valid config diff",
			toolName:  "kubectl_config",
			operation: "diff",
			resource:  "",
			wantErr:   false,
		},
		{
			name:      "valid auth can-i",
			toolName:  "kubectl_config",
			operation: "auth",
			resource:  "can-i",
			wantErr:   false,
		},
		{
			name:      "invalid auth resource",
			toolName:  "kubectl_config",
			operation: "auth",
			resource:  "invalid",
			wantErr:   true,
			errMsg:    "auth operation requires 'can-i' as resource",
		},
		{
			name:      "valid certificate approve",
			toolName:  "kubectl_config",
			operation: "certificate",
			resource:  "approve",
			wantErr:   false,
		},
		// Config operations tests
		{
			name:      "valid config current-context",
			toolName:  "kubectl_config",
			operation: "config",
			resource:  "current-context",
			wantErr:   false,
		},
		{
			name:      "valid config get-contexts",
			toolName:  "kubectl_config",
			operation: "config",
			resource:  "get-contexts",
			wantErr:   false,
		},
		{
			name:      "valid config use-context",
			toolName:  "kubectl_config",
			operation: "config",
			resource:  "use-context",
			wantErr:   false,
		},
		{
			name:      "invalid config resource",
			toolName:  "kubectl_config",
			operation: "config",
			resource:  "invalid-resource",
			wantErr:   true,
			errMsg:    "invalid config subcommand 'invalid-resource'",
		},
		{
			name:      "invalid config operation",
			toolName:  "kubectl_config",
			operation: "invalid-operation",
			resource:  "",
			wantErr:   true,
			errMsg:    "invalid operation 'invalid-operation' for config tool",
		},
		// Unknown tool test
		{
			name:      "unknown tool",
			toolName:  "kubectl_unknown",
			operation: "get",
			resource:  "pods",
			wantErr:   true,
			errMsg:    "unknown tool: kubectl_unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.validateCombination(tt.toolName, tt.operation, tt.resource)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateCombination() error = nil, wantErr %v", tt.wantErr)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateCombination() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateCombination() unexpected error = %v", err)
			}
		})
	}
}

func TestKubectlToolExecutor_BuildCommand(t *testing.T) {
	executor := NewKubectlToolExecutor()

	tests := []struct {
		name           string
		kubectlCommand string
		resource       string
		args           string
		want           string
	}{
		{
			name:           "simple command with resource and args",
			kubectlCommand: "get",
			resource:       "pods",
			args:           "-n default",
			want:           "get pods -n default",
		},
		{
			name:           "command without resource",
			kubectlCommand: "cluster-info",
			resource:       "",
			args:           "--kubeconfig=/path/to/config",
			want:           "cluster-info --kubeconfig=/path/to/config",
		},
		{
			name:           "command with subcommand already included",
			kubectlCommand: "rollout status",
			resource:       "deployment/myapp",
			args:           "-n production",
			want:           "rollout status -n production",
		},
		{
			name:           "command with empty args",
			kubectlCommand: "get",
			resource:       "nodes",
			args:           "",
			want:           "get nodes",
		},
		{
			name:           "auth can-i command",
			kubectlCommand: "auth can-i",
			resource:       "",
			args:           "create pods --namespace=default",
			want:           "auth can-i create pods --namespace=default",
		},
		{
			name:           "apply with file flag",
			kubectlCommand: "apply",
			resource:       "",
			args:           "-f deployment.yaml",
			want:           "apply -f deployment.yaml",
		},
		{
			name:           "exec command ignores resource parameter",
			kubectlCommand: "exec",
			resource:       "pod",
			args:           "mypod -- date",
			want:           "exec mypod -- date",
		},
		{
			name:           "cp command ignores resource parameter",
			kubectlCommand: "cp",
			resource:       "",
			args:           "/tmp/foo_dir some-pod:/tmp/bar_dir",
			want:           "cp /tmp/foo_dir some-pod:/tmp/bar_dir",
		},
		{
			name:           "events command with empty resource",
			kubectlCommand: "events",
			resource:       "",
			args:           "--all-namespaces",
			want:           "events --all-namespaces",
		},
		{
			name:           "api-resources command with empty resource",
			kubectlCommand: "api-resources",
			resource:       "",
			args:           "--namespaced=true",
			want:           "api-resources --namespaced=true",
		},
		{
			name:           "api-versions command with empty resource",
			kubectlCommand: "api-versions",
			resource:       "",
			args:           "",
			want:           "api-versions",
		},
		{
			name:           "run command with empty resource",
			kubectlCommand: "run",
			resource:       "",
			args:           "nginx --image=nginx",
			want:           "run nginx --image=nginx",
		},
		{
			name:           "diff command with empty resource",
			kubectlCommand: "diff",
			resource:       "",
			args:           "-f deployment.yaml",
			want:           "diff -f deployment.yaml",
		},
		{
			name:           "logs command with empty resource",
			kubectlCommand: "logs",
			resource:       "",
			args:           "nginx -c ruby-container",
			want:           "logs nginx -c ruby-container",
		},
		{
			name:           "logs command with selector",
			kubectlCommand: "logs",
			resource:       "",
			args:           "-l app=nginx --all-containers=true",
			want:           "logs -l app=nginx --all-containers=true",
		},
		{
			name:           "top command with resource subcommand",
			kubectlCommand: "top",
			resource:       "pod",
			args:           "",
			want:           "top pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executor.buildCommand(tt.kubectlCommand, tt.resource, tt.args)
			if got != tt.want {
				t.Errorf("buildCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubectlToolExecutor_Execute(t *testing.T) {
	// Note: This test validates parameter extraction and command construction
	// but doesn't execute actual kubectl commands

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"_tool_name": "kubectl_resources",
				"operation":  "get",
				"resource":   "pods",
				"args":       "-n default",
			},
			wantErr: false,
		},
		{
			name: "missing operation",
			params: map[string]interface{}{
				"_tool_name": "kubectl_resources",
				"resource":   "pods",
				"args":       "-n default",
			},
			wantErr: true,
			errMsg:  "operation parameter is required",
		},
		{
			name: "missing resource",
			params: map[string]interface{}{
				"_tool_name": "kubectl_resources",
				"operation":  "get",
				"args":       "-n default",
			},
			wantErr: true,
			errMsg:  "resource parameter is required",
		},
		{
			name: "missing args",
			params: map[string]interface{}{
				"_tool_name": "kubectl_resources",
				"operation":  "get",
				"resource":   "pods",
			},
			wantErr: true,
			errMsg:  "args parameter is required",
		},
		{
			name: "invalid combination",
			params: map[string]interface{}{
				"_tool_name": "kubectl_resources",
				"operation":  "rollout",
				"resource":   "pods",
				"args":       "",
			},
			wantErr: true,
			errMsg:  "invalid operation 'rollout' for resources tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewKubectlToolExecutor()
			cfg := &config.ConfigData{
				AccessLevel: "readwrite",
				SecurityConfig: &security.SecurityConfig{
					AccessLevel: security.AccessLevelReadWrite,
				},
			}

			_, err := executor.Execute(tt.params, cfg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Execute() error = nil, wantErr %v", tt.wantErr)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Execute() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
			// Note: We don't test successful execution as it would require mocking kubectl
		})
	}
}

func TestMapOperationToCommand(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		operation string
		resource  string
		want      string
	}{
		{
			name:      "resources tool simple operation",
			toolName:  "kubectl_resources",
			operation: "get",
			resource:  "pods",
			want:      "get",
		},
		{
			name:      "workloads rollout with subcommand",
			toolName:  "kubectl_workloads",
			operation: "rollout",
			resource:  "status",
			want:      "rollout status",
		},
		{
			name:      "config auth with can-i",
			toolName:  "kubectl_config",
			operation: "auth",
			resource:  "can-i",
			want:      "auth can-i",
		},
		{
			name:      "config certificate with approve",
			toolName:  "kubectl_config",
			operation: "certificate",
			resource:  "approve",
			want:      "certificate approve",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapOperationToCommand(tt.toolName, tt.operation, tt.resource)
			if err != nil {
				t.Errorf("MapOperationToCommand() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Errorf("MapOperationToCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
