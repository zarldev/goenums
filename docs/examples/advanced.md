---
layout: default
title: Advanced Examples
---

Explore more powerful features and use cases for goenums.

## Extended Enum Types with Custom Fields

Add custom fields to your enums using type comments:

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
    mars                  // Mars 0.377,3389.5,6.42e23,227900000
    // ...
)
```

After generation, we can use the extended enum type:

```go
earthWeight := 100.0
fmt.Printf("Weight on %s: %.2f kg\n", 
    solarsystem.Planets.MARS, 
    earthWeight * solarsystem.Planets.MARS.Gravity)

// Get the radius of Earth
fmt.Printf("Earth radius: %.1f km\n", solarsystem.Planets.EARTH.RadiusKm)
```

## JSON, Text, Binary, YAML, and Database Storage

The generated enum type implements the:

* `json.Marshaler` and `json.Unmarshaler` interfaces
* `sql.Scanner` and `sql.Valuer` interfaces
* `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` interfaces
* `encoding.TextMarshaler` and `encoding.TextUnmarshaler` interfaces
* `yaml.Marshaler` and `yaml.Unmarshaler` interfaces

These interfaces allow you to use the enum type in JSON, text, binary, YAML, and database storage seamlessly.

```go
// JSON example
type Mission struct {
    ID          int                `json:"id"`
    Destination solarsystem.Planet `json:"destination"`
}

mission := Mission{
    ID:          1,
    Destination: solarsystem.Planets.MARS,
}

// Marshals to {"id":1,"destination":"Mars"}
jsonData, _ := json.Marshal(mission)

// Database example
func SaveMission(db *sql.DB, mission Mission) error {
    _, err := db.Exec("INSERT INTO missions (id, destination) VALUES (?, ?)", 
        mission.ID, mission.Destination)
    return err
}

// The Planet type will be stored as "Mars" in the database
```

## Exhaustive Handling

Use exhaustive handling function to ensure you handle all enum values with the generated `Exhaustive` function:

```go
// Process all enum values safely
// This is especially useful in tests to ensure all enum values are covered
validation.ExhaustiveTicketStatuses(func(status validation.TicketStatus) {
    // Process each status
    switch status {
    case validation.TicketStatuses.FAILED:
        handleFailed()
    case validation.TicketStatuses.PASSED:
        handlePassed()
    // ...
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
## Failfast Mode / Strict Mode

Enable strict validation of enum values with the failfast flag:

```go
//go:generate goenums -f status.go
```

Generated code will return errors for invalid values:

```go
status, err := validation.ParseStatus("INVALID_STATUS")
if err != nil {
    fmt.Println("error:", err)
}
```

## Legacy vs Modern Mode

Choose between modern Go 1.23+ iterator support and legacy iteration styles.

```go
// Modern iteration (Go 1.23+)
// Using Go 1.23+ range-over-function iteration
for status := range validation.Statuses.All() {
    fmt.Printf("Status: %s\n", status)
}

// Legacy iteration (or with -l flag)
// Using a slice of all enum values
for _, status := range validation.Statuses.All() {
    fmt.Printf("Status: %s\n", status)
}
```

## Constraints Mode

Generate local type constraints instead of importing external dependencies:

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

## Compile-time Validation

The generated code includes compile-time validation to ensure enum values remain consistent:

```go
// Compile-time check that all enum values are valid.
func _() {
    // An "invalid array index" compiler error signifies that the constant values have changed.
    // Re-run the goenums command to generate them again.
    var x [7]struct{}
    _ = x[unknown-0]
    _ = x[failed-1]
    _ = x[passed-2]
    // ... other enum values
}
```

This ensures that if you change the order or values of your enum constants, you'll get a compile error reminding you to regenerate the enum code.

[Back to Examples]({{ '/examples' | relative_url }})