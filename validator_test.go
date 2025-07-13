package validator

import (
	"fmt"
	"reflect"
	"testing"
)

// mockPresenceVerifier is a test double for the PresenceVerifier interface.
type mockPresenceVerifier struct {
	// a simple map to simulate a database table
	db map[string]map[string]any
}

func (m *mockPresenceVerifier) Exists(
	table, column string,
	value any,
	excludeIDColumn, excludeIDValue string,
) (bool, error) {
	if tableData, ok := m.db[table]; ok {
		// Simulate finding a record
		if existingVal, found := tableData[column]; found {
			if fmt.Sprintf("%v", existingVal) == fmt.Sprintf("%v", value) {
				// We found a match. Now check if we should exclude it.
				if excludeIDValue != "" {
					if id, hasID := tableData[excludeIDColumn]; hasID && fmt.Sprintf("%v", id) == excludeIDValue {
						return false, nil // It's the same record, so it's "unique" for an update
					}
				}
				return true, nil // Record exists
			}
		}
	}
	return false, nil // Record does not exist
}

// Main test function with comprehensive table-driven tests.
func TestValidator_Validate_WithNewArchitecture(t *testing.T) {
	// --- Mock DB for 'unique' rule test ---
	mockDB := &mockPresenceVerifier{
		db: map[string]map[string]any{
			"users": {
				"email":  "taken@example.com",
				"id":     "123",
				"org_id": "org1",
			},
		},
	}

	// --- Test Cases ---
	testCases := []struct {
		name           string
		input          any
		rules          map[string]string
		options        []Option
		expectError    bool
		expectedErrors Errors
	}{
		{
			name:        "happy path - should pass",
			input:       map[string]any{"username": "gooduser", "age": 30},
			rules:       map[string]string{"username": "required|alpha", "age": "required|numeric|gt:18"},
			expectError: false,
		},
		{
			name:        "templating - :attribute and :param0",
			input:       map[string]any{"quantity": 0},
			rules:       map[string]string{"quantity": "gt:5"},
			expectError: true,
			expectedErrors: Errors{
				"quantity": []string{"the quantity field must be greater than 5"},
			},
		},
		{
			name:        "unique rule - email already exists",
			input:       map[string]any{"email": "taken@example.com"},
			rules:       map[string]string{"email": "unique:users,email"},
			options:     []Option{WithPresenceVerifier(mockDB)},
			expectError: true,
			expectedErrors: Errors{
				"email": []string{"the email has already been taken"},
			},
		},
		{
			name:        "unique rule - update own record should pass",
			input:       map[string]any{"id": "123", "email": "taken@example.com"},
			rules:       map[string]string{"email": "unique:users,email,id,id"},
			options:     []Option{WithPresenceVerifier(mockDB)},
			expectError: false,
		},
		{
			name:  "unique rule - no verifier provided",
			input: map[string]any{"email": "any@example.com"},
			rules: map[string]string{"email": "unique:users,email"},
			// No options with a verifier
			expectError: true,
			expectedErrors: Errors{
				"email": []string{"cannot use 'unique' rule without a PresenceVerifier"},
			},
		},
		{
			name:        "array rule - array should pass",
			input:       map[string]any{"items": [3]string{"a", "b", "c"}},
			rules:       map[string]string{"items": "array"},
			expectError: false,
		},
		{
			name:        "array rule - slice should pass",
			input:       map[string]any{"items": []string{"a", "b", "c"}},
			rules:       map[string]string{"items": "array"},
			expectError: false,
		},
		{
			name:        "array rule - map should pass",
			input:       map[string]any{"items": map[string]string{"a": "apple", "b": "banana"}},
			rules:       map[string]string{"items": "array"},
			expectError: false,
		},
		{
			name:        "array rule - string should fail",
			input:       map[string]any{"items": "not an array or map"},
			rules:       map[string]string{"items": "array"},
			expectError: true,
			expectedErrors: Errors{
				"items": []string{"the items field must be an array or map"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			validator := New(testCase.options...)

			err := validator.Validate(testCase.input, testCase.rules)

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected a validation error, but got nil")
				}
				validationErrors, ok := err.(Errors)
				if !ok {
					t.Fatalf("expected error of type validator.Errors, but got: %v", err)
				}
				if !reflect.DeepEqual(testCase.expectedErrors, validationErrors) {
					t.Errorf("incorrect validation errors:\n- want: %v\n- got:  %v", testCase.expectedErrors, validationErrors)
				}
			} else if err != nil {
				t.Fatalf("did not expect an error, but got: %v", err)
			}
		})
	}
}
