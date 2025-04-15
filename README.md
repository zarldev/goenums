# goenums

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![build](https://github.com/zarldev/goenums/actions/workflows/go.yml/badge.svg)

`goenums` addresses Go's lack of native enum support by generating comprehensive, type-safe enum implementations from simple constant declarations. Transform basic `iota` based constants into feature-rich enums with string conversion, validation, JSON handling, database integration, and more.

## Table of Contents
- [Documentation](#documentation)
- [Installation](#installation)
- [Key Features](#key-features)
- [Usage](#usage)
- [Getting Started](#getting-started)
  - [Basic Example](#basic-example)
- [Advanced Features](#advanced-features)
  - [Custom String Representations](#custom-string-representations)
  - [Extended Enum Types with Custom Fields](#extended-enum-types-with-custom-fields)
  - [Strict Validation](#strict-validation)
  - [Case Insensitive String Parsing](#case-insensitive-string-parsing)
  - [JSON & Database Storage](#json--database-storage)
  - [Exhaustive Handling](#exhaustive-handling)
  - [Iterator Support (Go 1.23+)](#iterator-support-go-123)
- [Requirements](#requirements)
- [Examples](#examples)
- [License](#license)

# Documentation
Documentation is available at [https://zarldev.github.io/goenums](https://zarldev.github.io/goenums).

# Installation
```
go install github.com/zarldev/goenums@latest
```

# Key Features
 - Type Safety: Wrapper types prevent accidental misuse of enum values
 - String Conversion: Automatic string representation and parsing
 - JSON Support: Built-in marshaling and unmarshaling
 - Database Integration: SQL Scanner and Valuer implementations
 - Validation: Methods to check for valid enum values
 - Iteration: Modern Go 1.23+ iteration support with legacy fallback
 - Extensibility: Add custom fields to enums via comments
 - Exhaustive Handling: Helper functions to ensure you handle all enum values
 - Zero Dependencies: Completely dependency-free, using only the Go standard library

# Usage
```
$ goenums -h
   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/
Usage: goenums [options] filename
Options:
  -f, -failfast
        Enable failfast mode - fail on invalid enum parsing (default: false)
  -l, -legacy
        Generate legacy code without Go 1.23+ iterator support (default: false)
  -i, -insensitive
        Enable case-insensitive string parsing (default: false)
  -h, -help
        Print help information
  -v, -version
        Print version information
  -vv, -verbose
        Enable verbose logging (default: false)
```

# Getting Started

## Basic Example

goenums is designed to work seamlessly with Go's standard tooling, particularly with `go:generate` directives. This allows you to automatically regenerate your enum code whenever your source files change, integrating smoothly into your existing build process.

1. Define your enum constant in a Go file:

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
2. Run `go generate ./...` to generate the enum implementations.

3. Use the generated `status` enums type in your code:

```go
// Access enum constants safely
myStatus := validation.Statuses.PASSED

// Convert to string
fmt.Println(myStatus.String()) // "PASSED"

// Parse from various sources
input := "SKIPPED"
parsed, _ := validation.ParseStatus(input)

// Validate enum values
if !parsed.IsValid() {
    fmt.Println("Invalid status")
}

// JSON marshaling/unmarshaling works automatically
type Task struct {
    ID     int              `json:"id"`
    Status validation.Status `json:"status"`
}
```

# Advanced Features

## Custom String Representations

Handle custom string representations defined in the comments of each enum.  Support for enum strings with spaces is supported by adding the alternative name in double quotes: 

### Standard Name Comment

When the Alternative name does not contain spaces there is no need
to add the double quotes.

```go
type ticketStatus int

//go:generate goenums status.go
const (
    unknown   ticketStatus = iota // invalid Unknown
    pending                       // Pending
    approved                      // Approved
    rejected                      // Rejected
    completed                     // Completed
)
```

### Name Comment with spaces

When using Alternative names that contain spaces, the double quotes are required.

```go
type ticketStatus int

//go:generate goenums status.go
const (
    unknown   ticketStatus = iota // invalid "Not Found"
    pending                       // "In Progress"
    approved                      // "Fully Approved"
    rejected                      // "Has Been Rejected"
    completed                     // "Successfully Completed"
)
```
## Extended Enum Types with Custom Fields
Add custom fields to your enums with type comments:

```go
// Define fields in the type comment using one of three formats:
// 1. Space-separated: "Field Type,AnotherField Type"
// 2. Brackets: "Field[Type],AnotherField[Type]"
// 3. Parentheses: "Field(Type),AnotherField(Type)"

type planet int // Gravity float64,RadiusKm float64,MassKg float64,OrbitKm float64

//go:generate goenums planets.go
const (
    unknown planet = iota // invalid
    mercury               // Mercury 0.378,2439.7,3.3e23,57910000
    venus                 // Venus 0.907,6051.8,4.87e24,108200000
    earth                 // Earth 1,6378.1,5.97e24,149600000
	... 
)
```
Then we can use the extended enum type:

```go
earthWeight := 100.0
fmt.Printf("Weight on %s: %.2f kg\n", 
    solarsystem.Planets.MARS, 
    earthWeight * solarsystem.Planets.MARS.Gravity)
```

## Strict Validation
Use the -f flag to enable strict validation that returns errors for invalid enum values:

```go
//go:generate goenums -f status.go

// Generated code will return errors for invalid values
status, err := validation.ParseStatus("INVALID_STATUS")
if err != nil {
    fmt.Println("error:", err)
}
```

## Case Insensitive String Parsing
Use the -i flag to enable case insensitive string parsing:

```go
//go:generate goenums -i status.go

// Generated code will parse case insensitive strings. All
// of the below will validate and produce the 'Pending' enum
status, err := validation.ParseStatus("Pending")
if err != nil {
    fmt.Println("error:", err)
}
status, err := validation.ParseStatus("pending")
if err != nil {
    fmt.Println("error:", err)
}
status, err := validation.ParseStatus("PENDING")
if err != nil {
    fmt.Println("error:", err)
}
```

## JSON & Database Storage
The generated enum type also implements the `json.Unmarshal` and `json.Marshal` interfaces along with the `sql.Scanner` and `sql.Valuer` interfaces to handle parsing over the wire via HTTP or via a Database.

```go
func (p Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *Status) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, `"`), ` `)
	newp, err := ParseStatus(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

func (p *Status) Scan(value any) error {
	newp, err := ParseStatus(value)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

func (p Status) Value() (driver.Value, error) {
	return p.String(), nil
}
```

## Exhaustive Handling
Ensure you handle all enum values with the generated Exhaustive function:

```go
// Process all enum values safely
// This is especially useful in tests to ensure all enum values are covered
validation.ExhaustiveStatuses(func(status validation.Status) {
    // Process each status exactly once
    switch status {
    case validation.Statuses.FAILED:
        handleFailed()
    case validation.Statuses.PASSED:
        handlePassed()
    // ... handle all other cases
    }
})

// We can also iterate over all enum values to do exhaustive calculations
weightKg := 100.0
solarsystem.ExhaustivePlanets(func(p solarsystem.Planet) {
	// calculate weight on each planet
	gravity := p.Gravity
	planetMass := weightKg * gravity
	fmt.Printf("Weight on %s is %fKg with gravity %f\n", p, planetMass, gravity)
})
```

## Iterator Support (Go 1.23+)
By default, goenums generates modern iterator support using Go 1.23's range-over-func feature:

```go 
// Using Go 1.23+ iterator
for status := range validation.Statuses.All() {
    fmt.Printf("Status: %s\n", status)
}
// There is the fallback method for slice based access
for _, status := range validation.Statuses.AllSlice() {
    fmt.Printf("Status: %s\n", status)
}
```

When using the legacy mode, the function is still called All() but it returns a slice of the enums.

```go
// Legacy mode (or with -l flag)
for _, status := range validation.Statuses.All() {
    fmt.Printf("Status: %s\n", status)
}
```


# Requirements
- Go 1.23+ for iterator support (or use -l flag for legacy mode)

# Examples

Input source go file:

```golang
package validation

type status int

//go:generate goenums status.go

const (
	failed    status = iota // FAILED
	passed                  // PASSED
	skipped                 // SKIPPED
	scheduled               // SCHEDULED
	running                 // RUNNING
	booked                  // BOOKED
)
```

Produces a go output file called `statuses_enums.go` with the following content:

```go
// Code generated by goenums v0.3.6. DO NOT EDIT.
// This file was generated by github.com/zarldev/goenums
// using the command:
// goenums testdata/multiple/multiple.go

package multipleenums

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"iter"
	"strconv"
)

type Status struct {
	status
}

type statusesContainer struct {
	FAILED    Status
	PASSED    Status
	SKIPPED   Status
	SCHEDULED Status
	RUNNING   Status
	BOOKED    Status
}

var Statuses = statusesContainer{
	FAILED: Status{
		status: failed,
	},
	PASSED: Status{
		status: passed,
	},
	SKIPPED: Status{
		status: skipped,
	},
	SCHEDULED: Status{
		status: scheduled,
	},
	RUNNING: Status{
		status: running,
	},
	BOOKED: Status{
		status: booked,
	},
}

var invalidStatus = Status{}

func (c statusesContainer) allSlice() []Status {
	return []Status{
		c.FAILED,
		c.PASSED,
		c.SKIPPED,
		c.SCHEDULED,
		c.RUNNING,
		c.BOOKED,
	}
}

// AllSlice returns all valid Status values as a slice.
// Deprecated: Use All() with Go 1.23+ range over function types instead.
func (c statusesContainer) AllSlice() []Status {
	return c.allSlice()
}

// All returns all valid Status values.
func (c statusesContainer) All() iter.Seq[Status] {
	return func(yield func(Status) bool) {
		for _, v := range c.allSlice() {
			if !yield(v) {
				return
			}
		}
	}
}

func ParseStatus(a any) (Status, error) {
	res := invalidStatus
	switch v := a.(type) {
	case Status:
		return v, nil
	case []byte:
		res = stringToStatus(string(v))
	case string:
		res = stringToStatus(v)
	case fmt.Stringer:
		res = stringToStatus(v.String())
	case int:
		res = intToStatus(v)
	case int64:
		res = intToStatus(int(v))
	case int32:
		res = intToStatus(int(v))
	}
	return res, nil
}

var (
	_statusesNameMap = map[string]Status{
		"FAILED":    Statuses.FAILED,
		"PASSED":    Statuses.PASSED,
		"SKIPPED":   Statuses.SKIPPED,
		"SCHEDULED": Statuses.SCHEDULED,
		"RUNNING":   Statuses.RUNNING,
		"BOOKED":    Statuses.BOOKED,
	}
)

func stringToStatus(s string) Status {
	if v, ok := _statusesNameMap[s]; ok {
		return v
	}
	return invalidStatus
}

func intToStatus(i int) Status {
	if i < 0 || i >= len(Statuses.allSlice()) {
		return invalidStatus
	}
	return Statuses.allSlice()[i]
}

func ExhaustiveStatuss(f func(Status)) {
	for _, p := range Statuses.allSlice() {
		f(p)
	}
}

var validStatuses = map[Status]bool{
	Statuses.FAILED:    true,
	Statuses.PASSED:    true,
	Statuses.SKIPPED:   true,
	Statuses.SCHEDULED: true,
	Statuses.RUNNING:   true,
	Statuses.BOOKED:    true,
}

func (p Status) IsValid() bool {
	return validStatuses[p]
}

func (p Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *Status) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, `"`), ` `)
	newp, err := ParseStatus(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

func (p *Status) Scan(value any) error {
	newp, err := ParseStatus(value)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

func (p Status) Value() (driver.Value, error) {
	return p.String(), nil
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the goenums command to generate them again.
	// Does not identify newly added constant values unless order changes
	var x [1]struct{}
	_ = x[failed-0]
	_ = x[passed-1]
	_ = x[skipped-2]
	_ = x[scheduled-3]
	_ = x[running-4]
	_ = x[booked-5]
}

const _statuses_name = "FAILEDPASSEDSKIPPEDSCHEDULEDRUNNINGBOOKED"

var _statuses_index = [...]uint16{0, 6, 12, 19, 28, 35, 41}

func (i status) String() string {
	if i < 0 || i >= status(len(_statuses_index)-1) {
		return "statuses(" + (strconv.FormatInt(int64(i), 10) + ")")
	}
	return _statuses_name[_statuses_index[i]:_statuses_index[i+1]]
}
```

For more examples, see the [examples](https://github.com/zarldev/goenums/tree/main/examples) directory.


### Mentions
[![go-recipes](https://raw.githubusercontent.com/nikolaydubina/go-recipes/main/badge.svg?raw=true)](https://github.com/nikolaydubina/go-recipes)

## License

MIT License - See [LICENSE](https://github.com/zarldev/goenums/blob/main/LICENSE) for full details.