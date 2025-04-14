---
layout: default
title: goenums - Type-Safe Enum Generation for Go
---

`goenums` addresses Go's lack of native enum support by generating comprehensive, type-safe enum implementations from simple constant declarations. Transform basic `iota` based constants into feature-rich enums with string conversion, validation, JSON handling, database integration, and more.

## Key Features

- **Type Safety**: Wrapper types prevent accidental misuse of enum values
- **String Conversion**: Automatic string representation and parsing
- **JSON Support**: Built-in marshaling and unmarshaling 
- **Database Integration**: SQL Scanner and Valuer implementations
- **Validation**: Methods to check for valid enum values
- **Iteration**: Modern Go 1.23+ iteration support with legacy fallback
- **Extensibility**: Add custom fields to enums via comments
- **Exhaustive Handling**: Helper functions to ensure you handle all enum values
- **Zero Dependencies**: Completely dependency-free, using only the Go standard library

## Quick Start

```go
package validation

type status int

//go:generate goenums status.go
const (
    unknown status = iota // invalid
    failed
    passed
    skipped
    scheduled
    running
    booked
)
```
Generate your enums:

```bash
$ go generate ./...
```

Now you can use the generated `status` enums type in your code:

```go
/// Access enum constants safely
myStatus := validation.Statuses.PASSED

// Convert to string
fmt.Println(myStatus.String()) // "PASSED"

// Parse from various sources
parsed, _ := validation.ParseStatus("SKIPPED")

// Validate enum values
if !parsed.IsValid() {
    fmt.Println("Invalid status")
}
```
Get Started → [Installation](/installation)