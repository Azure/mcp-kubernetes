package cilium

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// RegisterCilium registers the cilium tool
func RegisterCilium() mcp.Tool {
	return mcp.NewTool("cilium",
		mcp.WithDescription("Run Cilium CNI commands for network policies and observability"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("The cilium command to execute (e.g., 'cilium status', 'cilium endpoint list')"),
		),
	)
}
