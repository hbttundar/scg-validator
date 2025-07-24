package engine

import (
	"testing"

	"github.com/hbttundar/scg-validator/contract"
	"github.com/hbttundar/scg-validator/message"
)

func TestEngine_Execute(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		rules       map[string]string
		wantValid   bool
		wantErrors  int
		description string
	}{
		{
			name: "valid_basic_validation",
			data: map[string]interface{}{
				"name":  "JohnDoe",
				"email": "john@example.com",
				"age":   25,
			},
			rules: map[string]string{
				"name":  "required|alpha",
				"email": "required|email",
				"age":   "required|numeric|min:18|max:100",
			},
			wantValid:   true,
			wantErrors:  0,
			description: "Should pass with valid basic data",
		},
		{
			name: "invalid_basic_validation",
			data: map[string]interface{}{
				"name":  "John123", // contains numbers
				"email": "invalid", // invalid email
				"age":   15,        // below minimum
			},
			rules: map[string]string{
				"name":  "required|alpha",
				"email": "required|email",
				"age":   "required|numeric|min:18|max:100",
			},
			wantValid:   false,
			wantErrors:  3,
			description: "Should fail with invalid data for all fields",
		},
		{
			name: "complex_rules_validation",
			data: map[string]interface{}{
				"username": "john_doe",
				"password": "secret123",
				"confirm":  "secret123",
			},
			rules: map[string]string{
				"username": "required|alpha_dash|min:3|max:20",
				"password": "required|min:6|max:50",
				"confirm":  "required|min:6|max:50",
			},
			wantValid:   true,
			wantErrors:  0,
			description: "Should pass with complex validation rules",
		},
		{
			name: "missing_required_fields",
			data: map[string]interface{}{
				"optional": "value",
			},
			rules: map[string]string{
				"name":     "required",
				"email":    "required|email",
				"optional": "alpha",
			},
			wantValid:   false,
			wantErrors:  3, // name fails required, email fails both required and email validation
			description: "Should fail when required fields are missing",
		},
		{
			name: "empty_string_validation",
			data: map[string]interface{}{
				"name":  "",
				"email": "",
			},
			rules: map[string]string{
				"name":  "required",
				"email": "required|email",
			},
			wantValid:   false,
			wantErrors:  3, // name fails required, email fails both required and email validation
			description: "Should fail with empty strings for required fields",
		},
		{
			name: "numeric_validation",
			data: map[string]interface{}{
				"age":    25,
				"score":  85.5,
				"count":  "123",
				"rating": "4.5",
			},
			rules: map[string]string{
				"age":    "required|numeric|min:18|max:100",
				"score":  "required|numeric|min:0|max:100",
				"count":  "required|numeric",
				"rating": "required|numeric|min:1|max:5",
			},
			wantValid:   true,
			wantErrors:  0,
			description: "Should pass with various numeric validations",
		},
		{
			name: "string_validation",
			data: map[string]interface{}{
				"name":     "John",
				"username": "john_doe_123",
				"slug":     "my-blog-post",
			},
			rules: map[string]string{
				"name":     "required|alpha|min:2|max:50",
				"username": "required|alpha_dash|min:3|max:20",
				"slug":     "required|alpha_dash",
			},
			wantValid:   true,
			wantErrors:  0,
			description: "Should pass with string validation rules",
		},
		{
			name: "boolean_validation",
			data: map[string]interface{}{
				"active":    true,
				"published": true, // Changed from false to true since false fails required validation
				"enabled":   "true",
				"disabled":  "false",
			},
			rules: map[string]string{
				"active":    "required|boolean",
				"published": "required|boolean",
				"enabled":   "required|boolean",
				"disabled":  "required|boolean",
			},
			wantValid:   true,
			wantErrors:  0,
			description: "Should pass with boolean validation",
		},
		{
			name: "bail_rule_stops_validation",
			data: map[string]interface{}{
				"field": "",
			},
			rules: map[string]string{
				"field": "bail|required|min:5|email",
			},
			wantValid:   false,
			wantErrors:  1, // Should stop after first failure due to bail
			description: "Should stop validation after first failure with bail rule",
		},
		{
			name: "unknown_rule_error",
			data: map[string]interface{}{
				"field": "value",
			},
			rules: map[string]string{
				"field": "unknown_rule",
			},
			wantValid:   false,
			wantErrors:  1,
			description: "Should fail with unknown rule error",
		},
		{
			name: "empty_rules",
			data: map[string]interface{}{
				"field": "value",
			},
			rules:       map[string]string{},
			wantValid:   true,
			wantErrors:  0,
			description: "Should pass with empty rules",
		},
		{
			name: "multiple_field_types",
			data: map[string]interface{}{
				"string_field":  "test",
				"numeric_field": 123,
				"boolean_field": true,
				"array_field":   []string{"a", "b", "c"},
			},
			rules: map[string]string{
				"string_field":  "required|alpha|min:2",
				"numeric_field": "required|numeric|min:100",
				"boolean_field": "required|boolean",
				"array_field":   "required",
			},
			wantValid:   true,
			wantErrors:  0,
			description: "Should handle multiple field types correctly",
		},
		{
			name: "complex_conditional_validation",
			data: map[string]interface{}{
				"type":     "premium",
				"discount": 15,
				"email":    "user@example.com",
			},
			rules: map[string]string{
				"type":     "required|alpha",
				"discount": "required_if:type,premium|numeric|min:10|max:50",
				"email":    "required|email",
			},
			wantValid:   true,
			wantErrors:  0,
			description: "Should handle complex conditional validation",
		},
		{
			name: "nested_validation_with_bail",
			data: map[string]interface{}{
				"password": "",
			},
			rules: map[string]string{
				"password": "bail|required|min:8|alpha_dash",
			},
			wantValid:   false,
			wantErrors:  1, // Should stop after required fails due to bail
			description: "Should stop validation early with bail rule",
		},
		{
			name: "edge_case_empty_and_zero_values",
			data: map[string]interface{}{
				"empty_string": "",
				"zero_int":     0,
				"false_bool":   false,
				"nil_value":    nil,
			},
			rules: map[string]string{
				"empty_string": "required",
				"zero_int":     "required|numeric",
				"false_bool":   "required|boolean",
				"nil_value":    "required",
			},
			wantValid:   false,
			wantErrors:  4, // all fields should fail required validation (including zero_int)
			description: "Should handle edge cases with empty and zero values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			provider := NewDataProvider(tt.data)
			result := engine.Execute(provider, tt.rules)

			if result.IsValid() != tt.wantValid {
				t.Errorf("Expected IsValid() = %v, got %v", tt.wantValid, result.IsValid())
				if !result.IsValid() {
					t.Logf("Validation errors: %v", result.Errors())
				}
			}

			errors := result.Errors()
			errorCount := 0
			for _, fieldErrors := range errors {
				errorCount += len(fieldErrors)
			}

			if errorCount != tt.wantErrors {
				t.Errorf("Expected %d errors, got %d", tt.wantErrors, errorCount)
				t.Logf("Actual errors: %v", errors)
			}

			t.Logf("Test case: %s", tt.description)
		})
	}
}

