package validator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Helper functions

// Safely gets the numeric or len() value of a field.
func getNumericOrLen(value reflect.Value) (float64, bool) {
	switch value.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return float64(value.Len()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(value.Uint()), true
	case reflect.Float32, reflect.Float64:
		return value.Float(), true
	default:
		return 0, false
	}
}

// All validate functions now simply return `errValidationFailed` on failure.
// They no longer format the error message string.

func validateRequired(ctx *ValidationContext) error {
	value := reflect.ValueOf(ctx.Value)
	if !value.IsValid() || (value.Kind() == reflect.Ptr && value.IsNil()) || value.IsZero() {
		return errValidationFailed
	}
	return nil
}

func validateEmail(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}

	// More comprehensive email validation pattern
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(value) {
		return errValidationFailed
	}
	return nil
}

func validateUUID(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}

	// UUID regex pattern (8-4-4-4-12 format)
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidPattern.MatchString(strings.ToLower(value)) {
		return errValidationFailed
	}
	return nil
}

func validateAlpha(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}
	for _, char := range value {
		if !unicode.IsLetter(char) {
			return errValidationFailed
		}
	}
	return nil
}

func validateAlphanum(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}
	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			return errValidationFailed
		}
	}
	return nil
}

func validateNumeric(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return nil
	}
	value := reflect.ValueOf(ctx.Value)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return nil
	case reflect.String:
		if _, err := strconv.ParseFloat(value.String(), 64); err != nil {
			return errValidationFailed
		}
		return nil
	default:
		return errValidationFailed
	}
}

func validateBoolean(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return nil
	}
	value := reflect.ValueOf(ctx.Value)
	if value.Kind() == reflect.Bool {
		return nil
	}
	acceptable := map[string]bool{"true": true, "false": true, "1": true, "0": true}
	if !acceptable[strings.ToLower(fmt.Sprintf("%v", ctx.Value))] {
		return errValidationFailed
	}
	return nil
}

func validateMin(ctx *ValidationContext) error {
	if len(ctx.Params) != 1 {
		return fmt.Errorf("min rule requires one parameter")
	}
	minValue, _ := strconv.ParseFloat(ctx.Params[0], 64)
	val, ok := getNumericOrLen(reflect.ValueOf(ctx.Value))
	if !ok {
		return nil
	}
	if val < minValue {
		return errValidationFailed
	}
	return nil
}

func validateMax(ctx *ValidationContext) error {
	if len(ctx.Params) != 1 {
		return fmt.Errorf("max rule requires one parameter")
	}
	maxValue, _ := strconv.ParseFloat(ctx.Params[0], 64)
	val, ok := getNumericOrLen(reflect.ValueOf(ctx.Value))
	if !ok {
		return nil
	}
	if val > maxValue {
		return errValidationFailed
	}
	return nil
}

func validateGreaterThan(ctx *ValidationContext) error {
	if len(ctx.Params) != 1 {
		return fmt.Errorf("gt rule requires one parameter")
	}
	limit, _ := strconv.ParseFloat(ctx.Params[0], 64)
	val, ok := getNumericOrLen(reflect.ValueOf(ctx.Value))
	if !ok {
		return nil
	}
	if val <= limit {
		return errValidationFailed
	}
	return nil
}

func validateLessThan(ctx *ValidationContext) error {
	if len(ctx.Params) != 1 {
		return fmt.Errorf("lt rule requires one parameter")
	}
	limit, _ := strconv.ParseFloat(ctx.Params[0], 64)
	val, ok := getNumericOrLen(reflect.ValueOf(ctx.Value))
	if !ok {
		return nil
	}
	if val >= limit {
		return errValidationFailed
	}
	return nil
}

func validateIn(ctx *ValidationContext) error {
	valueStr := fmt.Sprintf("%v", ctx.Value)
	for _, param := range ctx.Params {
		if valueStr == param {
			return nil
		}
	}
	return errValidationFailed
}

func validateNotIn(ctx *ValidationContext) error {
	valueStr := fmt.Sprintf("%v", ctx.Value)
	for _, param := range ctx.Params {
		if valueStr == param {
			return errValidationFailed
		}
	}
	return nil
}

func validateConfirmed(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok {
		return nil
	}
	confirmationField := ctx.Field + "_confirmation"
	confirmationValue, exists := ctx.Data[confirmationField]
	if !exists || value != fmt.Sprintf("%v", confirmationValue) {
		return errValidationFailed
	}
	return nil
}

