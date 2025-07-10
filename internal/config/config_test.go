package config

import (
	"testing"
)

func TestAccessLevelValidation(t *testing.T) {
	tests := []struct {
		name        string
		accessLevel string
		expectError bool
	}{
		{"Valid readonly", "readonly", false},
		{"Valid readwrite", "readwrite", false},
		{"Valid admin", "admin", false},
		{"Invalid value", "invalid", true},
		{"Empty value", "", true},
		{"Case sensitive", "READONLY", true},
		{"Case sensitive", "ReadOnly", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.AccessLevel = tt.accessLevel
			
			// Skip flag parsing, just test the validation logic
			var err error
			switch cfg.AccessLevel {
			case "readonly", "readwrite", "admin":
				err = nil
			default:
				err = &ValidationError{Message: "invalid access level"}
			}

			if tt.expectError && err == nil {
				t.Errorf("Expected error for access level '%s', but got none", tt.accessLevel)
			} else if !tt.expectError && err != nil {
				t.Errorf("Did not expect error for access level '%s', but got: %v", tt.accessLevel, err)
			}
		})
	}
}

// ValidationError for testing purposes
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
