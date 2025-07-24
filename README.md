# SCG Validator

A comprehensive, extensible validation library for Go applications inspired by Laravel's validation system. SCG Validator provides a rich set of validation rules with support for custom rules, conditional validation, and flexible error handling.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Available Validation Rules](#available-validation-rules)
- [Usage Examples](#usage-examples)
- [Custom Rules](#custom-rules)
- [Custom Error Messages](#custom-error-messages)
- [Database Validation](#database-validation)
- [Development Tools](#development-tools)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Rich Rule Set**: 50+ built-in validation rules covering common use cases
- **Custom Rules**: Easy to add custom validation logic
- **Conditional Validation**: Rules that depend on other field values
- **Flexible Error Messages**: Customizable error messages with placeholders
- **Database Integration**: Built-in support for unique/exists validation
- **Type Safety**: Strong typing with interface-based design
- **Performance**: Optimized for high-performance applications

## Installation

```bash
go get github.com/hbttundar/scg-validator
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/hbttundar/scg-validator/validator"
    "github.com/hbttundar/scg-validator/contract"
)

func main() {
    // Create a validator instance
    v := validator.New()
    
    // Data to validate
    data := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   25,
    }
    
    // Validation rules
    rules := map[string]string{
        "name":  "required|string|min:2|max:50",
        "email": "required|email",
        "age":   "required|integer|min:18|max:120",
    }
    
    // Perform validation
    result := v.Validate(contract.NewSimpleDataProvider(data), rules)
    
    if !result.IsValid() {
        fmt.Println("Validation failed:")
        for field, errors := range result.Errors() {
            for _, err := range errors {
                fmt.Printf("- %s: %s\n", field, err)
            }
        }
        return
    }
    
    fmt.Println("Validation passed!")
}
```

## Available Validation Rules

### Basic Rules
- `required` - Field must be present and not empty
- `nullable` - Field can be null/empty
- `present` - Field must be present in input
- `filled` - Field must not be empty if present
- `sometimes` - Apply validation only if field is present

### String Rules
- `string` - Must be a string
- `alpha` - Only alphabetic characters
- `alphanum` - Only alphanumeric characters
- `alpha_dash` - Only alphanumeric, dashes, and underscores
- `ascii` - Only ASCII characters
- `lowercase` - Must be lowercase
- `uppercase` - Must be uppercase
- `slug` - Must be a valid slug

### Numeric Rules
- `numeric` - Must be numeric
- `integer` - Must be an integer
- `decimal:min,max` - Must have specific decimal places
- `multiple_of:value` - Must be a multiple of value

### Comparison Rules
- `min:value` - Minimum value/length
- `max:value` - Maximum value/length
- `size:value` - Exact size/length
- `between:min,max` - Between min and max values
- `gt:value` - Greater than value
- `gte:value` - Greater than or equal to value
- `lt:value` - Less than value
- `lte:value` - Less than or equal to value

### Format Rules
- `email` - Valid email address
- `url` - Valid URL
- `uuid` - Valid UUID
- `ip` - Valid IP address (IPv4, IPv6, or MAC)
- `ipv4` - Valid IPv4 address
- `ipv6` - Valid IPv6 address
- `mac` - Valid MAC address
- `json` - Valid JSON string
- `regex:pattern` - Matches regex pattern

### Date Rules
- `date` - Valid date
- `after:date` - After specified date
- `before:date` - Before specified date
- `after_or_equal:date` - After or equal to date
- `before_or_equal:date` - Before or equal to date
- `date_equals:date` - Equal to specified date

### Collection Rules
- `list` - Must be a list/array
- `map` - Must be a map/object
- `in:value1,value2` - Must be one of specified values
- `not_in:value1,value2` - Must not be one of specified values

### File Rules
- `file` - Must be a file upload
- `image` - Must be an image file
- `mimes:ext1,ext2` - Must have specified MIME types

### Conditional Rules
- `required_if:field,value` - Required if another field equals value
- `required_unless:field,value` - Required unless another field equals value
- `required_with:field` - Required if another field is present
- `required_without:field` - Required if another field is absent
- `required_with_all:field1,field2` - Required if all fields are present
- `required_without_all:field1,field2` - Required if all fields are absent

### Database Rules
- `exists:table,column` - Value must exist in database
- `unique:table,column` - Value must be unique in database

## Usage Examples

### Basic Validation

```go
v := validator.New()

data := map[string]interface{}{
    "username": "john_doe",
    "password": "secret123",
    "confirm_password": "secret123",
}

rules := map[string]string{
    "username": "required|alpha_dash|min:3|max:20",
    "password": "required|min:8",
    "confirm_password": "required|same:password",
}

result := v.Validate(contract.NewSimpleDataProvider(data), rules)
```

### Conditional Validation

```go
data := map[string]interface{}{
    "user_type": "admin",
    "admin_code": "12345",
    "regular_field": "value",
}

rules := map[string]string{
    "user_type": "required|in:admin,user",
    "admin_code": "required_if:user_type,admin|numeric",
    "regular_field": "required_unless:user_type,admin",
}
```

### Array and Nested Data Validation

```go
data := map[string]interface{}{
    "users": []map[string]interface{}{
        {"name": "John", "email": "john@example.com"},
        {"name": "Jane", "email": "jane@example.com"},
    },
    "settings": map[string]interface{}{
        "theme": "dark",
        "notifications": true,
    },
}

rules := map[string]string{
    "users": "required|list",
    "users.*.name": "required|string|min:2",
    "users.*.email": "required|email",
    "settings.theme": "required|in:light,dark",
    "settings.notifications": "required|boolean",
}
```

## Custom Rules

Create custom validation rules by implementing the `contract.Rule` interface:

```go
type UppercaseRule struct {
    common.BaseRule
}

func NewUppercaseRule() (contract.Rule, error) {
    return &UppercaseRule{
        BaseRule: common.NewBaseRule("uppercase", "The :attribute must be uppercase", nil),
    }, nil
}

func (r *UppercaseRule) Validate(ctx contract.RuleContext) error {
    value, ok := ctx.Value().(string)
    if !ok {
        return fmt.Errorf("value must be a string")
    }
    
    if strings.ToUpper(value) != value {
        return fmt.Errorf("the :attribute must be uppercase")
    }
    
    return nil
}

func (r *UppercaseRule) Name() string {
    return "uppercase"
}

// Register the custom rule
v := validator.NewWithOptions(
    rules.WithCustomRule("uppercase", func(params []string) (contract.Rule, error) {
        return NewUppercaseRule()
    }),
)
```

## Custom Error Messages

### Global Message Override

```go
v := validator.NewWithOptions(
    rules.WithCustomMessage("required", "The :attribute field is mandatory"),
    rules.WithCustomMessage("email", "Please provide a valid email address"),
)
```

### Per-Validation Messages

```go
messages := map[string]string{
    "email.required": "Email is required",
    "email.email": "Email format is invalid",
    "password.min": "Password must be at least 8 characters",
}

result := v.ValidateWithMessages(data, rules, messages)
```

### Custom Field Names

```go
v.SetAttribute("email", "Email Address")
v.SetAttribute("password", "Password")
// Error messages will use "Email Address" instead of "email"
```

## Database Validation

For `exists` and `unique` rules, implement the `PresenceVerifier` interface:

```go
type DatabaseVerifier struct {
    db *sql.DB
}

func (d *DatabaseVerifier) Exists(table, field string, value interface{}) (bool, error) {
    query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", table, field)
    var count int
    err := d.db.QueryRow(query, value).Scan(&count)
    return count > 0, err
}

func (d *DatabaseVerifier) Unique(table, field string, value interface{}) (bool, error) {
    exists, err := d.Exists(table, field, value)
    return !exists, err
}

// Register the verifier
database.RegisterPresenceVerifier("users", &DatabaseVerifier{db: db})

// Use in validation
rules := map[string]string{
    "email": "required|email|unique:users,email",
    "category_id": "required|exists:categories,id",
}
```

## Development Tools

This project includes a comprehensive development helper script (`./scg`) with the following commands:

### Available Commands

```bash
# Build the project
./scg build

# Run tests (optionally for specific package)
./scg test [package]

# Run benchmarks
./scg bench [package]

# Run linter
./scg lint

# Run linter and fix issues
./scg lint-fix

# Format code (gofmt + goimports)
./scg format [file]

# Run security checks
./scg security

# Clean build cache
./scg clean

# Manage dependencies
./scg deps

# Generate coverage report
./scg coverage

# Generate documentation
./scg docs

# Run all CI checks
./scg ci

# Install development tools
./scg install-tools
```

### Examples

```bash
# Run tests for a specific package
./scg test ./rules/types/string

# Format a specific file
./scg format rules/types/string/alpha.go

# Run full CI pipeline locally
./scg ci
```

## Testing

Run the test suite:

```bash
# Run all tests
./scg test

# Run tests with coverage
./scg coverage

# Run benchmarks
./scg bench
```

The project maintains high test coverage with comprehensive unit tests for all validation rules.

## Contributing

We welcome contributions! Please follow these guidelines:

1. **Fork the repository** and create a feature branch
2. **Write tests** for new functionality
3. **Follow Go conventions** and run `./scg format`
4. **Run the full test suite** with `./scg ci`
5. **Update documentation** as needed
6. **Submit a pull request** with a clear description

### Development Setup

```bash
# Clone the repository
git clone https://github.com/hbttundar/scg-validator.git
cd scg-validator

# Install development tools
./scg install-tools

# Run tests to ensure everything works
./scg test

# Make your changes and run CI checks
./scg ci
```

### Code Style

- Follow standard Go formatting (`gofmt`, `goimports`)
- Write comprehensive tests for new features
- Document public APIs with clear examples
- Use meaningful variable and function names
- Keep functions focused and concise

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by Laravel's validation system
- Built with Go's strong typing and interface system
- Designed for high-performance applications

---

For more examples and advanced usage, see the [example](example/) directory.
