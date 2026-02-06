# AGENTS.md

This file contains guidelines and commands for agentic coding agents working on the mbserver project.

## Project Overview

mbserver is a Go implementation of a Modbus server (slave) supporting TCP and RTU protocols. It provides complete Modbus function support (function codes 1, 2, 3, 4, 5, 6, 15, 16), custom register implementations, and extensible function handlers.

## Build/Test/Lint Commands

### Core Commands
```bash
# Run all tests
go test -v ./...

# Run a single test file
go test -v ./server_test.go

# Run a specific test function
go test -v -run TestModbus ./...

# Run tests with coverage
go test -v -cover ./...

# Build the main command
go build -o bin/mbserver ./cmd

# Format all Go files
go fmt ./...

# Run static analysis
go vet ./...

# Generate stringer code (for Exception type)
go generate ./...
```

### Module Management
```bash
# Tidy dependencies
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

## Code Style Guidelines

### Import Organization
- Group imports into three sections: standard library, third-party, and local packages
- Use blank line between groups
- Order within groups is alphabetical
- Example:
```go
import (
    "context"
    "encoding/binary"
    "sync"

    "github.com/goburrow/modbus"
    "github.com/goburrow/serial"
    "github.com/stretchr/testify/assert"

    "github.com/leijux/mbserver"
)
```

### Naming Conventions
- **Package names**: short, lowercase, single words (e.g., `mbserver`)
- **Constants**: `CamelCase` with public visibility starting with capital letter
- **Variables**: `camelCase` for local, `CamelCase` for exported
- **Functions**: `CamelCase` for exported, `camelCase` for unexported
- **Types/Interfaces**: `CamelCase`
- **Interface names**: typically end with `-er` suffix (e.g., `Register`, `Framer`)
- **Error variables**: `CamelCase` (e.g., `IllegalFunction`, `IllegalDataAddress`)

### Code Formatting
- Use `gofmt` for consistent formatting
- Maximum line length: 120 characters (not strictly enforced)
- Use tabs for indentation (Go standard)
- No trailing whitespace

### Error Handling
- Use custom `Exception` type for Modbus-specific errors
- Return errors as second return value following Go conventions
- Use `Success` constant for successful operations
- Implement `error` interface on custom types
- Example error handling pattern:
```go
if start+count > len(r.Coils) {
    return nil, IllegalDataAddress
}
return r.Coils[start : start+count], Success
```

### Type Definitions
- Use type definitions for clarity and code generation
- Implement interface compliance checks: `var _ Register = (*MemRegister)(nil)`
- Use `//go:generate` comments for code generation (e.g., stringer)

### Testing Guidelines
- Use `testify/assert` for assertions and `testify/require` for fatal assertions
- Write descriptive test names following `Test[FunctionName]` pattern
- Use table-driven tests for multiple similar cases
- Use `t.Cleanup()` for resource cleanup instead of `defer` in tests
- Organize related tests with `t.Run()` for subtests
- Example test structure:
```go
func TestModbus(t *testing.T) {
    s := NewServer()
    err := s.ListenTCP("127.0.0.1:3333")
    require.NoError(t, err)
    t.Cleanup(s.Shutdown)
    
    t.Run("Coils", func(t *testing.T) {
        // Test coils functionality
        assert.Equal(t, expected, actual)
    })
}
```

### Interface Design
- Keep interfaces small and focused
- Use composition for complex interfaces
- Provide default implementations (e.g., `MemRegister`)
- Use functional options pattern for configuration:
```go
type OptionFunc func(s *Server)

func WithRegister(register Register) OptionFunc {
    return func(s *Server) {
        s.register = register
    }
}
```

### Concurrency Patterns
- Use `sync.WaitGroup` for coordinating goroutines
- Use channels for communication between goroutines
- Protect shared state with mutexes when necessary
- Use `select` statements for handling multiple channels
- Example from server:
```go
func (s *Server) handler() {
    for {
        select {
        case <-s.closeSignalChan:
            return
        case request := <-s.requestChan:
            response := s.handle(request)
            request.conn.Write(response.Bytes())
        }
    }
}
```

### Constants and Magic Numbers
- Define constants for protocol-specific values
- Use `iota` for related constants
- Add comments explaining protocol requirements
- Example:
```go
const (
    Success Exception = iota
    IllegalFunction
    IllegalDataAddress
    // ...
)
```

### Documentation
- Add package comments explaining purpose
- Document public functions and types
- Use proper capitalization for exported symbols
- Include example code in comments for complex usage

## Dependencies
- Go 1.24+ required
- Key dependencies:
  - `github.com/goburrow/modbus` - Modbus client for testing
  - `github.com/goburrow/serial` - Serial communication
  - `github.com/stretchr/testify` - Testing utilities

## Testing Tips
- Tests use a local TCP server on port 3333 for integration testing
- Use `time.Sleep()` briefly after starting server to avoid connection refused
- Test both success and error cases for each function
- Verify boundary conditions (start+count validations)
- Test protocol compliance (byte order, data formats)

## File Organization
- `server.go` - Main server implementation
- `function.go` - Modbus function implementations
- `frame*.go` - Protocol frame implementations (TCP, RTU)
- `register.go` - Register interface and memory implementation
- `exception.go` - Exception types and codes
- `cmd/main.go` - Example server application
- `*_test.go` - Test files for corresponding implementations