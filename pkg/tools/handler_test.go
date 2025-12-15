package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/mcp-kubernetes/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
	"go.opentelemetry.io/otel/trace"
)

// Mock CommandExecutor for testing
type mockExecutor struct {
	shouldError bool
	result      string
}

func (m *mockExecutor) Execute(ctx context.Context, args map[string]interface{}, cfg *config.ConfigData) (string, error) {
	if m.shouldError {
		return "", errors.New("mock execution error")
	}
	return m.result, nil
}

// Mock TelemetryService for testing
type mockTelemetryService struct {
	invocations []invocation
}

type invocation struct {
	toolName  string
	operation string
	success   bool
}

func (m *mockTelemetryService) TrackToolInvocation(ctx context.Context, toolName string, operation string, success bool) {
	m.invocations = append(m.invocations, invocation{
		toolName:  toolName,
		operation: operation,
		success:   success,
	})
}

// Implement other methods to satisfy interface
func (m *mockTelemetryService) Initialize(ctx context.Context) error    { return nil }
func (m *mockTelemetryService) Shutdown(ctx context.Context) error      { return nil }
func (m *mockTelemetryService) TrackServiceStartup(ctx context.Context) {}
func (m *mockTelemetryService) StartActivity(ctx context.Context, name string) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx)
}

func TestCreateToolHandlerSuccess(t *testing.T) {
	executor := &mockExecutor{
		shouldError: false,
		result:      "success result",
	}

	mockTelemetry := &mockTelemetryService{}
	cfg := &config.ConfigData{}
	cfg.TelemetryService = mockTelemetry

	handler := CreateToolHandler(executor, cfg)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "test-tool",
			Arguments: map[string]interface{}{
				"operation": "get",
				"arg1":      "value1",
			},
		},
	}

	ctx := context.Background()
	result, err := handler(ctx, req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected result content to be non-empty")
	}

	// Check if telemetry was tracked
	if len(mockTelemetry.invocations) != 1 {
		t.Errorf("Expected 1 telemetry invocation, got %d", len(mockTelemetry.invocations))
	}

	invocation := mockTelemetry.invocations[0]
	if invocation.toolName != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got '%s'", invocation.toolName)
	}
	if invocation.operation != "get" {
		t.Errorf("Expected operation 'get', got '%s'", invocation.operation)
	}
	if !invocation.success {
		t.Error("Expected success to be true")
	}
}

func TestCreateToolHandlerError(t *testing.T) {
	executor := &mockExecutor{
		shouldError: true,
		result:      "",
	}

	mockTelemetry := &mockTelemetryService{}
	cfg := &config.ConfigData{}
	cfg.TelemetryService = mockTelemetry

	handler := CreateToolHandler(executor, cfg)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "test-tool",
			Arguments: map[string]interface{}{
				"operation": "delete",
				"arg1":      "value1",
			},
		},
	}

	ctx := context.Background()
	result, err := handler(ctx, req)

	if err != nil {
		t.Errorf("Expected no error from handler, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Check if telemetry was tracked with failure
	if len(mockTelemetry.invocations) != 1 {
		t.Errorf("Expected 1 telemetry invocation, got %d", len(mockTelemetry.invocations))
	}

	invocation := mockTelemetry.invocations[0]
	if invocation.toolName != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got '%s'", invocation.toolName)
	}
	if invocation.operation != "delete" {
		t.Errorf("Expected operation 'delete', got '%s'", invocation.operation)
	}
	if invocation.success {
		t.Error("Expected success to be false")
	}
}

func TestCreateToolHandlerWithoutTelemetry(t *testing.T) {
	executor := &mockExecutor{
		shouldError: false,
		result:      "success result",
	}

	cfg := &config.ConfigData{}
	// TelemetryService is nil

	handler := CreateToolHandler(executor, cfg)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "test-tool",
			Arguments: map[string]interface{}{
				"arg1": "value1",
			},
		},
	}

	ctx := context.Background()
	result, err := handler(ctx, req)

	// Should not panic or error when telemetry service is nil
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}
}

func TestCreateToolHandlerInvalidArguments(t *testing.T) {
	executor := &mockExecutor{
		shouldError: false,
		result:      "success result",
	}

	mockTelemetry := &mockTelemetryService{}
	cfg := &config.ConfigData{}
	cfg.TelemetryService = mockTelemetry

	handler := CreateToolHandler(executor, cfg)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "test-tool",
			Arguments: "invalid arguments", // Should be map[string]interface{}
		},
	}

	ctx := context.Background()
	result, err := handler(ctx, req)

	if err != nil {
		t.Errorf("Expected no error from handler, got %v", err)
	}

	// Should contain error in result - check the actual error structure
	if len(result.Content) == 0 {
		t.Error("Expected result to contain error content")
	}

	// Telemetry should track failed invocation for invalid arguments
	if len(mockTelemetry.invocations) != 1 {
		t.Errorf("Expected 1 telemetry invocation for invalid arguments, got %d", len(mockTelemetry.invocations))
	}

	if len(mockTelemetry.invocations) > 0 {
		invocation := mockTelemetry.invocations[0]
		if invocation.success {
			t.Error("Expected success to be false for invalid arguments")
		}
	}
}

func TestCreateToolHandlerWithNameSuccess(t *testing.T) {
	executor := &mockExecutor{
		shouldError: false,
		result:      "success result",
	}

	mockTelemetry := &mockTelemetryService{}
	cfg := &config.ConfigData{}
	cfg.TelemetryService = mockTelemetry

	toolName := "named-tool"
	handler := CreateToolHandlerWithName(executor, cfg, toolName)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "request-tool-name", // This should be overridden
			Arguments: map[string]interface{}{
				"operation": "apply",
				"arg1":      "value1",
			},
		},
	}

	ctx := context.Background()
	result, err := handler(ctx, req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Check if telemetry was tracked with the injected tool name
	if len(mockTelemetry.invocations) != 1 {
		t.Errorf("Expected 1 telemetry invocation, got %d", len(mockTelemetry.invocations))
	}

	invocation := mockTelemetry.invocations[0]
	if invocation.toolName != toolName {
		t.Errorf("Expected tool name '%s', got '%s'", toolName, invocation.toolName)
	}
	if invocation.operation != "apply" {
		t.Errorf("Expected operation 'apply', got '%s'", invocation.operation)
	}
	if !invocation.success {
		t.Error("Expected success to be true")
	}
}

func TestCreateToolHandlerWithNameError(t *testing.T) {
	executor := &mockExecutor{
		shouldError: true,
		result:      "",
	}

	mockTelemetry := &mockTelemetryService{}
	cfg := &config.ConfigData{}
	cfg.TelemetryService = mockTelemetry

	toolName := "named-tool"
	handler := CreateToolHandlerWithName(executor, cfg, toolName)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "request-tool-name",
			Arguments: map[string]interface{}{
				"operation": "patch",
				"arg1":      "value1",
			},
		},
	}

	ctx := context.Background()
	result, err := handler(ctx, req)

	if err != nil {
		t.Errorf("Expected no error from handler, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Check if telemetry was tracked with failure
	if len(mockTelemetry.invocations) != 1 {
		t.Errorf("Expected 1 telemetry invocation, got %d", len(mockTelemetry.invocations))
	}

	invocation := mockTelemetry.invocations[0]
	if invocation.toolName != toolName {
		t.Errorf("Expected tool name '%s', got '%s'", toolName, invocation.toolName)
	}
	if invocation.operation != "patch" {
		t.Errorf("Expected operation 'patch', got '%s'", invocation.operation)
	}
	if invocation.success {
		t.Error("Expected success to be false")
	}
}