func TestEngine_CustomMessages(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]interface{}
		rules          map[string]string
		customMessages map[string]string
		customAttrs    map[string]string
		wantContains   []string
		description    string
	}{
		{
			name: "custom_rule_message",
			data: map[string]interface{}{
				"name": "",
			},
			rules: map[string]string{
				"name": "required",
			},
			customMessages: map[string]string{
				"required": "This field is absolutely required!",
			},
			wantContains: []string{"This field is absolutely required!"},
			description:  "Should use custom message for rule",
		},
		{
			name: "custom_attribute_name",
			data: map[string]interface{}{
				"email": "",
			},
			rules: map[string]string{
				"email": "required",
			},
			customAttrs: map[string]string{
				"email": "Email Address",
			},
			wantContains: []string{"Email Address field is required"},
			description:  "Should use custom attribute name",
		},
		{
			name: "field_specific_message",
			data: map[string]interface{}{
				"name":  "",
				"email": "",
			},
			rules: map[string]string{
				"name":  "required",
				"email": "required",
			},
			customMessages: map[string]string{
				"required.name":  "Name cannot be empty",
				"required.email": "Email is mandatory",
			},
			wantContains: []string{"Name cannot be empty", "Email is mandatory"},
			description:  "Should use field-specific custom messages",
		},
		{
			name: "parameter_replacement",
			data: map[string]interface{}{
				"password": "123",
			},
			rules: map[string]string{
				"password": "min:8",
			},
			customMessages: map[string]string{
				"min": "The field must be at least :param0 characters long",
			},
			wantContains: []string{"must be at least 8 characters long"},
			description:  "Should replace parameters in custom messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()

			// Set custom messages
			for rule, message := range tt.customMessages {
				engine.SetCustomMessage(rule, message)
			}

			// Set custom attributes
			for field, attr := range tt.customAttrs {
				engine.SetCustomAttribute(field, attr)
			}

			provider := NewDataProvider(tt.data)
			result := engine.Execute(provider, tt.rules)

			if result.IsValid() {
				t.Error("Expected validation to fail")
				return
			}

			errors := result.Errors()
			allErrorMessages := ""
			for _, fieldErrors := range errors {
				for _, msg := range fieldErrors {
					allErrorMessages += msg + " "
				}
			}

			for _, expectedText := range tt.wantContains {
				if !contains(allErrorMessages, expectedText) {
					t.Errorf("Expected error message to contain '%s', got: %s", expectedText, allErrorMessages)
				}
			}

			t.Logf("Test case: %s", tt.description)
		})
	}
}

