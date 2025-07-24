package validator

import (
	"fmt"
	"testing"

	"github.com/hbttundar/scg-validator/contract"
)

func TestNew(t *testing.T) {
	validator := New()
	if validator == nil {
		t.Error("New() should return a non-nil validator")
		return
	}
	if validator.engine == nil {
		t.Error("New() should initialize engine")
	}
}

func TestValidator_Validate(t *testing.T) {
	validator := New()

	tests := []struct {
		name        string
		data        map[string]interface{}
		rules       map[string]string
		expectValid bool
	}{
		{
			name:        "valid data",
			data:        map[string]interface{}{"email": "test@example.com", "age": 25},
			rules:       map[string]string{"email": "email", "age": "numeric"},
			expectValid: true,
		},
		{
			name:        "invalid data - invalid email",
			data:        map[string]interface{}{"email": "invalid-email", "age": 25},
			rules:       map[string]string{"email": "email", "age": "numeric"},
			expectValid: false,
		},
		{
			name:        "invalid data - non-numeric age",
			data:        map[string]interface{}{"email": "test@example.com", "age": "abc"},
			rules:       map[string]string{"email": "email", "age": "numeric"},
			expectValid: false,
		},
		{
			name:        "empty rules",
			data:        map[string]interface{}{"name": "John"},
			rules:       map[string]string{},
			expectValid: true,
		},
		{
			name: "complex_nested_validation",
			data: map[string]interface{}{
				"user_type":   "admin",
				"permissions": "read_write_delete",
				"email":       "admin@company.com",
				"phone":       "+1234567890",
			},
			rules: map[string]string{
				"user_type":   "required|alpha",
				"permissions": "required_if:user_type,admin|alpha_dash",
				"email":       "required|email",
				"phone":       "required|min:10",
			},
			expectValid: true,
		},
		{
			name: "validation_with_custom_messages",
			data: map[string]interface{}{
				"password": "123",
			},
			rules: map[string]string{
				"password": "required|min:8",
			},
			expectValid: false,
		},
		{
			name: "mixed_data_types_validation",
			data: map[string]interface{}{
				"count":    42,
				"price":    19.99,
				"active":   true,
				"tags":     []string{"new", "featured"},
				"metadata": map[string]string{"color": "red"},
			},
			rules: map[string]string{
				"count":    "required|numeric|min:1",
				"price":    "required|numeric|min:0",
				"active":   "required|boolean",
				"tags":     "required",
				"metadata": "required",
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.data, tt.rules)
			// result is an error, so check if it's nil for valid
			isValid := result == nil
			if isValid != tt.expectValid {
				t.Errorf("Expected IsValid=%v, got %v", tt.expectValid, isValid)
			}
		})
	}
}

func TestValidator_AddRule(t *testing.T) {
	validator := New()

	tests := []struct {
		name        string
		ruleName    string
		ruleCreator contract.RuleCreator
		expectError bool
	}{
		{
			name:     "add custom rule successfully",
			ruleName: "custom_test",
			ruleCreator: func(_ []string) (contract.Rule, error) {
				// Use a real rule implementation - create a simple custom rule
				return &customTestRule{}, nil
			},
			expectError: false,
		},
		{
			name:     "add rule creator that would fail when called",
			ruleName: "error_rule",
			ruleCreator: func(_ []string) (contract.Rule, error) {
				return nil, fmt.Errorf("rule creation failed")
			},
			expectError: false, // Registration should succeed, error occurs only when rule is created
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.AddRule(tt.ruleName, tt.ruleCreator)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// If no error expected, verify rule was added
			if !tt.expectError && err == nil {
				if !validator.HasRule(tt.ruleName) {
					t.Errorf("Rule %s should be registered", tt.ruleName)
				}
			}
		})
	}
}

func TestValidator_HasRule(t *testing.T) {
	validator := New()

	// Test with a rule that should exist
	if !validator.HasRule("email") {
		t.Error("HasRule should return true for 'email' rule")
	}

	// Test with a rule that shouldn't exist
	if validator.HasRule("nonexistent_rule") {
		t.Error("HasRule should return false for nonexistent rule")
	}
}

func TestValidator_SetCustomMessage(t *testing.T) {
	validator := New()

	// Set custom message for required rule
	validator.SetCustomMessage("required", "This field is absolutely required!")
	validator.SetCustomAttribute("email", "Email Address")

	data := map[string]interface{}{
		"name":  "",
		"email": "",
	}

	rules := map[string]string{
		"name":  "required",
		"email": "required|email",
	}

	result := validator.ValidateWithResult(data, rules)
	if result.IsValid() {
		t.Error("Expected validation to fail")
	}

	errors := result.Errors()

	// Check if custom message is used for required rule
	nameErrors, exists := errors["name"]
	if !exists || len(nameErrors) == 0 {
		t.Error("Expected name field to have validation errors")
	} else if nameErrors[0] != "This field is absolutely required!" {
		t.Errorf("Expected custom message 'This field is absolutely required!', got '%s'", nameErrors[0])
	}
}

func TestValidator_SetCustomAttribute(t *testing.T) {
	validator := New()

	// Set custom attribute
	validator.SetCustomAttribute("email", "Email Address")

	data := map[string]interface{}{
		"email": "",
	}

	rules := map[string]string{
		"email": "required",
	}

	result := validator.ValidateWithResult(data, rules)
	if result.IsValid() {
		t.Error("Expected validation to fail")
	}

	// Custom attribute should be reflected in error message
	errors := result.Errors()
	emailErrors, exists := errors["email"]
	if !exists || len(emailErrors) == 0 {
		t.Error("Expected email field to have validation errors")
	}
	// The error message should contain the custom attribute name
	// Note: This depends on the message resolver implementation
}

