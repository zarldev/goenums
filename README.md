# goenums

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![build](https://github.com/zarldev/goenums/actions/workflows/go.yml/badge.svg)

`goenums` addresses Go's lack of native enum support by generating comprehensive, type-safe enum implementations from simple constant declarations. Transform basic `iota` based constants into feature-rich enums with string conversion, validation, JSON handling, database integration, and more.

## Table of Contents
- [Documentation](#documentation)
- [Installation](#installation)
- [Key Features](#key-features)
- [Usage](#usage)
- [Features Expanded](#features-expanded)
  - [Custom String Representations](#custom-string-representations)
    - [Standard Name Comment](#standard-name-comment)
    - [Name Comment with spaces](#name-comment-with-spaces)
  - [Extended Enum Types with Custom Fields](#extended-enum-types-with-custom-fields)
  - [Case Insensitive String Parsing](#case-insensitive-string-parsing)
  - [JSON & Database Storage](#json--database-storage)
  - [Exhaustive Handling](#exhaustive-handling)
  - [Iterator Support (Go 1.21+)](#iterator-support-go-121)
  - [Failfast Mode / Strict Mode](#failfast-mode--strict-mode)
  - [Legacy Mode](#legacy-mode)
  - [Verbose Mode](#verbose-mode)
  - [Output Format](#output-format)
- [Getting Started](#getting-started)
  - [Basic Example](#basic-example)
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
 - Iteration: Modern Go 1.21+ iteration support with legacy fallback
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
  -help, -h
        Print help information
  -version, -v
        Print version information
  -failfast, -f
        Enable failfast mode - fail on generation of invalid enum while parsing (default: false)
  -legacy, -l
        Generate legacy code without Go 1.23+ iterator support (default: false)
  -insensitive, -i
        Generate case insensitive string parsing (default: false)
  -verbose, -vv
        Enable verbose mode - prints out the generated code (default: false)
  -output, -o string
        Specify the output format (default: go)
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

## Iterator Support (Go 1.21+)
By default, goenums generates modern iterator support using Go 1.23's range-over-func feature:

```go 
// Using Go 1.21+ iterator
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

## Output Format
You can specify the output format by using the `-output` flag. The default is `go`.


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
- Go 1.21+ for iterator support (or use -l flag for legacy mode)

# Examples

Input source go file:

```golang
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
// Code generated by goenums 'v0.3.6' at 2025-04-19T03:02:40+01:00. DO NOT EDIT.
// This file was generated by github.com/zarldev/goenums
// using the command:
// goenums planets.go

package solarsystem

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"iter"
	"strconv"
)

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

var Planets = planetsContainer{
	MERCURY: Planet{
		planet:              mercury,
		Gravity:             0.378000,
		RadiusKm:            2439.700000,
		MassKg:              330000000000000029360128.000000,
		OrbitKm:             57910000.000000,
		OrbitDays:           88.000000,
		SurfacePressureBars: 0.000000,
		Moons:               0,
		Rings:               false,
	},
	VENUS: Planet{
		planet:              venus,
		Gravity:             0.907000,
		RadiusKm:            6051.800000,
		MassKg:              4869999999999999601541120.000000,
		OrbitKm:             108200000.000000,
		OrbitDays:           225.000000,
		SurfacePressureBars: 92.000000,
		Moons:               0,
		Rings:               false,
	},
	EARTH: Planet{
		planet:              earth,
		Gravity:             1.000000,
		RadiusKm:            6378.100000,
		MassKg:              5970000000000000281018368.000000,
		OrbitKm:             149600000.000000,
		OrbitDays:           365.000000,
		SurfacePressureBars: 1.000000,
		Moons:               1,
		Rings:               false,
	},
	MARS: Planet{
		planet:              mars,
		Gravity:             0.377000,
		RadiusKm:            3389.500000,
		MassKg:              642000000000000046137344.000000,
		OrbitKm:             227900000.000000,
		OrbitDays:           687.000000,
		SurfacePressureBars: 0.010000,
		Moons:               2,
		Rings:               false,
	},
	JUPITER: Planet{
		planet:              jupiter,
		Gravity:             2.360000,
		RadiusKm:            69911.000000,
		MassKg:              1900000000000000107709726720.000000,
		OrbitKm:             778600000.000000,
		OrbitDays:           4333.000000,
		SurfacePressureBars: 20.000000,
		Moons:               4,
		Rings:               true,
	},
	SATURN: Planet{
		planet:              saturn,
		Gravity:             0.916000,
		RadiusKm:            58232.000000,
		MassKg:              568000000000000011945377792.000000,
		OrbitKm:             1433500000.000000,
		OrbitDays:           10759.000000,
		SurfacePressureBars: 1.000000,
		Moons:               7,
		Rings:               true,
	},
	URANUS: Planet{
		planet:              uranus,
		Gravity:             0.889000,
		RadiusKm:            25362.000000,
		MassKg:              86800000000000000905969664.000000,
		OrbitKm:             2872500000.000000,
		OrbitDays:           30687.000000,
		SurfacePressureBars: 1.300000,
		Moons:               13,
		Rings:               true,
	},
	NEPTUNE: Planet{
		planet:              neptune,
		Gravity:             1.120000,
		RadiusKm:            24622.000000,
		MassKg:              102000000000000007952400384.000000,
		OrbitKm:             4495100000.000000,
		OrbitDays:           60190.000000,
		SurfacePressureBars: 1.500000,
		Moons:               2,
		Rings:               true,
	},
}

// invalidPlanet represents an invalid or undefined Planet value.
// It is used as a default return value for failed parsing or conversion operations.
var invalidPlanet = Planet{}

// allSlice is an internal method that returns all valid Planet values as a slice.
func (c planetsContainer) allSlice() []Planet {
	return []Planet{
		c.MERCURY,
		c.VENUS,
		c.EARTH,
		c.MARS,
		c.JUPITER,
		c.SATURN,
		c.URANUS,
		c.NEPTUNE,
	}
}

// AllSlice returns all valid Planet values as a slice.
// Deprecated: Use All() with Go 1.21+ range over function types instead.
func (c planetsContainer) AllSlice() []Planet {
	return c.allSlice()
}

// All returns all valid Planet values.
// In Go 1.21+, this can be used with range-over-function iteration:
// ```
//
//	for v := range Planets.All() {
//	    // process each enum value
//	}
//
// ```
func (c planetsContainer) All() iter.Seq[Planet] {
	return func(yield func(Planet) bool) {
		for _, v := range c.allSlice() {
			if !yield(v) {
				return
			}
		}
	}
}

// ParsePlanet converts various input types to a Planet value.
// It accepts the following types:
// - Planet: returns the value directly
// - string: parses the string representation
// - []byte: converts to string and parses
// - fmt.Stringer: uses the String() result for parsing
// - int/int32/int64: converts the integer to the corresponding enum value
//
// If the input cannot be converted to a valid Planet value, it returns
// the invalidPlanet value without an error.
func ParsePlanet(a any) (Planet, error) {
	res := invalidPlanet
	switch v := a.(type) {
	case Planet:
		return v, nil
	case []byte:
		res = stringToPlanet(string(v))
	case string:
		res = stringToPlanet(v)
	case fmt.Stringer:
		res = stringToPlanet(v.String())
	case int:
		res = intToPlanet(v)
	case int64:
		res = intToPlanet(int(v))
	case int32:
		res = intToPlanet(int(v))
	}
	return res, nil
}

// stringToPlanet is an internal function that converts a string to a Planet value.
// It uses a predefined mapping of string representations to enum values.
var (
	_planetsNameMap = map[string]Planet{
		"unknown": Planets.UNKNOWN, // Primary alias
		"Mercury": Planets.MERCURY, // Primary alias
		"mercury": Planets.MERCURY, // Enum name
		"Venus":   Planets.VENUS,   // Primary alias
		"venus":   Planets.VENUS,   // Enum name
		"Earth":   Planets.EARTH,   // Primary alias
		"earth":   Planets.EARTH,   // Enum name
		"Mars":    Planets.MARS,    // Primary alias
		"mars":    Planets.MARS,    // Enum name
		"Jupiter": Planets.JUPITER, // Primary alias
		"jupiter": Planets.JUPITER, // Enum name
		"Saturn":  Planets.SATURN,  // Primary alias
		"saturn":  Planets.SATURN,  // Enum name
		"Uranus":  Planets.URANUS,  // Primary alias
		"uranus":  Planets.URANUS,  // Enum name
		"Neptune": Planets.NEPTUNE, // Primary alias
		"neptune": Planets.NEPTUNE, // Enum name
	}
)

func stringToPlanet(s string) Planet {
	if v, ok := _planetsNameMap[s]; ok {
		return v
	}
	return invalidPlanet
}

// intToPlanet converts an integer to a Planet value.
// The integer is treated as the ordinal position in the enum sequence.
// If the integer doesn't correspond to a valid enum value, invalidPlanet is returned.
func intToPlanet(i int) Planet {
	if i < 0 || i >= len(Planets.allSlice()) {
		return invalidPlanet
	}
	return Planets.allSlice()[i]
}

// ExhaustivePlanets calls the provided function once for each valid Planets value.
// This is useful for switch statement exhaustiveness checking and for processing all enum values.
// Example usage:
// ```
//
//	ExhaustivePlanets(func(x Planet) {
//	    switch x {
//	    case Planets.Neptune:
//	        // handle Neptune
//	    }
//	})
//
// ```
func ExhaustivePlanets(f func(Planet)) {
	for _, p := range Planets.allSlice() {
		f(p)
	}
}

// validPlanets is a map of valid Planet values.
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

// IsValid checks whether the Planet value is valid.
// A valid value is one that is defined in the original enum and not marked as invalid.
func (p Planet) IsValid() bool {
	return validPlanets[p]
}

// MarshalJSON implements the json.Marshaler interface for Planet.
// The enum value is encoded as its string representation.
func (p Planet) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for Planet.
// It supports unmarshaling from a string representation of the enum.
func (p *Planet) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, `"`), ` `)
	newp, err := ParsePlanet(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

// Scan implements the sql.Scanner interface for Planet.
// This allows Planet values to be scanned directly from database queries.
// It supports scanning from strings, []byte, or integers.
func (p *Planet) Scan(value any) error {
	newp, err := ParsePlanet(value)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}

// Value implements the driver.Valuer interface for Planet.
// This allows Planet values to be saved to databases.
// The value is stored as a string representation of the enum.
func (p Planet) Value() (driver.Value, error) {
	return p.String(), nil
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the goenums command to generate them again.
	// Does not identify newly added constant values unless order changes
	var x [1]struct{}
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

const _planets_name = "unknownMercuryVenusEarthMarsJupiterSaturnUranusNeptune"

var _planets_index = [...]uint16{0, 7, 14, 19, 24, 28, 35, 41, 47, 54}

// String returns the string representation of the Planet value.
// For valid values, it returns the name of the constant.
// For invalid values, it returns a string in the format "planets(N)",
// where N is the numeric value.
func (i planet) String() string {
	if i < 0 || i >= planet(len(_planets_index)-1) {
		return "planets(" + (strconv.FormatInt(int64(i), 10) + ")")
	}
	return _planets_name[_planets_index[i]:_planets_index[i+1]]
}
```

For more examples, see the [examples](https://github.com/zarldev/goenums/tree/main/examples) directory.


### Mentions
[![go-recipes](https://raw.githubusercontent.com/nikolaydubina/go-recipes/main/badge.svg?raw=true)](https://github.com/nikolaydubina/go-recipes)

## License

MIT License - See [LICENSE](https://github.com/zarldev/goenums/blob/main/LICENSE) for full details.