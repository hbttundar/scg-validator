package conditional_test

import (
	"testing"

	"github.com/hbttundar/scg-validator/contract"
	"github.com/hbttundar/scg-validator/rules/conditional"
)

func TestRequiredRule(t *testing.T) {
	rule, _ := conditional.NewRequiredRule()

	tests := []struct {
		name       string
		value      any
		shouldPass bool
	}{
		{"non-empty string", "hello", true},
		{"empty string", "", false},
		{"nil", nil, false},
		{"non-zero int", 42, true},
		{"zero int", 0, false},
		{"true bool", true, true},
		{"false bool", false, false},
		{"non-empty slice", []int{1, 2}, true},
		{"empty slice", []int{}, false},
		{"non-empty map", map[string]int{"a": 1}, true},
		{"empty map", map[string]int{}, false},
		{"pointer to value", ptrToInt(5), true},
		{"pointer to zero", ptrToInt(0), false},
		{"pointer to nil", (*int)(nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := contract.NewValidationContext("field", tt.value, nil, nil)
			err := rule.Validate(ctx)
			if tt.shouldPass && err != nil {
				t.Errorf("expected pass, got error: %v", err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("expected failure, but passed for value: %#v", tt.value)
			}
		})
	}
}

// helper for pointer tests
func ptrToInt(v int) *int {
	return &v
}
