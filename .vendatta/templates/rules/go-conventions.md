---
title: "Go Language Conventions"
description: "Enforce Go language best practices and conventions"
applies_to: ["**/*.go"]
priority: high
enabled: true
---

# Go Language Conventions

This document outlines the coding standards and conventions for Go development in this project.

## Code Organization

### Package Structure
- Use short, concise package names
- Package names should be lowercase, no underscores
- Avoid package names like `util`, `common`, `misc` - be specific
- Group related functionality into packages

### File Naming
- Use snake_case for file names: `user_service.go`, `config_parser.go`
- Test files: `*_test.go`
- Package files should be named after their primary type or function

## Code Style

### Formatting
- Use `gofmt` for consistent formatting
- Run `go fmt ./...` before committing
- Maximum line length: 120 characters

### Imports
```go
// Standard library imports first
import (
    "fmt"
    "os"
    "strings"
)

// Blank line separates standard library from third-party
import (
    "github.com/spf13/cobra"
    "github.com/vibegear/vendetta/pkg/config"
)

// Local imports last
import (
    "project/internal/auth"
    "project/pkg/models"
)
```

### Variable Naming
- Use camelCase for variables and functions
- Exported identifiers: PascalCase
- Unexported identifiers: camelCase
- Acronyms: HTTPClient, not HttpClient
- Single letter variables only for loops and errors: `i`, `err`

### Constants
```go
// Use PascalCase for exported constants
const (
    DefaultPort     = 8080
    MaxRetries      = 3
    ConfigFileName  = "config.yaml"
)
```

## Error Handling

### Error Wrapping
- Always wrap errors with context
- Use `fmt.Errorf` with `%w` verb
```go
if err != nil {
    return fmt.Errorf("failed to connect to database: %w", err)
}
```

### Error Types
- Define custom error types for specific error conditions
- Use error variables for sentinel errors
```go
var ErrNotFound = errors.New("resource not found")

type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Message)
}
```

## Functions and Methods

### Function Signatures
- Keep functions focused on single responsibility
- Limit to 3-4 parameters maximum
- Use struct parameters for multiple related values
```go
// Good
func CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)

// Avoid
func CreateUser(name, email string, age int, active bool) (*User, error)
```

### Receivers
- Use pointer receivers for methods that modify the receiver
- Use value receivers for immutable methods
```go
func (u *User) UpdateEmail(email string) error {
    // Modifies user, use pointer receiver
}

func (u User) IsActive() bool {
    // Doesn't modify, use value receiver
}
```

## Structs and Types

### Struct Definition
```go
type User struct {
    ID        int64     `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

### Constructors
- Provide constructor functions for complex structs
- Use `New` prefix for constructors
```go
func NewUser(email string) (*User, error) {
    if !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }
    return &User{
        Email:     email,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}
```

## Testing

### Test Structure
- Use table-driven tests for multiple test cases
- Test files: `*_test.go`
- Test functions: `TestFunctionName`
- Helper functions: `testHelperFunction`

### Test Examples
```go
func TestUserCreation(t *testing.T) {
    tests := []struct {
        name     string
        email    string
        wantErr  bool
        errType  error
    }{
        {"valid email", "user@example.com", false, nil},
        {"empty email", "", true, ErrInvalidEmail},
        {"invalid format", "not-an-email", true, ErrInvalidEmail},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, err := NewUser(tt.email)
            if tt.wantErr {
                assert.Error(t, err)
                assert.IsType(t, tt.errType, err)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
                assert.Equal(t, tt.email, user.Email)
            }
        })
    }
}
```

## Performance Considerations

### Efficient Code
- Prefer `strings.Builder` for string concatenation in loops
- Use `sync.Pool` for frequently allocated objects
- Avoid unnecessary allocations in hot paths

### Memory Management
- Be mindful of pointer vs value semantics
- Use `context.WithCancel` for goroutine cancellation
- Properly close resources (files, connections, etc.)

## Documentation

### Package Comments
- Every package must have a package comment
- Explain the package's purpose and usage

### Function Comments
- Exported functions must have comments
- Start with the function name
- Explain parameters, return values, and behavior

### Example
```go
// Package user provides user management functionality.
package user

// CreateUser creates a new user with the given email address.
// It validates the email format and ensures uniqueness.
// Returns the created user or an error if creation fails.
func CreateUser(ctx context.Context, email string) (*User, error) {
    // implementation...
}
```

## Security

### Input Validation
- Always validate user input
- Use allowlists rather than blocklists
- Sanitize data before processing

### Sensitive Data
- Never log sensitive information (passwords, tokens, etc.)
- Use secure random number generation
- Implement proper authentication and authorization

## Tools and Linting

### Required Tools
- `gofmt` - Code formatting
- `go vet` - Basic static analysis
- `golint` or `golangci-lint` - Comprehensive linting

### CI/CD
- Run tests with `go test ./...`
- Check formatting with `gofmt -l .`
- Run linters in CI pipeline

## Common Anti-Patterns

### Avoid
- Global variables
- init() functions for complex initialization
- Panic for expected errors
- interface{} overuse
- Deep nesting (keep functions shallow)

### Prefer
- Dependency injection
- Explicit initialization
- Error return values
- Specific types and interfaces
- Early returns
