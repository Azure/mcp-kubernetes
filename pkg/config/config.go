package config

import (
	"fmt"
	"strings"

	"github.com/Azure/mcp-kubernetes/pkg/security"
	flag "github.com/spf13/pflag"
)

// ConfigData holds the global configuration
type ConfigData struct {
	// Map of additional tools enabled
	AdditionalTools map[string]bool
	// Command execution timeout in seconds
	Timeout int
	// Security configuration
	SecurityConfig *security.SecurityConfig

	// Command-line specific options
	Transport       string
	Host            string
	Port            int
	AccessLevel     string
	AllowNamespaces string
}

// NewConfig creates and returns a new configuration instance
func NewConfig() *ConfigData {
	return &ConfigData{
		AdditionalTools: make(map[string]bool),
		Timeout:         60,
		SecurityConfig:  security.NewSecurityConfig(),
		Transport:       "stdio",
		Port:            8000,
		AccessLevel:     "readonly",
		AllowNamespaces: "",
	}
}

// ParseFlags parses command line arguments and updates the configuration
func (cfg *ConfigData) ParseFlags() error {
	// Server configuration
	flag.StringVar(&cfg.Transport, "transport", "stdio", "Transport mechanism to use (stdio, sse or streamable-http)")
	flag.StringVar(&cfg.Host, "host", "127.0.0.1", "Host to listen for the server (only used with transport sse or streamable-http)")
	flag.IntVar(&cfg.Port, "port", 8000, "Port to listen for the server (only used with transport sse or streamable-http)")
	flag.IntVar(&cfg.Timeout, "timeout", 60, "Timeout for command execution in seconds, default is 60s")

	// Tools configuration
	additionalTools := flag.String("additional-tools", "",
		"Comma-separated list of additional tools to support (kubectl is always enabled). Available: helm,cilium,hubble")

	// Security settings
	flag.StringVar(&cfg.AccessLevel, "access-level", "readonly", "Access level (readonly, readwrite, or admin)")
	flag.StringVar(&cfg.AllowNamespaces, "allow-namespaces", "",
		"Comma-separated list of namespaces to allow (empty means all allowed)")

	flag.Parse()

	// Update security config with access level
	switch cfg.AccessLevel {
	case "readonly":
		cfg.SecurityConfig.AccessLevel = security.AccessLevelReadOnly
	case "readwrite":
		cfg.SecurityConfig.AccessLevel = security.AccessLevelReadWrite
	case "admin":
		cfg.SecurityConfig.AccessLevel = security.AccessLevelAdmin
	default:
		return fmt.Errorf("invalid access level '%s'. Valid values are: readonly, readwrite, admin", cfg.AccessLevel)
	}

	if cfg.AllowNamespaces != "" {
		cfg.SecurityConfig.SetAllowedNamespaces(cfg.AllowNamespaces)
	}

	// Parse additional tools
	if *additionalTools != "" {
		for _, tool := range strings.Split(*additionalTools, ",") {
			tool = strings.TrimSpace(tool)
			if tool == "" {
				continue
			}
			cfg.AdditionalTools[tool] = true
		}
	}

	return nil
}

var availableTools = []string{"kubectl", "helm", "cilium", "hubble"}

// IsToolSupported checks if a tool is supported
func IsToolSupported(tool string) bool {
	for _, t := range availableTools {
		if t == tool {
			return true
		}
	}
	return false
}