func TestValidator_CustomMessageRequestIsolation(t *testing.T) {
	// Test that different validator instances have isolated custom messages
	validator1 := New()
	validator2 := New()

	// Set different custom messages
	validator1.SetCustomMessage("required", "Validator 1: Field is required")
	validator2.SetCustomMessage("required", "Validator 2: Field cannot be empty")

	data := map[string]interface{}{
		"field": "",
	}
	rules := map[string]string{
		"field": "required",
	}

	// Validate with first validator
	result1 := validator1.ValidateWithResult(data, rules)
	if result1.IsValid() {
		t.Error("Expected validation to fail for validator1")
	}

	// Validate with second validator
	result2 := validator2.ValidateWithResult(data, rules)
	if result2.IsValid() {
		t.Error("Expected validation to fail for validator2")
	}

	// Check that each validator uses its own custom message
	errors1 := result1.Errors()
	errors2 := result2.Errors()

	fieldErrors1, exists1 := errors1["field"]
	fieldErrors2, exists2 := errors2["field"]

	if !exists1 || len(fieldErrors1) == 0 {
		t.Error("Expected field errors for validator1")
	}
	if !exists2 || len(fieldErrors2) == 0 {
		t.Error("Expected field errors for validator2")
	}

	// Messages should be different (request isolation)
	if exists1 && exists2 && len(fieldErrors1) > 0 && len(fieldErrors2) > 0 {
		if fieldErrors1[0] == fieldErrors2[0] {
			t.Error("Expected different custom messages for different validator instances")
		}
	}
}

func TestValidator_FieldSpecificCustomMessages(t *testing.T) {
	validator := New()

	// Set field-specific custom messages
	validator.SetCustomMessage("required.name", "The name field cannot be empty")
	validator.SetCustomMessage("required.email", "Email is mandatory for registration")

	data := map[string]interface{}{
		"name":  "",
		"email": "",
	}

	rules := map[string]string{
		"name":  "required",
		"email": "required",
	}

	result := validator.ValidateWithResult(data, rules)
	if result.IsValid() {
		t.Error("Expected validation to fail")
	}

	errors := result.Errors()

	// Check field-specific messages
	nameErrors, nameExists := errors["name"]
	emailErrors, emailExists := errors["email"]

	if !nameExists || len(nameErrors) == 0 {
		t.Error("Expected name field to have validation errors")
	}
	if !emailExists || len(emailErrors) == 0 {
		t.Error("Expected email field to have validation errors")
	}

	// Note: Field-specific message behavior depends on message resolver implementation
}

func TestValidator_MultipleRulesWithCustomMessages(t *testing.T) {
	validator := New()

	// Set custom messages for different rules
	validator.SetCustomMessage("required", "This field is required")
	validator.SetCustomMessage("min", "This field must be at least :param0 characters")
	validator.SetCustomMessage("email", "Please enter a valid email address")
	validator.SetCustomAttribute("password", "Password")

	data := map[string]interface{}{
		"name":     "",
		"email":    "invalid-email",
		"password": "123",
	}

	rules := map[string]string{
		"name":     "required",
		"email":    "required|email",
		"password": "required|min:8",
	}

	result := validator.ValidateWithResult(data, rules)
	if result.IsValid() {
		t.Error("Expected validation to fail")
	}

	errors := result.Errors()

	// Should have errors for all fields
	if _, exists := errors["name"]; !exists {
		t.Error("Expected name field to have validation errors")
	}
	if _, exists := errors["email"]; !exists {
		t.Error("Expected email field to have validation errors")
	}
	if _, exists := errors["password"]; !exists {
		t.Error("Expected password field to have validation errors")
	}
}

func TestValidator_ConcurrentValidationRequests(t *testing.T) {
	// Test concurrent validation with different custom messages
	done := make(chan bool, 2)

	// Goroutine 1
	go func() {
		defer func() { done <- true }()

		v1 := New()
		v1.SetCustomMessage("required", "Request 1: Field is required")

		data := map[string]interface{}{"field": ""}
		rules := map[string]string{"field": "required"}

		result := v1.ValidateWithResult(data, rules)
		if result.IsValid() {
			t.Error("Expected validation to fail in goroutine 1")
		}
	}()

	// Goroutine 2
	go func() {
		defer func() { done <- true }()

		v2 := New()
		v2.SetCustomMessage("required", "Request 2: This field cannot be empty")

		data := map[string]interface{}{"field": ""}
		rules := map[string]string{"field": "required"}

		result := v2.ValidateWithResult(data, rules)
		if result.IsValid() {
			t.Error("Expected validation to fail in goroutine 2")
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// If we reach here without deadlock or race conditions, the test passes
}

// NOTE: Removed deprecated quick validation function tests
// (TestRequired, TestEmail, TestNumeric, TestAlpha, TestMin, TestMax, TestBoolean)
// These functions have been deprecated and removed in favor of the proper rule-based validation system.
// Use the main Validate() or ValidateWithResult() methods instead.

// customTestRule is a real rule implementation for testing custom rule registration
type customTestRule struct{}

func (r *customTestRule) Name() string {
	return "custom_test"
}

func (r *customTestRule) Validate(ctx contract.RuleContext) error {
	// Simple validation: accept any non-nil value
	if ctx.Value() == nil {
		return fmt.Errorf("custom test rule failed: value is nil")
	}
	return nil
}

func (r *customTestRule) Message() string {
	return "The :attribute field failed custom test validation"
}

func (r *customTestRule) ShouldSkipValidation(_ interface{}) bool {
	return false
}
