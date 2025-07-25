package server

import (
	"fmt"
	"log"

	"github.com/Azure/mcp-kubernetes/pkg/cilium"
	"github.com/Azure/mcp-kubernetes/pkg/config"
	"github.com/Azure/mcp-kubernetes/pkg/helm"
	"github.com/Azure/mcp-kubernetes/pkg/hubble"
	"github.com/Azure/mcp-kubernetes/pkg/kubectl"
	"github.com/Azure/mcp-kubernetes/pkg/tools"
	"github.com/Azure/mcp-kubernetes/pkg/version"
	"github.com/mark3labs/mcp-go/server"
)

// Service represents the MCP Kubernetes service
type Service struct {
	cfg       *config.ConfigData
	mcpServer *server.MCPServer
}

// NewService creates a new MCP Kubernetes service
func NewService(cfg *config.ConfigData) *Service {
	return &Service{
		cfg: cfg,
	}
}

// Initialize initializes the service
func (s *Service) Initialize() error {
	// Initialize configuration

	// Create MCP server
	s.mcpServer = server.NewMCPServer(
		"MCP Kubernetes",
		version.GetVersion(),
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	// Register individual kubectl commands based on permission level
	s.registerKubectlCommands()

	// Register additional tools
	if s.cfg.AdditionalTools["helm"] {
		helmTool := helm.RegisterHelm()
		s.mcpServer.AddTool(helmTool, tools.CreateToolHandler(helm.NewExecutor(), s.cfg))
	}

	if s.cfg.AdditionalTools["cilium"] {
		ciliumTool := cilium.RegisterCilium()
		s.mcpServer.AddTool(ciliumTool, tools.CreateToolHandler(cilium.NewExecutor(), s.cfg))
	}

	if s.cfg.AdditionalTools["hubble"] {
		hubbleTool := hubble.RegisterHubble()
		s.mcpServer.AddTool(hubbleTool, tools.CreateToolHandler(hubble.NewExecutor(), s.cfg))
	}

	return nil
}

// Run starts the service with the specified transport
func (s *Service) Run() error {
	log.Println("MCP Kubernetes version:", version.GetVersion())

	// Start the server
	switch s.cfg.Transport {
	case "stdio":
		log.Println("Listening for requests on STDIO...")
		return server.ServeStdio(s.mcpServer)
	case "sse":
		sse := server.NewSSEServer(s.mcpServer)
		addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
		log.Printf("SSE server listening on %s", addr)
		return sse.Start(addr)
	case "streamable-http":
		streamableServer := server.NewStreamableHTTPServer(s.mcpServer)
		addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
		log.Printf("Streamable HTTP server listening on %s", addr)
		return streamableServer.Start(addr)
	default:
		return fmt.Errorf("invalid transport type: %s (must be 'stdio', 'sse' or 'streamable-http')", s.cfg.Transport)
	}
}

// registerKubectlCommands registers kubectl tools based on access level
func (s *Service) registerKubectlCommands() {
	// Get kubectl tools filtered by access level
	kubectlTools := kubectl.RegisterKubectlTools(s.cfg.AccessLevel)

	// Create a kubectl executor
	kubectlExecutor := kubectl.NewKubectlToolExecutor()

	// Register each kubectl tool
	for _, tool := range kubectlTools {
		// Create a handler that injects the tool name into params
		handler := tools.CreateToolHandlerWithName(kubectlExecutor, s.cfg, tool.Name)
		s.mcpServer.AddTool(tool, handler)
	}
}
