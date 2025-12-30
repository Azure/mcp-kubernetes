package hubble

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

// RegisterHubble registers the hubble tool
func RegisterHubble() mcp.Tool {
	return mcp.NewTool("call_hubble",
		mcp.WithDescription("Run Hubble observability commands for network monitoring and debugging"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("Full hubble command to execute (e.g., 'hubble status', 'hubble observe', 'hubble list nodes')"),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Call Hubble",
			ReadOnlyHint: boolPtr(true),
		}),
	)
}