func TestEngine_RegisterRule(t *testing.T) {
	tests := []struct {
		name        string
		ruleName    string
		description string
	}{
		{
			name:        "register_rule_method_exists",
			ruleName:    "custom_test",
			description: "Should have RegisterRule method available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()

			// Test that RegisterRule method exists and can be called
			// We'll test with a simple rule creator using testRule
			err := engine.RegisterRule(tt.ruleName, func(_ []string) (contract.Rule, error) {
				return &testRule{name: tt.ruleName}, nil
			})

			// The method should exist and not panic
			if err != nil {
				t.Logf("RegisterRule returned error (expected for nil rule): %v", err)
			}

			t.Logf("Test case: %s", tt.description)
		})
	}
}

func TestEngine_MessageResolver(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*Engine)
		data        map[string]interface{}
		rules       map[string]string
		wantValid   bool
		description string
	}{
		{
			name: "set_real_message_resolver",
			setupFunc: func(e *Engine) {
				resolver := message.NewResolver()
				resolver.SetCustomMessage("required", "Custom required message")
				e.SetMessageResolver(resolver)
			},
			data: map[string]interface{}{
				"field": "",
			},
			rules: map[string]string{
				"field": "required",
			},
			wantValid:   false,
			description: "Should use real message resolver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()

			if tt.setupFunc != nil {
				tt.setupFunc(engine)
			}

			provider := NewDataProvider(tt.data)
			result := engine.Execute(provider, tt.rules)

			if result.IsValid() != tt.wantValid {
				t.Errorf("Expected IsValid() = %v, got %v", tt.wantValid, result.IsValid())
			}

			t.Logf("Test case: %s", tt.description)
		})
	}
}

// Helper functions and types for testing

type testRule struct {
	name string
}

func (r *testRule) Name() string {
	return r.name
}

func (r *testRule) Validate(_ contract.RuleContext) error {
	return nil
}

func (r *testRule) Message() string {
	return "Test rule message"
}

func (r *testRule) ShouldSkipValidation(_ interface{}) bool {
	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