func validateUnique(ctx *ValidationContext) error {
	if ctx.PresenceVerifier == nil {
		return fmt.Errorf("cannot use 'unique' rule without a PresenceVerifier")
	}
	if len(ctx.Params) < 2 {
		return fmt.Errorf("unique rule requires at least 2 parameters (table, column)")
	}

	table := ctx.Params[0]
	column := ctx.Params[1]
	var excludeIDColumn, excludeIDValue string
	if len(ctx.Params) == 4 {
		excludeIDColumn = ctx.Params[2]
		// The value for the ID to exclude might be in the data map
		if id, ok := ctx.Data[ctx.Params[3]]; ok {
			excludeIDValue = fmt.Sprintf("%v", id)
		}
	}

	exists, err := ctx.PresenceVerifier.Exists(table, column, ctx.Value, excludeIDColumn, excludeIDValue)
	if err != nil {
		return err // Pass through database errors
	}
	if exists {
		return errValidationFailed
	}
	return nil
}

// Additional Laravel-inspired validation rules

func validateDate(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}

	// Simple date format validation (YYYY-MM-DD)
	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !datePattern.MatchString(value) {
		return errValidationFailed
	}

	// Parse the date to check if it's valid
	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		return errValidationFailed
	}

	return nil
}

func validateURL(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}

	// URL validation pattern
	urlPattern := regexp.MustCompile(`^(https?|ftp)://[^\s/$.?#].[^\s]*$`)
	if !urlPattern.MatchString(value) {
		return errValidationFailed
	}
	return nil
}

func validateIP(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}

	// IPv4 pattern
	octet := `(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`
	ipv4Pattern := regexp.MustCompile(`^` + octet + `(\.` + octet + `){3}$`)
	// IPv6 pattern (simplified)
	ipv6Pattern := regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)

	if !ipv4Pattern.MatchString(value) && !ipv6Pattern.MatchString(value) {
		return errValidationFailed
	}
	return nil
}

func validateBetween(ctx *ValidationContext) error {
	if len(ctx.Params) != 2 {
		return fmt.Errorf("between rule requires two parameters (min, max)")
	}

	minValue, err1 := strconv.ParseFloat(ctx.Params[0], 64)
	maxValue, err2 := strconv.ParseFloat(ctx.Params[1], 64)

	if err1 != nil || err2 != nil {
		return fmt.Errorf("between rule parameters must be numeric")
	}

	val, ok := getNumericOrLen(reflect.ValueOf(ctx.Value))
	if !ok {
		return nil
	}

	if val < minValue || val > maxValue {
		return errValidationFailed
	}
	return nil
}

func validateRegex(ctx *ValidationContext) error {
	if len(ctx.Params) != 1 {
		return fmt.Errorf("regex rule requires one parameter (pattern)")
	}

	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}

	pattern := ctx.Params[0]
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}

	if !regex.MatchString(value) {
		return errValidationFailed
	}
	return nil
}

func validateArray(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return nil
	}

	value := reflect.ValueOf(ctx.Value)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array && value.Kind() != reflect.Map {
		return errValidationFailed
	}
	return nil
}

func validateSize(ctx *ValidationContext) error {
	if len(ctx.Params) != 1 {
		return fmt.Errorf("size rule requires one parameter")
	}

	size, err := strconv.ParseFloat(ctx.Params[0], 64)
	if err != nil {
		return fmt.Errorf("size parameter must be numeric")
	}

	val, ok := getNumericOrLen(reflect.ValueOf(ctx.Value))
	if !ok {
		return nil
	}

	if val != size {
		return errValidationFailed
	}
	return nil
}

// Conditional validation rules

func validateRequiredIf(ctx *ValidationContext) error {
	if len(ctx.Params) < 2 {
		return fmt.Errorf("required_if rule requires at least 2 parameters (field, value)")
	}

	// Get the field and value to check against
	otherField := ctx.Params[0]
	otherValue := ctx.Params[1]

	// Check if the other field has the specified value
	if fieldValue, exists := ctx.Data[otherField]; exists {
		if fmt.Sprintf("%v", fieldValue) == otherValue {
			// If condition is met, apply the required rule
			return validateRequired(ctx)
		}
	}

	return nil
}

func validateRequiredUnless(ctx *ValidationContext) error {
	if len(ctx.Params) < 2 {
		return fmt.Errorf("required_unless rule requires at least 2 parameters (field, value)")
	}

	// Get the field and value to check against
	otherField := ctx.Params[0]
	otherValue := ctx.Params[1]

	// Check if the other field does NOT have the specified value
	if fieldValue, exists := ctx.Data[otherField]; exists {
		if fmt.Sprintf("%v", fieldValue) != otherValue {
			// If condition is met, apply the required rule
			return validateRequired(ctx)
		}
	} else {
		// If the other field doesn't exist, apply the required rule
		return validateRequired(ctx)
	}

	return nil
}

