package hubble

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// RegisterHubble registers the hubble tool
func RegisterHubble() mcp.Tool {
	return mcp.NewTool("hubble",
		mcp.WithDescription("Run Hubble observability commands for network monitoring and debugging"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("The hubble command to execute (e.g., 'hubble status', 'hubble observe', 'hubble list nodes')"),
		),
	)
}
