package helm

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

// RegisterHelm registers the helm tool
func RegisterHelm() mcp.Tool {
	return mcp.NewTool("call_helm",
		mcp.WithDescription("Run Helm package manager commands for Kubernetes"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("Full helm command to execute (e.g., 'helm list', 'helm install myapp ./chart', 'helm upgrade myapp ./chart')"),
		),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Call Helm",
			DestructiveHint: boolPtr(true),
		}),
	)
}
