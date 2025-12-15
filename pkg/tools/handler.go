package tools

import (
	"context"
	"fmt"

	"github.com/Azure/mcp-kubernetes/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// CreateToolHandler creates an adapter that converts CommandExecutor to the format expected by MCP server
func CreateToolHandler(executor CommandExecutor, cfg *config.ConfigData) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("arguments must be a map[string]interface{}, got %T", req.Params.Arguments)
			// Track failed tool invocation
			if cfg.TelemetryService != nil {
				cfg.TelemetryService.TrackToolInvocation(ctx, req.Params.Name, "", false)
			}
			return mcp.NewToolResultError(err.Error()), nil
		}

		result, err := executor.Execute(ctx, args, cfg)
		if cfg.TelemetryService != nil {
			operation, _ := args["operation"].(string)
			cfg.TelemetryService.TrackToolInvocation(ctx, req.Params.Name, operation, err == nil)
		}

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(result), nil
	}
}

// CreateToolHandlerWithName creates an adapter for tools that need the tool name injected
func CreateToolHandlerWithName(executor CommandExecutor, cfg *config.ConfigData, toolName string) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("arguments must be a map[string]interface{}, got %T", req.Params.Arguments)
			// Track failed tool invocation
			if cfg.TelemetryService != nil {
				cfg.TelemetryService.TrackToolInvocation(ctx, req.Params.Name, "", false)
			}
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Inject the tool name into the arguments
		args["_tool_name"] = toolName

		result, err := executor.Execute(ctx, args, cfg)
		if cfg.TelemetryService != nil {
			operation, _ := args["operation"].(string)
			cfg.TelemetryService.TrackToolInvocation(ctx, toolName, operation, err == nil)
		}

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(result), nil
	}
}
