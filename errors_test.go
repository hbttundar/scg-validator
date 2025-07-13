package validator

import (
	"testing"
)

func TestErrors_Get(t *testing.T) {
	tests := []struct {
		name     string
		errors   Errors
		field    string
		expected []string
	}{
		{
			name:     "existing field",
			errors:   Errors{"name": {"Name is required", "Name must be at least 3 characters"}},
			field:    "name",
			expected: []string{"Name is required", "Name must be at least 3 characters"},
		},
		{
			name:     "non-existing field",
			errors:   Errors{"name": {"Name is required"}},
			field:    "email",
			expected: []string{},
		},
		{
			name:     "empty errors",
			errors:   Errors{},
			field:    "name",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Get(tt.field)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d messages, got %d", len(tt.expected), len(result))
				return
			}
			for i, msg := range result {
				if msg != tt.expected[i] {
					t.Errorf("Expected message %q at index %d, got %q", tt.expected[i], i, msg)
				}
			}
		})
	}
}

func TestErrors_First(t *testing.T) {
	tests := []struct {
		name     string
		errors   Errors
		field    string
		expected string
	}{
		{
			name:     "field with multiple errors",
			errors:   Errors{"name": {"Name is required", "Name must be at least 3 characters"}},
			field:    "name",
			expected: "Name is required",
		},
		{
			name:     "field with single error",
			errors:   Errors{"email": {"Email is invalid"}},
			field:    "email",
			expected: "Email is invalid",
		},
		{
			name:     "non-existing field",
			errors:   Errors{"name": {"Name is required"}},
			field:    "email",
			expected: "",
		},
		{
			name:     "empty errors",
			errors:   Errors{},
			field:    "name",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.First(tt.field)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestErrors_Has(t *testing.T) {
	tests := []struct {
		name     string
		errors   Errors
		field    string
		expected bool
	}{
		{
			name:     "existing field",
			errors:   Errors{"name": {"Name is required"}},
			field:    "name",
			expected: true,
		},
		{
			name:     "non-existing field",
			errors:   Errors{"name": {"Name is required"}},
			field:    "email",
			expected: false,
		},
		{
			name:     "empty errors",
			errors:   Errors{},
			field:    "name",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Has(tt.field)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   Errors
		contains []string
	}{
		{
			name:     "multiple fields with errors",
			errors:   Errors{"name": {"Name is required"}, "email": {"Email is invalid", "Email is required"}},
			contains: []string{"field 'name': Name is required", "field 'email': Email is invalid, Email is required"},
		},
		{
			name:     "single field with error",
			errors:   Errors{"name": {"Name is required"}},
			contains: []string{"field 'name': Name is required"},
		},
		{
			name:     "empty errors",
			errors:   Errors{},
			contains: []string{"validation failed with the following errors:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Error()
			for _, substr := range tt.contains {
				if !contains(result, substr) {
					t.Errorf("Expected error string to contain %q, but it doesn't: %q", substr, result)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestErrors_Add tests the Add method of the Errors type
func TestErrors_Add(t *testing.T) {
	tests := []struct {
		name           string
		initialErrors  Errors
		field          string
		message        string
		expectedErrors Errors
	}{
		{
			name:           "add to empty errors",
			initialErrors:  Errors{},
			field:          "name",
			message:        "Name is required",
			expectedErrors: Errors{"name": {"Name is required"}},
		},
		{
			name:           "add to existing field",
			initialErrors:  Errors{"name": {"Name is required"}},
			field:          "name",
			message:        "Name must be at least 3 characters",
			expectedErrors: Errors{"name": {"Name is required", "Name must be at least 3 characters"}},
		},
		{
			name:           "add to new field",
			initialErrors:  Errors{"name": {"Name is required"}},
			field:          "email",
			message:        "Email is invalid",
			expectedErrors: Errors{"name": {"Name is required"}, "email": {"Email is invalid"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.initialErrors
			errors.Add(tt.field, tt.message)

			// Check that the field has the expected messages
			messages := errors.Get(tt.field)
			expectedMessages := tt.expectedErrors[tt.field]
			if len(messages) != len(expectedMessages) {
				t.Errorf("Expected %d messages for field %q, got %d", len(expectedMessages), tt.field, len(messages))
				return
			}
			for i, msg := range messages {
				if msg != expectedMessages[i] {
					t.Errorf("Expected message %q at index %d for field %q, got %q", expectedMessages[i], i, tt.field, msg)
				}
			}

			// Check that all fields have the expected messages
			for field, expectedMessages := range tt.expectedErrors {
				messages := errors.Get(field)
				if len(messages) != len(expectedMessages) {
					t.Errorf("Expected %d messages for field %q, got %d", len(expectedMessages), field, len(messages))
					continue
				}
				for i, msg := range messages {
					if msg != expectedMessages[i] {
						t.Errorf("Expected message %q at index %d for field %q, got %q", expectedMessages[i], i, field, msg)
					}
				}
			}
		})
	}
}
