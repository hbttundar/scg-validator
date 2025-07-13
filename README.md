# SCG Validator

[![Go Report Card](https://goreportcard.com/badge/github.com/hbttundar/scg-validator)](https://goreportcard.com/report/github.com/hbttundar/scg-validator)
[![GoDoc](https://godoc.org/github.com/hbttundar/scg-validator?status.svg)](https://godoc.org/github.com/hbttundar/scg-validator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful, flexible validation package for Go applications, inspired by Laravel's validation system. SCG Validator provides a simple, intuitive API for validating data in your Go applications with minimal boilerplate code.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Available Validation Rules](#available-validation-rules)
- [Custom Error Messages](#custom-error-messages)
- [Custom Attribute Names](#custom-attribute-names)
- [Database Validation](#database-validation)
- [Array and Map Validation](#array-and-map-validation)
- [Custom Validation Rules](#custom-validation-rules)
- [Advanced Usage](#advanced-usage)
- [Testing](#testing)
- [Contributing](#contributing)
- [Development](#development)
- [Versioning](#versioning)
- [License](#license)

## Features

- Laravel-style validation syntax
- Zero external dependencies
- Extensive set of validation rules (40+ built-in rules)
- Custom error messages with parameter substitution
- Custom attribute names
- Conditional validation
- Array and map validation
- Database presence verification
- Extensible architecture for custom rules

## Installation

```bash
go get github.com/hbttundar/scg-validator
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/hbttundar/scg-validator"
)

func main() {
    // Create a new validator
    v := validator.New()

    // Data to validate
    data := map[string]any{
        "username": "johndoe",
        "email": "john@example.com",
        "age": 25,
        "items": []string{"item1", "item2"},
    }

    // Define validation rules
    rules := map[string]string{
        "username": "required|alpha",
        "email": "required|email",
        "age": "required|numeric|gt:18",
        "items": "required|array",
    }

    // Validate the data
    err := v.Validate(data, rules)
    if err != nil {
        // Handle validation errors
        validationErrors := err.(validator.Errors)
        for field, messages := range validationErrors {
            for _, message := range messages {
                fmt.Printf("%s: %s\n", field, message)
            }
        }
        return
    }

    fmt.Println("Validation passed!")
}
```

## Examples

Here are some practical examples of how to use SCG Validator in different scenarios:

### Example 1: User Registration Form

```go
package main

import (
    "fmt"
    "github.com/hbttundar/scg-validator"
)

func main() {
    v := validator.New(
        validator.WithCustomMessages(map[string]string{
            "password.min": "Password must be at least :param0 characters long",
            "password.password": "Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character",
            "email.email": "Please provide a valid email address",
        }),
        validator.WithCustomAttributes(map[string]string{
            "password_confirmation": "Password Confirmation",
        }),
    )

    // User registration data
    data := map[string]any{
        "first_name": "John",
        "last_name": "Doe",
        "email": "john.doe@example.com",
        "password": "SecureP@ss123",
        "password_confirmation": "SecureP@ss123",
        "age": 25,
        "terms_accepted": true,
    }

    // Validation rules
    rules := map[string]string{
        "first_name": "required|alpha|min:2|max:50",
        "last_name": "required|alpha|min:2|max:50",
        "email": "required|email",
        "password": "required|min:8|password",
        "password_confirmation": "required|same:password",
        "age": "required|numeric|min:18",
        "terms_accepted": "required|accepted",
    }

    // Validate
    err := v.Validate(data, rules)
    if err != nil {
        validationErrors := err.(validator.Errors)
        for field, messages := range validationErrors {
            for _, message := range messages {
                fmt.Printf("%s: %s\n", field, message)
            }
        }
        return
    }

    fmt.Println("Registration successful!")
}
```

### Example 2: Product Creation with Conditional Validation

```go
package main

import (
    "fmt"
    "github.com/hbttundar/scg-validator"
)

func main() {
    v := validator.New()

    // Product data
    data := map[string]any{
        "name": "Smartphone",
        "category": "electronics",
        "price": 599.99,
        "stock": 100,
        "is_digital": false,
        "weight": 0.3, // in kg
        "dimensions": map[string]float64{
            "length": 15.0,
            "width": 7.5,
            "height": 0.8,
        },
    }

    // Validation rules
    rules := map[string]string{
        "name": "required|string|min:3|max:100",
        "category": "required|in:electronics,clothing,books,home,food",
        "price": "required|numeric|gt:0",
        "stock": "required|integer|min:0",
        "is_digital": "required|boolean",
        "weight": "required_if:is_digital,false|numeric|gt:0",
        "dimensions.length": "required_if:is_digital,false|numeric|gt:0",
        "dimensions.width": "required_if:is_digital,false|numeric|gt:0",
        "dimensions.height": "required_if:is_digital,false|numeric|gt:0",
    }

    // Validate
    err := v.Validate(data, rules)
    if err != nil {
        validationErrors := err.(validator.Errors)
        for field, messages := range validationErrors {
            for _, message := range messages {
                fmt.Printf("%s: %s\n", field, message)
            }
        }
        return
    }

    fmt.Println("Product validation successful!")
}
```

### Example 3: API Request Validation

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/hbttundar/scg-validator"
    "net/http"
)

// SearchRequest represents a search API request
type SearchRequest struct {
    Query       string   `json:"query"`
    Filters     []string `json:"filters"`
    Page        int      `json:"page"`
    PerPage     int      `json:"per_page"`
    SortBy      string   `json:"sort_by"`
    SortOrder   string   `json:"sort_order"`
}

func validateSearchRequest(w http.ResponseWriter, r *http.Request) {
    // Parse JSON request
    var req SearchRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Convert struct to map for validation
    data := map[string]any{
        "query":       req.Query,
        "filters":     req.Filters,
        "page":        req.Page,
        "per_page":    req.PerPage,
        "sort_by":     req.SortBy,
        "sort_order":  req.SortOrder,
    }

    // Create validator
    v := validator.New()

    // Define validation rules
    rules := map[string]string{
        "query":       "required|string|min:3",
        "filters":     "array",
        "page":        "required|integer|min:1",
        "per_page":    "required|integer|in:10,25,50,100",
        "sort_by":     "required|in:relevance,date,price",
        "sort_order":  "required|in:asc,desc",
    }

    // Validate
    if err := v.Validate(data, rules); err != nil {
        // Convert validation errors to JSON
        validationErrors := err.(validator.Errors)
        errorResponse := map[string]any{
            "status": "error",
            "errors": validationErrors,
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(errorResponse)
        return
    }

    // Process valid request
    response := map[string]string{
        "status": "success",
        "message": "Search request is valid",
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    http.HandleFunc("/search", validateSearchRequest)
    fmt.Println("Server started on :8080")
    http.ListenAndServe(":8080", nil)
}
```

### Example 4: Custom Validation Rule

```go
package main

import (
    "fmt"
    "github.com/hbttundar/scg-validator"
    "strings"
)

// Custom rule to validate a username format
type UsernameFormatRule struct{}

func (r *UsernameFormatRule) Name() string {
    return "username_format"
}

func (r *UsernameFormatRule) Validate(field string, value any, params []string, data map[string]any) bool {
    username, ok := value.(string)
    if !ok {
        return false
    }

    // Username must:
    // 1. Start with a letter
    // 2. Contain only letters, numbers, and underscores
    // 3. Be between 3 and 20 characters

    if len(username) < 3 || len(username) > 20 {
        return false
    }

    firstChar := username[0]
    if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z')) {
        return false
    }

    for _, char := range username {
        if !((char >= 'a' && char <= 'z') || 
             (char >= 'A' && char <= 'Z') || 
             (char >= '0' && char <= '9') || 
             char == '_') {
            return false
        }
    }

    return true
}

func (r *UsernameFormatRule) Message() string {
    return "The :attribute must be 3-20 characters, start with a letter, and contain only letters, numbers, and underscores."
}

func main() {
    // Create validator and register custom rule
    v := validator.New()
    v.RegisterRule("username_format", "The :attribute has an invalid format.", func(ctx *validator.ValidationContext) error {
        rule := &UsernameFormatRule{}
        if !rule.Validate(ctx.Field, ctx.Value, ctx.Params, ctx.Data) {
            return fmt.Errorf(rule.Message())
        }
        return nil
    })

    // Data to validate
    data := map[string]any{
        "username": "john_doe123",
        "display_name": "John Doe",
    }

    // Validation rules
    rules := map[string]string{
        "username": "required|username_format",
        "display_name": "required|string|min:2|max:50",
    }

    // Validate
    err := v.Validate(data, rules)
    if err != nil {
        validationErrors := err.(validator.Errors)
        for field, messages := range validationErrors {
            for _, message := range messages {
                fmt.Printf("%s: %s\n", field, message)
            }
        }
        return
    }

    fmt.Println("Validation passed with custom rule!")
}
```

## Available Validation Rules

### Presence & Type

- `required`: The field must be present and not empty.
- `boolean`: The field must be a boolean value.
- `numeric`: The field must be a numeric value.
- `array`: The field must be an array, slice, or map.
- `integer`: The field must be an integer.
- `string`: The field must be a string.
- `json`: The field must be a valid JSON string.
- `nullable`: The field is optional and can be null.
- `accepted`: The field must be "yes", "on", 1, or true.

### Format

- `email`: The field must be a valid email address.
- `uuid`: The field must be a valid UUID.
- `alpha`: The field must only contain letters.
- `alphanum`: The field must only contain letters and numbers.
- `date`: The field must be a valid date.
- `url`: The field must be a valid URL.
- `ip`: The field must be a valid IP address.
- `regex`: The field must match the given regular expression.
- `password`: The field must meet password complexity requirements.

### File Validation

- `file`: The field must be a file.
- `image`: The field must be an image file.
- `mimes`: The field must be a file of the specified MIME types.

### Size

- `size`: The field must have the specified size.
- `between`: The field must be between the specified minimum and maximum values.
- `min`: The field must be at least the specified value.
- `max`: The field must not be greater than the specified value.
- `gt`: The field must be greater than the specified value.
- `lt`: The field must be less than the specified value.

### Inclusion & Comparison

- `in`: The field must be included in the specified list of values.
- `not_in`: The field must not be included in the specified list of values.
- `confirmed`: The field must have a matching confirmation field.

### Conditional Validation

- `required_if`: The field is required when another field has a specific value.
- `required_unless`: The field is required unless another field has a specific value.
- `required_with`: The field is required when another field is present.
- `required_without`: The field is required when another field is not present.

### Database

- `unique`: The field must be unique in the specified database table.

## Custom Error Messages

You can provide custom error messages for specific fields and rules:

```go
v := validator.New(
    validator.WithCustomMessages(map[string]string{
        "email.required": "Please provide your email address",
        "age.gt": "You must be over :param0 years old",
    }),
)
```

## Custom Attribute Names

You can provide custom attribute names for fields:

```go
v := validator.New(
    validator.WithCustomAttributes(map[string]string{
        "email": "Email Address",
        "first_name": "First Name",
    }),
)
```

## Database Validation

For database validation rules like `unique`, you need to provide a presence verifier:

```go
// Implement the PresenceVerifier interface
type MyDBVerifier struct {
    // Your database connection
}

func (v *MyDBVerifier) Exists(table, column string, value any, excludeIDColumn, excludeIDValue string) (bool, error) {
    // Implement the logic to check if the value exists in the database
}

// Create a validator with the presence verifier
v := validator.New(
    validator.WithPresenceVerifier(&MyDBVerifier{}),
)
```

## Array and Map Validation

The `array` validation rule checks if a field is an array, slice, or map:

```go
data := map[string]any{
    "items": []string{"item1", "item2"},
    "properties": map[string]string{"key1": "value1", "key2": "value2"},
}

rules := map[string]string{
    "items": "array",
    "properties": "array",
}
```

## Custom Validation Rules

You can extend SCG Validator with your own custom validation rules by implementing the `Rule` interface:

```go
// Implement the Rule interface
type MyCustomRule struct{}

func (r *MyCustomRule) Name() string {
    return "custom_rule"
}

func (r *MyCustomRule) Validate(field string, value any, params []string, data map[string]any) bool {
    // Implement your validation logic here
    // Return true if validation passes, false otherwise
    return true
}

func (r *MyCustomRule) Message() string {
    return "The :attribute failed the custom validation rule."
}

// Register your custom rule with the validator
v := validator.New()
v.RegisterRule(&MyCustomRule{})

// Use your custom rule
rules := map[string]string{
    "field": "required|custom_rule",
}
```

## Advanced Usage

### Conditional Validation

You can apply validation rules conditionally:

```go
data := map[string]any{
    "user_type": "admin",
    "admin_code": "12345",
}

rules := map[string]string{
    "user_type": "required|in:admin,user",
    "admin_code": "required_if:user_type,admin|numeric",
}
```

### Nested Data Validation

You can validate nested data structures using dot notation:

```go
data := map[string]any{
    "user": map[string]any{
        "name": "John Doe",
        "email": "john@example.com",
    },
}

rules := map[string]string{
    "user.name": "required|string",
    "user.email": "required|email",
}
```

## Testing

To run the tests for SCG Validator:

```bash
go test ./... -v
```

For test coverage:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Development

### CI/CD Pipeline

This project uses GitHub Actions for continuous integration and continuous deployment. The pipeline includes:

- Building and testing the code
- Linting with golangci-lint
- Security scanning with Gosec and Nancy

The configuration for the CI/CD pipeline is in the `.github/workflows/ci.yml` file.

### Linting

This project uses golangci-lint for code quality checks. The configuration is in the `.golangci.yml` file.

To run the linter locally:

```bash
golangci-lint run
```

### Security Scanning

This project uses Gosec for security scanning. The configuration is integrated into the `.golangci.yml` file for a unified linting and security scanning approach.

To run the security scanner through golangci-lint:

```bash
golangci-lint run --enable=gosec
```

Alternatively, you can run gosec directly:

```bash
gosec ./...
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/). The version format is `MAJOR.MINOR.PATCH`:

- `MAJOR` version increments for incompatible API changes
- `MINOR` version increments for backwards-compatible functionality additions
- `PATCH` version increments for backwards-compatible bug fixes

## License

This project is licensed under the MIT License - see the LICENSE file for details.