func validateRequiredWith(ctx *ValidationContext) error {
	if len(ctx.Params) < 1 {
		return fmt.Errorf("required_with rule requires at least 1 parameter (field)")
	}

	// Check if any of the specified fields are present and not empty
	for _, otherField := range ctx.Params {
		if fieldValue, exists := ctx.Data[otherField]; exists {
			// Check if the other field is not empty
			otherCtx := &ValidationContext{
				Field: otherField,
				Value: fieldValue,
			}
			if validateRequired(otherCtx) == nil {
				// If at least one field is present and not empty, apply the required rule
				return validateRequired(ctx)
			}
		}
	}

	return nil
}

func validateRequiredWithout(ctx *ValidationContext) error {
	if len(ctx.Params) < 1 {
		return fmt.Errorf("required_without rule requires at least 1 parameter (field)")
	}

	// Check if any of the specified fields are absent or empty
	for _, otherField := range ctx.Params {
		fieldValue, exists := ctx.Data[otherField]

		// Field is absent, apply the required rule
		if !exists {
			return validateRequired(ctx)
		}

		// Check if the other field is empty
		otherCtx := &ValidationContext{
			Field: otherField,
			Value: fieldValue,
		}
		if validateRequired(otherCtx) != nil {
			// Field is empty, apply the required rule
			return validateRequired(ctx)
		}
	}

	return nil
}

// Additional Laravel-inspired validation rules

func validateAccepted(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return errValidationFailed
	}

	// Check for common "accepted" values
	switch v := ctx.Value.(type) {
	case bool:
		if v {
			return nil
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if reflect.ValueOf(v).Int() == 1 {
			return nil
		}
	case float32, float64:
		if reflect.ValueOf(v).Float() == 1.0 {
			return nil
		}
	case string:
		lv := strings.ToLower(v)
		if lv == "yes" || lv == "on" || lv == "1" || lv == "true" {
			return nil
		}
	}

	return errValidationFailed
}

func validateSame(ctx *ValidationContext) error {
	if len(ctx.Params) != 1 {
		return fmt.Errorf("same rule requires one parameter (field)")
	}

	otherField := ctx.Params[0]
	otherValue, exists := ctx.Data[otherField]
	if !exists {
		return nil // Other field doesn't exist, can't compare
	}

	// Compare values as strings for simplicity
	if fmt.Sprintf("%v", ctx.Value) != fmt.Sprintf("%v", otherValue) {
		return errValidationFailed
	}

	return nil
}

func validateDifferent(ctx *ValidationContext) error {
	if len(ctx.Params) != 1 {
		return fmt.Errorf("different rule requires one parameter (field)")
	}

	otherField := ctx.Params[0]
	otherValue, exists := ctx.Data[otherField]
	if !exists {
		return nil // Other field doesn't exist, can't compare
	}

	// Compare values as strings for simplicity
	if fmt.Sprintf("%v", ctx.Value) == fmt.Sprintf("%v", otherValue) {
		return errValidationFailed
	}

	return nil
}

func validateStartsWith(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}

	if len(ctx.Params) == 0 {
		return fmt.Errorf("starts_with rule requires at least one parameter")
	}

	for _, prefix := range ctx.Params {
		if strings.HasPrefix(value, prefix) {
			return nil
		}
	}

	return errValidationFailed
}

func validateEndsWith(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}

	if len(ctx.Params) == 0 {
		return fmt.Errorf("ends_with rule requires at least one parameter")
	}

	for _, suffix := range ctx.Params {
		if strings.HasSuffix(value, suffix) {
			return nil
		}
	}

	return errValidationFailed
}

func validateInteger(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return nil
	}

	value := reflect.ValueOf(ctx.Value)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return nil
	case reflect.String:
		// Check if string can be parsed as an integer
		_, err := strconv.ParseInt(value.String(), 10, 64)
		if err != nil {
			return errValidationFailed
		}
		return nil
	default:
		return errValidationFailed
	}
}

func validateString(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return nil
	}

	value := reflect.ValueOf(ctx.Value)
	if value.Kind() != reflect.String {
		return errValidationFailed
	}

	return nil
}

func validateJSON(ctx *ValidationContext) error {
	value, ok := ctx.Value.(string)
	if !ok || value == "" {
		return nil
	}

	var js map[string]interface{}
	if err := json.Unmarshal([]byte(value), &js); err != nil {
		return errValidationFailed
	}

	return nil
}

