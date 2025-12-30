package config

import (
	"fmt"
	"os/exec"
)

// Validator handles all validation logic for MCP Kubernetes
type Validator struct {
	// Configuration to validate
	config *ConfigData
	// Errors discovered during validation
	errors []string
}

// NewValidator creates a new validator instance
func NewValidator(cfg *ConfigData) *Validator {
	return &Validator{
		config: cfg,
		errors: make([]string, 0),
	}
}

// isCliInstalled checks if a CLI tool is installed and available in the system PATH
func (v *Validator) isCliInstalled(cliName string) bool {
	_, err := exec.LookPath(cliName)
	return err == nil
}

// validateCli checks if the required CLI tools are installed
func (v *Validator) validateCli() bool {
	valid := true

	// kubectl is always required
	if !v.isCliInstalled("kubectl") {
		v.errors = append(v.errors, "kubectl is not installed or not found in PATH")
		valid = false
	}

	// Check additional tools based on configuration
	for tool := range v.config.AdditionalTools {
		if !v.isCliInstalled(tool) {
			v.errors = append(v.errors, fmt.Sprintf("%s is enabled but not installed or not found in PATH", tool))
			valid = false
		}
	}

	return valid
}

// validateAdditionalTools checks if the configured additional tools are supported
func (v *Validator) validateAdditionalTools() bool {
	valid := true

	for tool := range v.config.AdditionalTools {
		if !IsToolSupported(tool) {
			v.errors = append(v.errors, fmt.Sprintf("Unsupported tool: %s", tool))
			valid = false
		}
	}

	return valid
}

// validateKubeconfig checks if kubectl is properly configured and can connect to the cluster
func (v *Validator) validateKubeconfig() bool {
	cmd := exec.Command("kubectl", "version")
	if err := cmd.Run(); err != nil {
		v.errors = append(v.errors, "kubectl is not properly configured or cannot connect to the cluster: "+err.Error())
		return false
	}
	return true
}

// Validate runs all validation checks
func (v *Validator) Validate() bool {
	// Reset errors before validation
	v.errors = make([]string, 0)

	// Run all validation checks
	validTools := v.validateAdditionalTools()
	validCli := v.validateCli()
	validKubeconfig := v.validateKubeconfig()

	return validTools && validCli && validKubeconfig
}

// GetErrors returns all errors found during validation
func (v *Validator) GetErrors() []string {
	return v.errors
}

// PrintErrors prints all validation errors to stdout
func (v *Validator) PrintErrors() {
	for _, err := range v.errors {
		fmt.Println(err)
	}
}
