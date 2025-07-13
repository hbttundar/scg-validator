package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Sentinel error to indicate validation failure within a rule function.
var errValidationFailed = errors.New("validation failed")

// --- INTERFACES AND CORE TYPES ---

// PresenceVerifier defines the contract for checking if a value exists or is unique.
// This is the "Port" for our database "Adapter".
type PresenceVerifier interface {
	Exists(table, column string, value any, excludeIDColumn, excludeIDValue string) (bool, error)
}

// ValidationFunc defines the signature for a validation rule's logic.
// It returns a generic error on failure, which the engine then uses to apply a template.
type ValidationFunc func(ctx *ValidationContext) error

// RuleDefinition bundles the validation logic with its error message template.
type RuleDefinition struct {
	Function    ValidationFunc
	MessageTmpl string
}

// ValidationContext provides the validation function with all necessary data and dependencies.
type ValidationContext struct {
	Field            string
	Value            any
	Params           []string
	Data             map[string]any
	Rule             string           // Current rule being validated
	PresenceVerifier PresenceVerifier // <-- Dependency Injected
}

// Validator is the main engine, holding registered rules and the presence verifier.
type Validator struct {
	rules           map[string]RuleDefinition
	verifier        PresenceVerifier
	customMessages  map[string]string // Custom error messages
	customAttribute map[string]string // Custom attribute names
}

// --- CONFIGURATION AND EXTENSION (FUNCTIONAL OPTIONS) ---

// Option is a function that configures a Validator instance.
type Option func(v *Validator)

// WithRule is a functional option for passing a temporary, one-off validation rule.
// nolint:unused
func WithRule(name, messageTmpl string, fn ValidationFunc) Option {
	return func(v *Validator) {
		// This option adds a temporary rule definition to the validator instance
		// that only exists for the lifetime of the Validate call where it's used.
		// NOTE: In the `Validate` function, these will be handled specially.
		// For this example's simplicity, we are showing how it would be structured.
		// A full implementation would merge these into a temporary map.
		v.rules[name] = RuleDefinition{Function: fn, MessageTmpl: messageTmpl}
	}
}

// WithPresenceVerifier injects a database verifier into the Validator instance.
func WithPresenceVerifier(verifier PresenceVerifier) Option {
	return func(v *Validator) {
		v.verifier = verifier
	}
}

// WithCustomMessages sets custom error messages for specific fields and rules.
// The key format can be:
// - "field.rule" for field-specific rule messages
// - "rule" for global rule messages
func WithCustomMessages(messages map[string]string) Option {
	return func(v *Validator) {
		for key, message := range messages {
			v.customMessages[key] = message
		}
	}
}

// WithCustomAttributes sets custom attribute names for fields.
func WithCustomAttributes(attributes map[string]string) Option {
	return func(v *Validator) {
		for field, attribute := range attributes {
			v.customAttribute[field] = attribute
		}
	}
}

// --- CONSTRUCTOR AND MAIN METHODS ---

// New creates a new Validator instance, optionally configured with functional options.
func New(options ...Option) *Validator {
	validator := &Validator{
		rules:           make(map[string]RuleDefinition),
		customMessages:  make(map[string]string),
		customAttribute: make(map[string]string),
	}
	validator.registerDefaultRules()

	// Apply all provided options
	for _, option := range options {
		option(validator)
	}

	return validator
}

// Validate is the primary method for validating data against a set of rules.
func (v *Validator) Validate(data any, rules map[string]string) error {
	dataMap, err := v.dataToMap(data)
	if err != nil {
		return fmt.Errorf("validator: %w", err)
	}

	valErrors := make(Errors)
	for field, ruleString := range rules {
		parsedRules := ParseRules(ruleString)
		if parsedRules == nil {
			continue
		}

		fieldValue := dataMap[field]

		for _, rule := range parsedRules {
			ruleDef, ok := v.rules[rule.Name]
			if !ok {
				valErrors.Add(field, fmt.Sprintf("validation rule '%s' is not registered", rule.Name))
				continue
			}

			context := &ValidationContext{
				Field:            field,
				Value:            fieldValue,
				Params:           rule.Params,
				Data:             dataMap,
				Rule:             rule.Name,
				PresenceVerifier: v.verifier, // Pass verifier to the context
			}

			if err := ruleDef.Function(context); err != nil {
				// Check if it's a specific error message or just a validation failure
				if err != errValidationFailed {
					// Use the specific error message
					valErrors.Add(field, err.Error())
				} else {
					// The rule failed. Now format the message.
					message := v.formatErrorMessage(ruleDef.MessageTmpl, context)
					valErrors.Add(field, message)
				}
			}
		}
	}

	if len(valErrors) > 0 {
		return valErrors
	}

	return nil
}

// --- HELPERS AND PRIVATE METHODS ---

func (v *Validator) RegisterRule(name, messageTmpl string, fn ValidationFunc) {
	v.rules[name] = RuleDefinition{Function: fn, MessageTmpl: messageTmpl}
}

