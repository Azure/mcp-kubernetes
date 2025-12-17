package kubectl

import (
	"strings"
	"testing"
)

func TestRegisterKubectlTools(t *testing.T) {
	// Test admin access level gets all tools with specialized mode
	tools := RegisterKubectlTools("admin", false)

	// Verify we have the expected number of tools
	expectedCount := 6
	if len(tools) != expectedCount {
		t.Errorf("Expected %d consolidated tools, got %d", expectedCount, len(tools))
	}

	// Verify tool names
	expectedNames := GetKubectlToolNames()
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expectedName := range expectedNames {
		if !toolNames[expectedName] {
			t.Errorf("Expected tool %s not found", expectedName)
		}
	}
}

func TestRegisterKubectlTools_UnifiedMode(t *testing.T) {
	// Test that unified mode returns only call_kubectl tool
	tools := RegisterKubectlTools("admin", true)

	// Verify we have exactly one tool
	if len(tools) != 1 {
		t.Errorf("Expected 1 unified tool, got %d", len(tools))
	}

	// Verify it's the call_kubectl tool
	if tools[0].Name != "call_kubectl" {
		t.Errorf("Expected tool name 'call_kubectl', got '%s'", tools[0].Name)
	}

	// Verify description contains key indicators
	desc := tools[0].Description
	expectedContent := []string{
		"Pass full kubectl command including 'kubectl' prefix",
		"kubectl get pods",
		"command=",
	}
	for _, content := range expectedContent {
		if !strings.Contains(desc, content) {
			t.Errorf("Unified tool description missing content: %s", content)
		}
	}
}

func TestConsolidatedToolDescriptions(t *testing.T) {
	// Test with admin access to get all tools in specialized mode
	tools := RegisterKubectlTools("admin", false)

	tests := []struct {
		toolName           string
		expectedOperations []string
		expectedInDesc     []string
	}{
		{
			toolName:           "kubectl_resources",
			expectedOperations: []string{"get", "describe", "create", "delete", "apply", "patch", "replace"},
			expectedInDesc:     []string{"CRUD operations", "Examples:"},
		},
		{
			toolName:           "kubectl_workloads",
			expectedOperations: []string{"run", "expose", "scale", "autoscale", "rollout"},
			expectedInDesc:     []string{"workloads", "lifecycle", "Examples:"},
		},
		{
			toolName:           "kubectl_metadata",
			expectedOperations: []string{"label", "annotate", "set"},
			expectedInDesc:     []string{"metadata", "Examples:"},
		},
		{
			toolName:           "kubectl_diagnostics",
			expectedOperations: []string{"logs", "events", "top", "exec", "cp"},
			expectedInDesc:     []string{"Diagnose", "debug", "Examples:"},
		},
		{
			toolName:           "kubectl_cluster",
			expectedOperations: []string{"cluster-info", "api-resources", "api-versions", "explain"},
			expectedInDesc:     []string{"cluster", "API", "Examples:"},
		},
		{
			toolName:           "kubectl_config",
			expectedOperations: []string{"diff", "auth", "certificate", "config"},
			expectedInDesc:     []string{"configurations", "Examples:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			var found bool
			var toolDesc string

			for _, tool := range tools {
				if tool.Name == tt.toolName {
					found = true
					toolDesc = tool.Description
					break
				}
			}

			if !found {
				t.Errorf("Tool %s not found", tt.toolName)
				return
			}

			// Check that all expected operations are mentioned
			for _, op := range tt.expectedOperations {
				if !strings.Contains(toolDesc, op) {
					t.Errorf("Tool %s description missing operation: %s", tt.toolName, op)
				}
			}

			// Check for expected content in description
			for _, expected := range tt.expectedInDesc {
				if !strings.Contains(toolDesc, expected) {
					t.Errorf("Tool %s description missing expected content: %s", tt.toolName, expected)
				}
			}
		})
	}
}

func TestConsolidatedToolParameters(t *testing.T) {
	// Test with admin access to get all tools in specialized mode
	tools := RegisterKubectlTools("admin", false)

	// All tools should have the same three parameters
	// expectedParams := []string{"operation", "resource", "args"}

	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			// We can't directly access tool parameters in the test,
			// but we can verify the tool was created successfully
			if tool.Name == "" {
				t.Error("Tool has empty name")
			}
			if tool.Description == "" {
				t.Error("Tool has empty description")
			}
		})
	}
}

