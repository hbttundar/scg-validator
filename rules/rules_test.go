// Package rules provides tests for the main rules package
package rules

import (
	"testing"

	"github.com/hbttundar/scg-validator/contract"
)

func TestNewRuleRegistry(t *testing.T) {
	t.Run("DefaultRegistry", func(t *testing.T) {
		reg := NewRuleRegistry()
		if reg == nil {
			t.Fatal("NewRuleRegistry returned nil")
		}
		if !reg.Has(RuleRequiredIf) {
			t.Errorf("Expected registry to have rule '%s'", RuleRequiredIf)
		}
		if !reg.Has(RuleAlpha) {
			t.Errorf("Expected registry to have rule '%s'", RuleAlpha)
		}
	})

	t.Run("WithExcludeRules", func(t *testing.T) {
		reg := NewRuleRegistry(WithExcludeRules(RuleEmail, RuleURL))
		if reg.Has(RuleEmail) {
			t.Error("Expected registry to not have excluded rule 'email'")
		}
		if reg.Has(RuleURL) {
			t.Error("Expected registry to not have excluded rule 'url'")
		}
		if !reg.Has(RuleMin) {
			t.Error("Expected registry to still have non-excluded rule 'min'")
		}
	})

	t.Run("WithIncludeOnly", func(t *testing.T) {
		reg := NewRuleRegistry(WithIncludeOnly(RuleMin, RuleMax))
		if !reg.Has(RuleMin) {
			t.Error("Expected registry to have included rule 'min'")
		}
		if !reg.Has(RuleMax) {
			t.Error("Expected registry to have included rule 'max'")
		}
		if reg.Has(RuleEmail) {
			t.Error("Expected registry to not have non-included rule 'email'")
		}
	})

	t.Run("WithCustomRule", func(t *testing.T) {
		customRuleName := "my_custom_rule"
		customCreator := func(_ []string) (contract.Rule, error) {
			return &mockRule{}, nil
		}
		reg := NewRuleRegistry(WithCustomRule(customRuleName, customCreator))
		if !reg.Has(customRuleName) {
			t.Errorf("Expected registry to have custom rule '%s'", customRuleName)
		}
	})

	t.Run("WithCustomMessage", func(t *testing.T) {
		// Custom messages are handled by the MessageResolver, not the Registry
		// This test just verifies that the option doesn't cause panics
		reg := NewRuleRegistry(WithCustomMessage(RuleMin, "custom message"))
		if reg == nil {
			t.Error("Expected registry to be created successfully with custom message option")
		}
		// Verify the registry still has the basic rule
		if !reg.Has(RuleMin) {
			t.Error("Expected registry to still have the min rule")
		}
	})
}

// mockRule is a simple rule for testing custom rule registration.
type mockRule struct{}

func (m *mockRule) Validate(_ contract.RuleContext) error {
	return nil
}

func (m *mockRule) Message() string {
	return "mock rule message"
}

func (m *mockRule) Name() string {
	return "mock_rule"
}

func (m *mockRule) ShouldSkipValidation(_ interface{}) bool {
	return false
}
