package helm

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// RegisterHelm registers the helm tool
func RegisterHelm() mcp.Tool {
	return mcp.NewTool("call_helm",
		mcp.WithDescription("Run Helm package manager commands for Kubernetes"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("Full helm command to execute (e.g., 'helm list', 'helm install myapp ./chart', 'helm upgrade myapp ./chart')"),
		),
	)
}
