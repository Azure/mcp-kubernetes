package security

import (
	"regexp"
	"strings"
)

// AccessLevel defines the level of access allowed
type AccessLevel string

const (
	AccessLevelReadOnly   AccessLevel = "readonly"
	AccessLevelReadWrite  AccessLevel = "readwrite"
	AccessLevelAdmin      AccessLevel = "admin"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// AccessLevel defines the level of access allowed (readonly, readwrite, admin)
	AccessLevel AccessLevel
	// AllowedNamespaces is a list of literal namespace names
	allowedNamespaces []string
	// allowedNamespacesRe is a list of compiled regex patterns for namespace matching
	allowedNamespacesRe []*regexp.Regexp
}

// NewSecurityConfig creates a new SecurityConfig instance
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AccessLevel:         AccessLevelReadOnly,
		allowedNamespaces:   []string{},
		allowedNamespacesRe: []*regexp.Regexp{},
	}
}

// SetAllowedNamespaces sets the list of allowed namespaces
func (s *SecurityConfig) SetAllowedNamespaces(namespaces string) {
	s.allowedNamespaces = []string{}
	s.allowedNamespacesRe = []*regexp.Regexp{}

	if namespaces == "" {
		return
	}

	regexSpecialChars := ".*+?[](){}|^$\\"

	for _, ns := range strings.Split(namespaces, ",") {
		ns = strings.TrimSpace(ns)
		if ns == "" {
			continue
		}

		// Check if the namespace pattern contains regex special characters
		isRegex := false
		for _, char := range ns {
			if strings.ContainsRune(regexSpecialChars, char) {
				isRegex = true
				break
			}
		}

		if isRegex {
			// Try to compile the regex pattern
			re, err := regexp.Compile("^" + ns + "$")
			if err == nil {
				s.allowedNamespacesRe = append(s.allowedNamespacesRe, re)
			} else {
				// If regex is invalid, treat it as a literal string
				s.allowedNamespaces = append(s.allowedNamespaces, ns)
			}
		} else {
			s.allowedNamespaces = append(s.allowedNamespaces, ns)
		}
	}
}

// IsNamespaceAllowed checks if a namespace is allowed to be accessed
func (s *SecurityConfig) IsNamespaceAllowed(namespace string) bool {
	// If no restrictions are defined, allow all namespaces
	if len(s.allowedNamespaces) == 0 && len(s.allowedNamespacesRe) == 0 {
		return true
	}

	// Check for exact match in allowed namespaces
	for _, ns := range s.allowedNamespaces {
		if ns == namespace {
			return true
		}
	}

	// Check for match against regex patterns
	for _, pattern := range s.allowedNamespacesRe {
		if pattern.MatchString(namespace) {
			return true
		}
	}

	return false
}