func validateNullable(_ *ValidationContext) error {
	// This rule always passes - it's used to mark a field as optional
	// The actual validation happens in the Validate method
	return nil
}

// --- File Validation Rules ---

func validateFile(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return nil
	}

	// In a real implementation, this would check if the value is a file
	// For this example, we'll just check if it's a string (file path) or a map with file metadata
	value := reflect.ValueOf(ctx.Value)
	if value.Kind() == reflect.String {
		// Assume it's a file path
		return nil
	}

	if value.Kind() == reflect.Map {
		// Assume it's a file metadata map
		return nil
	}

	return errValidationFailed
}

func validateImage(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return nil
	}

	if err := validateFile(ctx); err != nil {
		return err
	}

	// In a real implementation, this would check if the file is an image
	// For this example, we'll just check if the string ends with an image extension
	value := reflect.ValueOf(ctx.Value)
	if value.Kind() == reflect.String {
		path := value.String()
		extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp"}
		for _, ext := range extensions {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				return nil
			}
		}
	}

	if value.Kind() == reflect.Map {
		// If it's a map, assume it has a "mime_type" field
		mimeTypeValue := value.MapIndex(reflect.ValueOf("mime_type"))
		if !mimeTypeValue.IsValid() {
			return errValidationFailed
		}

		mimeType, ok := mimeTypeValue.Interface().(string)
		if ok && strings.HasPrefix(mimeType, "image/") {
			return nil
		}
	}

	return errValidationFailed
}

func validateMimes(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return nil
	}

	if err := validateFile(ctx); err != nil {
		return err
	}

	if len(ctx.Params) == 0 {
		return fmt.Errorf("mimes rule requires at least one parameter")
	}

	// In a real implementation, this would check if the file's MIME type matches any of the specified types
	// For this example, we'll just check if the string ends with any of the specified extensions
	value := reflect.ValueOf(ctx.Value)
	if value.Kind() == reflect.String {
		path := strings.ToLower(value.String())
		for _, mime := range ctx.Params {
			ext := "." + mime
			if strings.HasSuffix(path, ext) {
				return nil
			}
		}
	}

	if value.Kind() == reflect.Map {
		// If it's a map, assume it has a "mime_type" field
		mimeTypeValue := value.MapIndex(reflect.ValueOf("mime_type"))
		if !mimeTypeValue.IsValid() {
			return errValidationFailed
		}

		mimeType, ok := mimeTypeValue.Interface().(string)
		if !ok {
			return errValidationFailed
		}

		for _, mime := range ctx.Params {
			if strings.HasPrefix(mimeType, mime+"/") || mimeType == mime {
				return nil
			}
		}
	}

	return errValidationFailed
}

// --- Date Validation Rules ---

// validateDateComparison is a helper function for date comparison validations
func validateDateComparison(ctx *ValidationContext, isAfter bool) error {
	if ctx.Value == nil {
		return nil
	}

	ruleName := "before"
	if isAfter {
		ruleName = "after"
	}

	if len(ctx.Params) != 1 {
		return fmt.Errorf("%s rule requires one parameter", ruleName)
	}

	// Parse the date from the value
	valueDate, err := parseDate(ctx.Value)
	if err != nil {
		return err
	}

	// Parse the date from the parameter
	var paramDate time.Time
	if ctx.Params[0] == "now" {
		paramDate = time.Now()
	} else if otherField, exists := ctx.Data[ctx.Params[0]]; exists {
		paramDate, err = parseDate(otherField)
		if err != nil {
			return err
		}
	} else {
		paramDate, err = parseDate(ctx.Params[0])
		if err != nil {
			return err
		}
	}

	// Check if the value date is after/before the parameter date
	if isAfter {
		if !valueDate.After(paramDate) {
			return errValidationFailed
		}
	} else {
		if !valueDate.Before(paramDate) {
			return errValidationFailed
		}
	}

	return nil
}

func validateAfter(ctx *ValidationContext) error {
	return validateDateComparison(ctx, true)
}

func validateBefore(ctx *ValidationContext) error {
	return validateDateComparison(ctx, false)
}