func TestGetKubectlToolNames(t *testing.T) {
	names := GetKubectlToolNames()

	expected := []string{
		"kubectl_resources",
		"kubectl_workloads",
		"kubectl_metadata",
		"kubectl_diagnostics",
		"kubectl_cluster",
		"kubectl_config",
	}

	if len(names) != len(expected) {
		t.Errorf("Expected %d tool names, got %d", len(expected), len(names))
	}

	for i, name := range expected {
		if names[i] != name {
			t.Errorf("Expected tool name %s at index %d, got %s", name, i, names[i])
		}
	}
}

func TestMapOperationToCommand_AllTools(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		operation string
		resource  string
		want      string
		wantErr   bool
	}{
		// kubectl_resources tests
		{
			name:      "resources get",
			toolName:  "kubectl_resources",
			operation: "get",
			resource:  "pods",
			want:      "get",
		},
		{
			name:      "resources apply",
			toolName:  "kubectl_resources",
			operation: "apply",
			resource:  "",
			want:      "apply",
		},
		// kubectl_workloads tests
		{
			name:      "workloads scale",
			toolName:  "kubectl_workloads",
			operation: "scale",
			resource:  "deployment",
			want:      "scale",
		},
		{
			name:      "workloads rollout status",
			toolName:  "kubectl_workloads",
			operation: "rollout",
			resource:  "status",
			want:      "rollout status",
		},
		{
			name:      "workloads rollout undo",
			toolName:  "kubectl_workloads",
			operation: "rollout",
			resource:  "undo",
			want:      "rollout undo",
		},
		// kubectl_metadata tests
		{
			name:      "metadata label",
			toolName:  "kubectl_metadata",
			operation: "label",
			resource:  "pods",
			want:      "label",
		},
		{
			name:      "metadata set",
			toolName:  "kubectl_metadata",
			operation: "set",
			resource:  "image",
			want:      "set",
		},
		// kubectl_diagnostics tests
		{
			name:      "diagnostics logs",
			toolName:  "kubectl_diagnostics",
			operation: "logs",
			resource:  "pod",
			want:      "logs",
		},
		{
			name:      "diagnostics exec",
			toolName:  "kubectl_diagnostics",
			operation: "exec",
			resource:  "pod",
			want:      "exec",
		},
		// kubectl_cluster tests
		{
			name:      "cluster info",
			toolName:  "kubectl_cluster",
			operation: "cluster-info",
			resource:  "",
			want:      "cluster-info",
		},
		{
			name:      "cluster explain",
			toolName:  "kubectl_cluster",
			operation: "explain",
			resource:  "pod",
			want:      "explain",
		},
		// Node operations in kubectl_resources
		{
			name:      "resources cordon",
			toolName:  "kubectl_resources",
			operation: "cordon",
			resource:  "node",
			want:      "cordon",
		},
		{
			name:      "resources taint",
			toolName:  "kubectl_resources",
			operation: "taint",
			resource:  "nodes",
			want:      "taint",
		},
		// kubectl_config tests
		{
			name:      "config diff",
			toolName:  "kubectl_config",
			operation: "diff",
			resource:  "",
			want:      "diff",
		},
		{
			name:      "config auth can-i",
			toolName:  "kubectl_config",
			operation: "auth",
			resource:  "can-i",
			want:      "auth can-i",
		},
		{
			name:      "config certificate approve",
			toolName:  "kubectl_config",
			operation: "certificate",
			resource:  "approve",
			want:      "certificate approve",
		},
		// kubectl_config config operations tests
		{
			name:      "config current-context",
			toolName:  "kubectl_config",
			operation: "config",
			resource:  "current-context",
			want:      "config current-context",
		},
		{
			name:      "config get-contexts",
			toolName:  "kubectl_config",
			operation: "config",
			resource:  "get-contexts",
			want:      "config get-contexts",
		},
		{
			name:      "config use-context",
			toolName:  "kubectl_config",
			operation: "config",
			resource:  "use-context",
			want:      "config use-context",
		},
		// Unknown tool test
		{
			name:      "unknown tool",
			toolName:  "kubectl_unknown",
			operation: "get",
			resource:  "pods",
			want:      "",
			wantErr:   false, // MapOperationToCommand returns empty string for unknown tools
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapOperationToCommand(tt.toolName, tt.operation, tt.resource)
			if tt.wantErr && err == nil {
				t.Errorf("MapOperationToCommand() error = nil, wantErr %v", tt.wantErr)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("MapOperationToCommand() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Errorf("MapOperationToCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterKubectlTools_AccessLevelFiltering(t *testing.T) {
	tests := []struct {
		name            string
		accessLevel     string
		expectedTools   []string
		unexpectedTools []string
	}{
		{
			name:        "readonly access level",
			accessLevel: "readonly",
			expectedTools: []string{
				"kubectl_resources",
				"kubectl_diagnostics",
				"kubectl_cluster",
				"kubectl_config",
			},
			unexpectedTools: []string{
				"kubectl_workloads",
				"kubectl_metadata",
			},
		},
		{
			name:        "readwrite access level",
			accessLevel: "readwrite",
			expectedTools: []string{
				"kubectl_resources",
				"kubectl_workloads",
				"kubectl_metadata",
				"kubectl_diagnostics",
				"kubectl_cluster",
				"kubectl_config",
			},
			unexpectedTools: []string{},
		},
		{
			name:        "admin access level",
			accessLevel: "admin",
			expectedTools: []string{
				"kubectl_resources",
				"kubectl_workloads",
				"kubectl_metadata",
				"kubectl_diagnostics",
				"kubectl_cluster",
				"kubectl_config",
			},
			unexpectedTools: []string{}, // admin has access to all tools
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := RegisterKubectlTools(tt.accessLevel, false)

			// Check that expected tools are present
			for _, expectedTool := range tt.expectedTools {
				found := false
				for _, tool := range tools {
					if tool.Name == expectedTool {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tool %s not found for access level %s", expectedTool, tt.accessLevel)
				}
			}

			// Check that unexpected tools are not present
			for _, unexpectedTool := range tt.unexpectedTools {
				for _, tool := range tools {
					if tool.Name == unexpectedTool {
						t.Errorf("Unexpected tool %s found for access level %s", unexpectedTool, tt.accessLevel)
					}
				}
			}
		})
	}
}

func TestRegisterKubectlTools_ReadOnlyDescriptions(t *testing.T) {
	// Test that read-only access level has appropriate descriptions in specialized mode
	tools := RegisterKubectlTools("readonly", false)

	for _, tool := range tools {
		switch tool.Name {
		case "kubectl_resources":
			// Check that description doesn't mention write operations
			if strings.Contains(tool.Description, "create") ||
				strings.Contains(tool.Description, "delete") ||
				strings.Contains(tool.Description, "apply") {
				t.Errorf("kubectl_resources in readonly mode should not mention write operations")
			}
			// Check that it mentions read operations
			if !strings.Contains(tool.Description, "get") ||
				!strings.Contains(tool.Description, "describe") {
				t.Errorf("kubectl_resources in readonly mode should mention read operations")
			}
		case "kubectl_config":
			// Check that description doesn't mention certificate operations
			if strings.Contains(tool.Description, "certificate") {
				t.Errorf("kubectl_config in readonly mode should not mention certificate operations")
			}
		}
	}
}

func TestRegisterKubectlTools_DefaultsToReadOnly(t *testing.T) {
	// Test that unknown access level defaults to readonly
	tools := RegisterKubectlTools("unknown", false)
	readonlyTools := RegisterKubectlTools("readonly", false)

	if len(tools) != len(readonlyTools) {
		t.Errorf("Unknown access level should default to readonly, got %d tools, expected %d",
			len(tools), len(readonlyTools))
	}

	// Verify tool names match
	for i, tool := range tools {
		if tool.Name != readonlyTools[i].Name {
			t.Errorf("Tool mismatch at index %d: got %s, expected %s",
				i, tool.Name, readonlyTools[i].Name)
		}
	}
}

func TestCallKubectlToolStructure(t *testing.T) {
	// Test that call_kubectl tool has simplified structure with only args parameter
	testCases := []struct {
		name        string
		accessLevel string
		mustContain []string
	}{
		{
			name:        "readonly",
			accessLevel: "readonly",
			mustContain: []string{
				"Pass full kubectl command including 'kubectl' prefix",
				"kubectl get pods -n default",
				"kubectl describe deployment myapp",
			},
		},
		{
			name:        "readwrite",
			accessLevel: "readwrite",
			mustContain: []string{
				"Pass full kubectl command including 'kubectl' prefix",
				"kubectl get pods -n default",
				"kubectl create -f deployment.yaml",
			},
		},
		{
			name:        "admin",
			accessLevel: "admin",
			mustContain: []string{
				"Pass full kubectl command including 'kubectl' prefix",
				"kubectl get pods -n default",
				"kubectl cordon worker-1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tools := RegisterKubectlTools(tc.accessLevel, true)

			if len(tools) != 1 {
				t.Fatalf("Expected 1 tool, got %d", len(tools))
			}

			tool := tools[0]
			if tool.Name != "call_kubectl" {
				t.Errorf("Expected tool name 'call_kubectl', got '%s'", tool.Name)
			}

			// Verify description contains expected content
			for _, expected := range tc.mustContain {
				if !strings.Contains(tool.Description, expected) {
					t.Errorf("call_kubectl description missing: %s", expected)
				}
			}
		})
	}
}
