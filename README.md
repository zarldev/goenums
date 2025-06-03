# goenums

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![build](https://github.com/zarldev/goenums/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/zarldev/goenums)](https://goreportcard.com/report/github.com/zarldev/goenums)

`goenums` addresses Go's lack of native enum support by generating comprehensive, type-safe enum implementations from simple constant declarations. Transform basic `iota` based constants into feature-rich enums with string conversion, validation, JSON handling, database integration, and more.

# Installation
```
go install github.com/zarldev/goenums@latest
```

# Documentation
Documentation is available at [https://zarldev.github.io/goenums](https://zarldev.github.io/goenums).

## Table of Contents
- [Key Features](#key-features)
- [Usage](#usage)
- [Features Expanded](#features-expanded)
  - [Custom String Representations](#custom-string-representations)
    - [Standard Name Comment](#standard-name-comment)
    - [Name Comment with spaces](#name-comment-with-spaces)
  - [Extended Enum Types with Custom Fields](#extended-enum-types-with-custom-fields)
  - [Case Insensitive String Parsing](#case-insensitive-string-parsing)
  - [JSON, Text, Binary, YAML, and Database Storage](#json-text-binary-yaml-and-database-storage)
  - [Numeric Parsing Support](#numeric-parsing-support)
  - [Exhaustive Handling](#exhaustive-handling)
  - [Iterator Support (Go 1.23+)](#iterator-support-go-123)
  - [Failfast Mode / Strict Mode](#failfast-mode--strict-mode)
  - [Legacy Mode](#legacy-mode)
  - [Verbose Mode](#verbose-mode)
  - [Constraints Mode](#constraints-mode)
  - [Output Format](#output-format)
  - [Compile-time Validation](#compile-time-validation)
- [Getting Started](#getting-started)
  - [Basic Example](#basic-example)
- [Requirements](#requirements)
- [Examples](#examples)
- [License](#license)

# Key Features
 - Type Safety: Wrapper types prevent accidental misuse of enum values
 - String Conversion: Automatic string representation and parsing
 - JSON Support: Built-in marshaling and unmarshaling
 - YAML Support: Built-in YAML marshaling and unmarshaling
 - Database Integration: SQL Scanner and Valuer implementations
 - Text/Binary Marshaling: Support for encoding.TextMarshaler/TextUnmarshaler and BinaryMarshaler/BinaryUnmarshaler
 - Numeric Parsing: Parse enums from various numeric types (int, float, etc.)
 - Validation: Methods to check for valid enum values
 - Iteration: Modern Go 1.23+ iteration support with legacy fallback
 - Extensibility: Add custom fields to enums via comments
 - Exhaustive Handling: Helper functions to ensure you handle all enum values
 - Alias Support: Alternative enum names via comment syntax
 - Zero Dependencies: Completely dependency-free, using only the Go standard library

# Usage
```
$ goenums -h
   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/
Usage: goenums [options] file.go[,file2.go,...]
Options:
  -c
  -constraints
    	Specify whether to generate the float and integer constraints or import 'golang.org/x/exp/constraints' (default: false - imports)
  -f
  -failfast
    	Enable failfast mode - fail on generation of invalid enum while parsing (default: false)
  -h
  -help
    	Print help information
  -i
  -insensitive
    	Generate case insensitive string parsing (default: false)
  -l
  -legacy
    	Generate legacy code without Go 1.23+ iterator support (default: false)
  -o string
  -output string
    	Specify the output format (default: go)
  -v
  -version
    	Print version information
  -vv
  -verbose
    	Enable verbose mode - prints out the generated code (default: false)
```
# Features Expanded

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
type order int

//go:generate goenums status.go
const (
	created     order = iota // "CREATED"
	approved                 // "APPROVED"
	processing               // "PROCESSING"
	readyToShip              // "READY TO SHIP"
	shipped                  // "SHIPPED TO CUSTOMER"
	delivered                // "DELIVERED TO CUSTOMER"
	cancelled                // "CANCELLED BY CUSTOMER"
	refunded                 // "REFUNDED TO CUSTOMER"
	closed                   // "CLOSED"
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

## JSON, Text, Binary, YAML, and Database Storage
The generated enum type also implements several common interfaces:
* `json.Marshaler` and `json.Unmarshaler`
* `sql.Scanner` and `sql.Valuer`
* `encoding.TextMarshaler` and `encoding.TextUnmarshaler`
* `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler`

These interfaces are used to handle parsing for JSON, Text, Binary, and Database storage using the common standard library packages.
As there is no standard library support for YAML, the generated YAML marshaling and unmarshaling methods are based on the * `yaml.Marshaler` and `yaml.Unmarshaler` interfaces from the [goccy/go-yaml](https://github.com/goccy/go-yaml) module.

Here is an example of the generated handling code:

```go
// MarshalJSON implements the json.Marshaler interface for Status.
// It returns the JSON representation of the enum value as a byte slice.
func (p Status) MarshalJSON() ([]byte, error) {
	return []byte("\"" + p.String() + "\""), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for Status.
// It parses the JSON representation of the enum value from the byte slice.
// It returns an error if the input is not a valid JSON representation.
func (p *Status) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, "\""), "\"")
	newp, err := ParseStatus(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface for Status.
// It returns the string representation of the enum value as a byte slice
func (p Status) MarshalText() ([]byte, error) {
	return []byte("\"" + p.String() + "\""), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for Status.
// It parses the string representation of the enum value from the byte slice.
// It returns an error if the byte slice does not contain a valid enum value.
func (p *Status) UnmarshalText(b []byte) error {
	newp, err := ParseStatus(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

// Scan implements the database/sql.Scanner interface for Status.
// It parses the string representation of the enum value from the database row.
// It returns an error if the row does not contain a valid enum value.
func (p *Status) Scan(value any) error {
	newp, err := ParseStatus(value)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

// Value implements the database/sql/driver.Valuer interface for Status.
// It returns the string representation of the enum value.
func (p Status) Value() (driver.Value, error) {
	return p.String(), nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface for Status.
// It returns the binary representation of the enum value as a byte slice.
func (p Status) MarshalBinary() ([]byte, error) {
	return []byte("\"" + p.String() + "\""), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for Status.
// It parses the binary representation of the enum value from the byte slice.
// It returns an error if the byte slice does not contain a valid enum value.
func (p *Status) UnmarshalBinary(b []byte) error {
	newp, err := ParseStatus(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}
```

## Numeric Parsing Support
The generated enums support parsing from various numeric types, automatically converting them to the appropriate enum value:

```go
// Parse from different numeric types
status1, _ := validation.ParseStatus(1)        // int
status2, _ := validation.ParseStatus(int32(2)) // int32
status3, _ := validation.ParseStatus(3.0)      // float64
status4, _ := validation.ParseStatus(uint8(4)) // uint8

// All numeric types are supported: int, int8, int16, int32, int64,
// uint, uint8, uint16, uint32, uint64, float32, float64
```

The numeric parsing validates that:
- Float values are whole numbers (no fractional part)
- The numeric value corresponds to a valid enum position
- Values are within the valid range of enum constants

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
```

When using the legacy mode, the function is still called All() but it returns a slice of the enums.

```go
// Legacy mode (or with -l flag)
for _, status := range validation.Statuses.All() {
    fmt.Printf("Status: %s\n", status)
}
```

## Failfast Mode / Strict Mode
You can enable failfast mode by using the `-failfast` flag. This will cause the generator to fail on the first invalid enum it encounters while parsing.
```go
//go:generate goenums -f status.go

// Generated code will return errors for invalid values
status, err := validation.ParseStatus("INVALID_STATUS")
if err != nil {
    fmt.Println("error:", err)
}
```

## Legacy Mode
You can enable legacy mode by using the `-legacy` flag. This will generate code that is compatible with Go versions before 1.23.

## Verbose Mode
You can enable verbose mode by using the `-verbose` flag. This will print out the generated code to the console.

## Constraints Mode
You can enable constraints mode by using the `-constraints` flag. This will generate local type constraints instead of importing `golang.org/x/exp/constraints`. This is useful if you want to avoid external dependencies.

```go
//go:generate goenums -c status.go

// Generated code will include local constraint definitions:
type float interface {
    float32 | float64
}
type integer interface {
    int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr
}
type number interface {
    integer | float
}
```

## Output Format
You can specify the output format by using the `-output` flag. The default is `go`.

## Compile-time Validation
The generated code includes compile-time validation to ensure enum values remain consistent. If you modify the underlying enum constants, the compiler will detect changes and prompt you to regenerate the enum code:

```go
// Compile-time check that all enum values are valid.
// This function is used to ensure that all enum values are defined and valid.
// It is called by the compiler to verify that the enum values are valid.
func _() {
    // An "invalid array index" compiler error signifies that the constant values have changed.
    // Re-run the goenums command to generate them again.
    // Does not identify newly added constant values unless order changes
    var x [7]struct{}
    _ = x[unknown-0]
    _ = x[failed-1]
    _ = x[passed-2]
    _ = x[skipped-3]
    _ = x[scheduled-4]
    _ = x[running-5]
    _ = x[booked-6]
}
```

This ensures that if you change the order or values of your enum constants, you'll get a compile error reminding you to regenerate the enum code.

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

# Requirements
- Go 1.23+ for iterator support (or use -l flag for legacy mode)

# Examples

Input source go file:

```go
package solarsystem

type planet int // Gravity float64,RadiusKm float64,MassKg float64,OrbitKm float64,OrbitDays float64,SurfacePressureBars float64,Moons int,Rings bool

//go:generate goenums planets.go
const (
	unknown planet = iota // invalid
	mercury               // Mercury 0.378,2439.7,3.3e23,57910000,88,0.0000000001,0,false
	venus                 // Venus 0.907,6051.8,4.87e24,108200000,225,92,0,false
	earth                 // Earth 1,6378.1,5.97e24,149600000,365,1,1,false
	mars                  // Mars 0.377,3389.5,6.42e23,227900000,687,0.01,2,false
	jupiter               // Jupiter 2.36,69911,1.90e27,778600000,4333,20,4,true
	saturn                // Saturn 0.916,58232,5.68e26,1433500000,10759,1,7,true
	uranus                // Uranus 0.889,25362,8.68e25,2872500000,30687,1.3,13,true
	neptune               // Neptune 1.12,24622,1.02e26,4495100000,60190,1.5,2,true
)
```

Produces a go output file called `planets_enums.go` with the following content:

```go
// code generated by goenums 'v0.4.0' at Jun  2 00:22:41. DO NOT EDIT.
//
// github.com/zarldev/goenums
//
// using the command:
// goenums  planets.go

package solarsystem

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"iter"
	"math"

	"golang.org/x/exp/constraints"
)

// Planet is a type that represents a single enum value.
// It combines the core information about the enum constant and it's defined fields.
type Planet struct {
	planet
	Gravity             float64
	RadiusKm            float64
	MassKg              float64
	OrbitKm             float64
	OrbitDays           float64
	SurfacePressureBars float64
	Moons               int
	Rings               bool
}

// planetsContainer is the container for all enum values.
// It is private and should not be used directly use the public methods on the Planet type.
type planetsContainer struct {
	UNKNOWN Planet
	MERCURY Planet
	VENUS   Planet
	EARTH   Planet
	MARS    Planet
	JUPITER Planet
	SATURN  Planet
	URANUS  Planet
	NEPTUNE Planet
}

// Planets is a main entry point using the Planet type.
// It it a container for all enum values and provides a convenient way to access all enum values and perform
// operations, with convenience methods for common use cases.
var Planets = planetsContainer{
	MERCURY: Planet{
		planet:              mercury,
		Gravity:             0.378,
		RadiusKm:            2439.7,
		MassKg:              3.3e+23,
		OrbitKm:             5.791e+07,
		OrbitDays:           88,
		SurfacePressureBars: 1e-10,
		Moons:               0,
		Rings:               false,
	},
	VENUS: Planet{
		planet:              venus,
		Gravity:             0.907,
		RadiusKm:            6051.8,
		MassKg:              4.87e+24,
		OrbitKm:             1.082e+08,
		OrbitDays:           225,
		SurfacePressureBars: 92,
		Moons:               0,
		Rings:               false,
	},
	EARTH: Planet{
		planet:              earth,
		Gravity:             1,
		RadiusKm:            6378.1,
		MassKg:              5.97e+24,
		OrbitKm:             1.496e+08,
		OrbitDays:           365,
		SurfacePressureBars: 1,
		Moons:               1,
		Rings:               false,
	},
	MARS: Planet{
		planet:              mars,
		Gravity:             0.377,
		RadiusKm:            3389.5,
		MassKg:              6.42e+23,
		OrbitKm:             2.279e+08,
		OrbitDays:           687,
		SurfacePressureBars: 0.01,
		Moons:               2,
		Rings:               false,
	},
	JUPITER: Planet{
		planet:              jupiter,
		Gravity:             2.36,
		RadiusKm:            69911,
		MassKg:              1.9e+27,
		OrbitKm:             7.786e+08,
		OrbitDays:           4333,
		SurfacePressureBars: 20,
		Moons:               4,
		Rings:               true,
	},
	SATURN: Planet{
		planet:              saturn,
		Gravity:             0.916,
		RadiusKm:            58232,
		MassKg:              5.68e+26,
		OrbitKm:             1.4335e+09,
		OrbitDays:           10759,
		SurfacePressureBars: 1,
		Moons:               7,
		Rings:               true,
	},
	URANUS: Planet{
		planet:              uranus,
		Gravity:             0.889,
		RadiusKm:            25362,
		MassKg:              8.68e+25,
		OrbitKm:             2.8725e+09,
		OrbitDays:           30687,
		SurfacePressureBars: 1.3,
		Moons:               13,
		Rings:               true,
	},
	NEPTUNE: Planet{
		planet:              neptune,
		Gravity:             1.12,
		RadiusKm:            24622,
		MassKg:              1.02e+26,
		OrbitKm:             4.4951e+09,
		OrbitDays:           60190,
		SurfacePressureBars: 1.5,
		Moons:               2,
		Rings:               true,
	},
}

// invalidPlanet is an invalid sentinel value for Planet
var invalidPlanet = Planet{}

// allSlice returns a slice of all enum values.
// This method is useful for iterating over all enum values in a loop.
func (p planetsContainer) allSlice() []Planet {
	return []Planet{
		Planets.MERCURY,
		Planets.VENUS,
		Planets.EARTH,
		Planets.MARS,
		Planets.JUPITER,
		Planets.SATURN,
		Planets.URANUS,
		Planets.NEPTUNE,
	}
}

// All returns an iterator over all enum values.
// This method is useful for iterating over all enum values in a loop.
func (p planetsContainer) All() iter.Seq[Planet] {
	return func(yield func(Planet) bool) {
		for _, v := range p.allSlice() {
			if !yield(v) {
				return
			}
		}
	}
}

// ParsePlanet parses the input value into an enum value.
// It returns the parsed enum value or an error if the input is invalid.
// It is a convenience function that can be used to parse enum values from
// various input types, such as strings, byte slices, or other enum types.
func ParsePlanet(input any) (Planet, error) {
	var res = invalidPlanet
	switch v := input.(type) {
	case Planet:
		return v, nil
	case string:
		res = stringToPlanet(v)
	case fmt.Stringer:
		res = stringToPlanet(v.String())
	case []byte:
		res = stringToPlanet(string(v))
	case int:
		res = numberToPlanet(v)
	case int8:
		res = numberToPlanet(v)
	case int16:
		res = numberToPlanet(v)
	case int32:
		res = numberToPlanet(v)
	case int64:
		res = numberToPlanet(v)
	case uint:
		res = numberToPlanet(v)
	case uint8:
		res = numberToPlanet(v)
	case uint16:
		res = numberToPlanet(v)
	case uint32:
		res = numberToPlanet(v)
	case uint64:
		res = numberToPlanet(v)
	case float32:
		res = numberToPlanet(v)
	case float64:
		res = numberToPlanet(v)
	default:
		return res, fmt.Errorf("invalid type %T", input)
	}
	return res, nil
}

// planetsNameMap is a map of enum values to their Planet representation
// It is used to convert string representations of enum values into their Planet representation.
var planetsNameMap = map[string]Planet{
	"Mercury": Planets.MERCURY,
	"Venus":   Planets.VENUS,
	"Earth":   Planets.EARTH,
	"Mars":    Planets.MARS,
	"Jupiter": Planets.JUPITER,
	"Saturn":  Planets.SATURN,
	"Uranus":  Planets.URANUS,
	"Neptune": Planets.NEPTUNE,
}

// stringToPlanet converts a string representation of an enum value into its Planet representation
// It returns the Planet representation of the enum value if the string is valid
// Otherwise, it returns invalidPlanet
func stringToPlanet(s string) Planet {
	if t, ok := planetsNameMap[s]; ok {
		return t
	}
	return invalidPlanet
}

// numberToPlanet converts a numeric value to a Planet
// It returns the Planet representation of the enum value if the numeric value is valid
// Otherwise, it returns invalidPlanet
func numberToPlanet[T constraints.Integer | constraints.Float](num T) Planet {
	f := float64(num)
	if math.Floor(f) != f {
		return invalidPlanet
	}
	i := int(f)
	if i <= 0 || i > len(Planets.allSlice()) {
		return invalidPlanet
	}
	return Planets.allSlice()[i]
}

// ExhaustivePlanets iterates over all enum values and calls the provided function for each value.
// This function is useful for performing operations on all valid enum values in a loop.
func ExhaustivePlanets(f func(Planet)) {
	for _, p := range Planets.allSlice() {
		f(p)
	}
}

// validPlanets is a map of enum values to their validity
var validPlanets = map[Planet]bool{
	Planets.MERCURY: true,
	Planets.VENUS:   true,
	Planets.EARTH:   true,
	Planets.MARS:    true,
	Planets.JUPITER: true,
	Planets.SATURN:  true,
	Planets.URANUS:  true,
	Planets.NEPTUNE: true,
}

// IsValid checks whether the Planets value is valid.
// A valid value is one that is defined in the original enum and not marked as invalid.
func (p Planet) IsValid() bool {
	return validPlanets[p]
}

// MarshalJSON implements the json.Marshaler interface for Planet.
// It returns the JSON representation of the enum value as a byte slice.
func (p Planet) MarshalJSON() ([]byte, error) {
	return []byte("\"" + p.String() + "\""), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for Planet.
// It parses the JSON representation of the enum value from the byte slice.
// It returns an error if the input is not a valid JSON representation.
func (p *Planet) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, "\""), "\"")
	newp, err := ParsePlanet(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface for Planet.
// It returns the string representation of the enum value as a byte slice
func (p Planet) MarshalText() ([]byte, error) {
	return []byte("\"" + p.String() + "\""), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for Planet.
// It parses the string representation of the enum value from the byte slice.
// It returns an error if the byte slice does not contain a valid enum value.
func (p *Planet) UnmarshalText(b []byte) error {
	newp, err := ParsePlanet(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

// Scan implements the database/sql.Scanner interface for Planet.
// It parses the string representation of the enum value from the database row.
// It returns an error if the row does not contain a valid enum value.
func (p *Planet) Scan(value any) error {
	newp, err := ParsePlanet(value)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

// Value implements the database/sql/driver.Valuer interface for Planet.
// It returns the string representation of the enum value.
func (p Planet) Value() (driver.Value, error) {
	return p.String(), nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface for Planet.
// It returns the binary representation of the enum value as a byte slice.
func (p Planet) MarshalBinary() ([]byte, error) {
	return []byte("\"" + p.String() + "\""), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for Planet.
// It parses the binary representation of the enum value from the byte slice.
// It returns an error if the byte slice does not contain a valid enum value.
func (p *Planet) UnmarshalBinary(b []byte) error {
	newp, err := ParsePlanet(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}


// MarshalYAML implements the yaml.Marshaler interface for Planet.
// It returns the string representation of the enum value.
func (p Planet) MarshalYAML() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Planet.
// It parses the byte slice representation of the enum value and returns an error
// if the YAML byte slice does not contain a valid enum value.
func (p *Planet) UnmarshalYAML(b []byte) error {
	newp, err := ParsePlanet(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

// planetNames is a constant string slice containing all enum values cononical absolute names
const planetNames = "MercuryVenusEarthMarsJupiterSaturnUranusNeptune"

// planetNamesMap is a map of enum values to their canonical absolute
// name positions within the planetNames string slice
var planetNamesMap = map[Planet]string{
	Planets.MERCURY: planetNames[0:7],
	Planets.VENUS:   planetNames[7:12],
	Planets.EARTH:   planetNames[12:17],
	Planets.MARS:    planetNames[17:21],
	Planets.JUPITER: planetNames[21:28],
	Planets.SATURN:  planetNames[28:34],
	Planets.URANUS:  planetNames[34:40],
	Planets.NEPTUNE: planetNames[40:47],
}

// String implements the Stringer interface.
// It returns the canonical absolute name of the enum value.
func (p Planet) String() string {
	if str, ok := planetNamesMap[p]; ok {
		return str
	}
	return fmt.Sprintf("planet(%d)", p.planet)
}

// Compile-time check that all enum values are valid.
// This function is used to ensure that all enum values are defined and valid.
// It is called by the compiler to verify that the enum values are valid.
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the goenums command to generate them again.
	// Does not identify newly added constant values unless order changes
	var x [9]struct{}
	_ = x[unknown-0]
	_ = x[mercury-1]
	_ = x[venus-2]
	_ = x[earth-3]
	_ = x[mars-4]
	_ = x[jupiter-5]
	_ = x[saturn-6]
	_ = x[uranus-7]
	_ = x[neptune-8]
}
```

For more examples, see those used for testing in the [testdata](https://github.com/zarldev/goenums/tree/main/internal/testdata) directory.


### Mentions
[![go-recipes](https://raw.githubusercontent.com/nikolaydubina/go-recipes/main/badge.svg?raw=true)](https://github.com/nikolaydubina/go-recipes)

## License

MIT License - See [LICENSE](https://github.com/zarldev/goenums/blob/main/LICENSE) for full details.