// parseDate is a helper function to parse a date from various formats
func parseDate(value any) (time.Time, error) {
	if value == nil {
		return time.Time{}, fmt.Errorf("cannot parse nil as date")
	}

	// If it's already a time.Time, return it
	if t, ok := value.(time.Time); ok {
		return t, nil
	}

	// If it's a string, try to parse it
	if s, ok := value.(string); ok {
		// Try common date formats
		formats := []string{
			time.RFC3339,
			time.RFC1123,
			time.ANSIC,
			"2006-01-02",
			"2006-01-02 15:04:05",
			"01/02/2006",
			"01/02/2006 15:04:05",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, s); err == nil {
				return t, nil
			}
		}

		return time.Time{}, fmt.Errorf("cannot parse %q as date", s)
	}

	// If it's a number, assume it's a Unix timestamp
	if n, ok := value.(int64); ok {
		return time.Unix(n, 0), nil
	}
	if n, ok := value.(int); ok {
		return time.Unix(int64(n), 0), nil
	}
	if n, ok := value.(float64); ok {
		return time.Unix(int64(n), 0), nil
	}

	return time.Time{}, fmt.Errorf("cannot parse %v as date", value)
}

// --- Password Validation Rules ---

// parsePasswordParams parses password validation parameters
// parseKeyValueParam handles key=value format parameters for password validation
func parseKeyValueParam(
	key, value string,
	minLength *int,
	requireUppercase, requireLowercase, requireNumeric, requireSpecial *bool,
) {
	switch key {
	case "min":
		if val, err := strconv.Atoi(value); err == nil {
			*minLength = val
		}
	case "uppercase":
		if val, err := strconv.ParseBool(value); err == nil {
			*requireUppercase = val
		}
	case "lowercase":
		if val, err := strconv.ParseBool(value); err == nil {
			*requireLowercase = val
		}
	case "numeric":
		if val, err := strconv.ParseBool(value); err == nil {
			*requireNumeric = val
		}
	case "special":
		if val, err := strconv.ParseBool(value); err == nil {
			*requireSpecial = val
		}
	}
}

// parsePasswordParams parses password validation parameters
func parsePasswordParams(
	params []string,
) (minLength int, requireUppercase, requireLowercase, requireNumeric, requireSpecial bool) {
	// Default requirements
	minLength = 8
	requireUppercase = false
	requireLowercase = false
	requireNumeric = false
	requireSpecial = false

	// Parse parameters if provided
	for _, param := range params {
		// Check if the parameter is a number for minLength
		if val, err := strconv.Atoi(param); err == nil {
			minLength = val
			continue
		}

		// Check if the parameter is a keyword
		switch param {
		case "uppercase":
			requireUppercase = true
		case "lowercase":
			requireLowercase = true
		case "numeric":
			requireNumeric = true
		case "special":
			requireSpecial = true
		default:
			// Handle key=value format for backward compatibility
			parts := strings.SplitN(param, "=", 2)
			if len(parts) == 2 {
				parseKeyValueParam(
					parts[0], parts[1],
					&minLength,
					&requireUppercase, &requireLowercase, &requireNumeric, &requireSpecial,
				)
			}
		}
	}

	return minLength, requireUppercase, requireLowercase, requireNumeric, requireSpecial
}

// checkPasswordRequirements checks if a password meets the specified requirements
func checkPasswordRequirements(
	password string,
	minLength int,
	requireUppercase, requireLowercase, requireNumeric, requireSpecial bool,
) error {
	// Check length
	if len(password) < minLength {
		return errValidationFailed
	}

	// Check character requirements
	charTypes := checkPasswordCharacterTypes(password)

	if requireUppercase && !charTypes.hasUpper {
		return errValidationFailed
	}
	if requireLowercase && !charTypes.hasLower {
		return errValidationFailed
	}
	if requireNumeric && !charTypes.hasNumeric {
		return errValidationFailed
	}
	if requireSpecial && !charTypes.hasSpecial {
		return errValidationFailed
	}

	return nil
}

// passwordCharTypes holds information about character types in a password
type passwordCharTypes struct {
	hasUpper   bool
	hasLower   bool
	hasNumeric bool
	hasSpecial bool
}

// checkPasswordCharacterTypes analyzes a password and returns which character types it contains
func checkPasswordCharacterTypes(password string) passwordCharTypes {
	var types passwordCharTypes

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			types.hasUpper = true
		case unicode.IsLower(char):
			types.hasLower = true
		case unicode.IsDigit(char):
			types.hasNumeric = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			types.hasSpecial = true
		}
	}

	return types
}

func validatePassword(ctx *ValidationContext) error {
	if ctx.Value == nil {
		return nil
	}

	// Get the password string
	value := reflect.ValueOf(ctx.Value)
	if value.Kind() != reflect.String {
		return nil // Allow non-string values
	}

	password := value.String()

	// Parse parameters and check requirements
	minLength, requireUppercase, requireLowercase, requireNumeric, requireSpecial := parsePasswordParams(ctx.Params)
	return checkPasswordRequirements(
		password,
		minLength,
		requireUppercase,
		requireLowercase,
		requireNumeric,
		requireSpecial,
	)
}