func (v *Validator) registerDefaultRules() {
	// Presence & Type
	v.RegisterRule("required", "the :attribute field is required", validateRequired)
	v.RegisterRule("boolean", "the :attribute field must be true or false", validateBoolean)
	v.RegisterRule("numeric", "the :attribute field must be a number", validateNumeric)
	v.RegisterRule("array", "the :attribute field must be an array or map", validateArray)

	// Format
	v.RegisterRule("email", "the :attribute field must be a valid email address", validateEmail)
	v.RegisterRule("uuid", "the :attribute field must be a valid UUID", validateUUID)
	v.RegisterRule("alpha", "the :attribute field must only contain letters", validateAlpha)
	v.RegisterRule("alphanum", "the :attribute field must only contain letters and numbers", validateAlphanum)
	v.RegisterRule("date", "the :attribute field must be a valid date", validateDate)
	v.RegisterRule("url", "the :attribute field must be a valid URL", validateURL)
	v.RegisterRule("ip", "the :attribute field must be a valid IP address", validateIP)
	v.RegisterRule("regex", "the :attribute field format is invalid", validateRegex)

	// Size
	v.RegisterRule("size", "the :attribute field must be :param0", validateSize)
	v.RegisterRule("between", "the :attribute field must be between :param0 and :param1", validateBetween)
	v.RegisterRule("min", "the :attribute field must be at least :param0", validateMin)
	v.RegisterRule("max", "the :attribute field must not be greater than :param0", validateMax)
	v.RegisterRule("gt", "the :attribute field must be greater than :param0", validateGreaterThan)
	v.RegisterRule("lt", "the :attribute field must be less than :param0", validateLessThan)

	// Inclusion & Comparison
	v.RegisterRule("in", "the selected :attribute is invalid", validateIn)
	v.RegisterRule("not_in", "the selected :attribute is invalid", validateNotIn)
	v.RegisterRule("confirmed", "the :attribute confirmation does not match", validateConfirmed)

	// Conditional Validation
	v.RegisterRule("required_if", "the :attribute field is required when :param0 is :param1", validateRequiredIf)
	v.RegisterRule("required_unless", "the :attribute field is required unless :param0 is :param1", validateRequiredUnless)
	v.RegisterRule("required_with", "the :attribute field is required when :param0 is present", validateRequiredWith)
	v.RegisterRule(
		"required_without",
		"the :attribute field is required when :param0 is not present",
		validateRequiredWithout,
	)

	// Database
	v.RegisterRule("unique", "the :attribute has already been taken", validateUnique)

	// Additional Laravel-inspired rules
	v.RegisterRule("accepted", "the :attribute field must be accepted", validateAccepted)
	v.RegisterRule("same", "the :attribute field must match :param0", validateSame)
	v.RegisterRule("different", "the :attribute field must be different from :param0", validateDifferent)
	v.RegisterRule("starts_with", "the :attribute field must start with one of the following: :param0", validateStartsWith)
	v.RegisterRule("ends_with", "the :attribute field must end with one of the following: :param0", validateEndsWith)
	v.RegisterRule("integer", "the :attribute field must be an integer", validateInteger)
	v.RegisterRule("string", "the :attribute field must be a string", validateString)
	v.RegisterRule("json", "the :attribute field must be a valid JSON string", validateJSON)
	v.RegisterRule("nullable", "the :attribute field is optional", validateNullable)

	// File validation rules
	v.RegisterRule("file", "the :attribute field must be a file", validateFile)
	v.RegisterRule("image", "the :attribute field must be an image", validateImage)
	v.RegisterRule("mimes", "the :attribute field must be a file of type: :param0", validateMimes)

	// Date validation rules
	v.RegisterRule("after", "the :attribute field must be a date after :param0", validateAfter)
	v.RegisterRule("before", "the :attribute field must be a date before :param0", validateBefore)

	// Password validation rule
	v.RegisterRule("password", "the :attribute field must meet the password requirements", validatePassword)
}

// formatErrorMessage replaces placeholders in the message template with context values.
func (v *Validator) formatErrorMessage(template string, ctx *ValidationContext) string {
	// Check for custom message for this specific field and rule
	fieldRuleKey := fmt.Sprintf("%s.%s", ctx.Field, ctx.Rule)
	if customMsg, exists := v.customMessages[fieldRuleKey]; exists {
		template = customMsg
	} else if customMsg, exists := v.customMessages[ctx.Rule]; exists {
		// Check for global rule message
		template = customMsg
	}

	// Get attribute name (use custom if available)
	attributeName := ctx.Field
	if customAttr, exists := v.customAttribute[ctx.Field]; exists {
		attributeName = customAttr
	}

	// Replace placeholders
	message := strings.ReplaceAll(template, ":attribute", attributeName)
	message = strings.ReplaceAll(message, ":value", fmt.Sprintf("%v", ctx.Value))

	for i, param := range ctx.Params {
		placeholder := fmt.Sprintf(":param%d", i)
		message = strings.ReplaceAll(message, placeholder, param)
	}

	return message
}

func (v *Validator) dataToMap(data any) (map[string]any, error) {
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	kind := value.Kind()

	if kind == reflect.Map {
		if dataMap, ok := data.(map[string]any); ok {
			return dataMap, nil
		}
		return nil, fmt.Errorf("unsupported map type")
	}
	if kind != reflect.Struct {
		return nil, fmt.Errorf("unsupported data type '%s'", kind)
	}
	dataMap := make(map[string]any, value.NumField())
	valueType := value.Type()
	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Field(i)
		if !field.IsExported() {
			continue
		}
		key := field.Tag.Get("validate")
		if key == "" {
			key = strings.Split(field.Tag.Get("json"), ",")[0]
		}
		if key == "" || key == "-" {
			key = field.Name
		}
		dataMap[key] = value.Field(i).Interface()
	}
	return dataMap, nil
}
