// Package examples demonstrates the usage of goenums with comprehensive examples.
//
// This file contains executable examples that showcase the key features and
// capabilities of the goenums tool, including basic enum generation,
// custom fields, JSON marshaling, and database integration.
package examples_test

import (
	"encoding/json"
	"fmt"

	"github.com/zarldev/goenums/examples/solarsystem"
)

// Example_basicEnum demonstrates the most basic usage of goenums.
// This shows how to define a simple enum and use the generated methods.
func Example_basicEnum() {
	// Using the Planet enum from the solarsystem package
	planet := solarsystem.Planets.EARTH // Index 3 (Earth)

	// String representation
	fmt.Println("Planet:", planet.String())

	// Parsing from string
	parsed, err := solarsystem.ParsePlanet("Mars")
	if err == nil {
		fmt.Println("Parsed:", parsed.String())
	}

	// Validation
	fmt.Println("Is valid:", planet.IsValid())

	// Output: Planet: Earth
	// Parsed: Mars
	// Is valid: true
}

// Example_enumWithFields demonstrates enums with custom fields.
// This shows how to access metadata associated with enum values.
func Example_enumWithFields() {
	// Using the Planet enum which has custom fields
	earth := solarsystem.Planets.EARTH

	fmt.Println("Planet:", earth.String())
	fmt.Println("Gravity:", earth.Gravity)
	fmt.Println("Radius (km):", earth.RadiusKm)
	fmt.Println("Mass (kg):", earth.MassKg)
	fmt.Println("Moons:", earth.Moons)
	fmt.Println("Has rings:", earth.Rings)

	// Output: Planet: Earth
	// Gravity: 1
	// Radius (km): 6378.1
	// Mass (kg): 5.97e+24
	// Moons: 1
	// Has rings: false
}

// Example_jsonMarshaling demonstrates JSON marshaling and unmarshaling.
// This shows how generated enums integrate with Go's JSON package.
func Example_jsonMarshaling() {
	// Create a struct with enum field
	type SpaceMission struct {
		Name        string             `json:"name"`
		Destination solarsystem.Planet `json:"destination"`
	}

	// Marshal to JSON
	mission := SpaceMission{
		Name:        "Mars Rover",
		Destination: solarsystem.Planets.MARS,
	}

	jsonData, err := json.Marshal(mission)
	if err != nil {
		fmt.Println("failed to marshal:", err)
		return
	}
	fmt.Println("JSON:", string(jsonData))

	// Unmarshal from JSON
	var parsed SpaceMission
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		fmt.Println("failed to unmarshal:", err)
		return
	}
	fmt.Println("Parsed mission:", parsed.Name, "to", parsed.Destination.String())

	// Output: JSON: {"name":"Mars Rover","destination":"Mars"}
	// Parsed mission: Mars Rover to Mars
}

// Example_iteratorSupport demonstrates Go 1.23+ iterator support.
// This shows how to iterate over enum values using modern Go features.
func Example_iteratorSupport() {
	fmt.Println("All planets:")

	// Iterate over all planets using the All() iterator
	for planet := range solarsystem.Planets.All() {
		fmt.Printf("- %s (Gravity: %.3f)\n", planet.String(), planet.Gravity)
	}

	// Output: All planets:
	// - Mercury (Gravity: 0.378)
	// - Venus (Gravity: 0.907)
	// - Earth (Gravity: 1.000)
	// - Mars (Gravity: 0.377)
	// - Jupiter (Gravity: 2.360)
	// - Saturn (Gravity: 0.916)
	// - Uranus (Gravity: 0.889)
	// - Neptune (Gravity: 1.120)
}

// Example_exhaustiveHandling demonstrates exhaustive processing.
// This shows how to ensure all enum values are handled.
func Example_exhaustiveHandling() {
	fmt.Println("Processing all planets:")

	// Use ExhaustivePlanets to ensure all values are processed
	solarsystem.ExhaustivePlanets(func(planet solarsystem.Planet) {
		category := "Rocky"
		if planet.Gravity > 1.5 {
			category = "Gas Giant"
		}
		fmt.Printf("%s: %s planet\n", planet.String(), category)
	})

	// Output: Processing all planets:
	// Mercury: Rocky planet
	// Venus: Rocky planet
	// Earth: Rocky planet
	// Mars: Rocky planet
	// Jupiter: Gas Giant planet
	// Saturn: Rocky planet
	// Uranus: Rocky planet
	// Neptune: Rocky planet
}

// Example_numericParsing demonstrates parsing enums from numeric values.
// This shows how enums can be created from their underlying numeric representation.
func Example_numericParsing() {
	// Parse from various numeric types
	planet1, err := solarsystem.ParsePlanet(3) // Earth (index 3)
	if err == nil {
		fmt.Println("From int:", planet1.String())
	}

	planet2, err := solarsystem.ParsePlanet(4.0) // Mars (index 4)
	if err == nil {
		fmt.Println("From float:", planet2.String())
	}

	// Invalid index - this returns invalidPlanet, not an error
	invalidPlanet, err := solarsystem.ParsePlanet(99)
	if err == nil && !invalidPlanet.IsValid() {
		fmt.Println("Invalid index produces error")
	}

	// Output: From int: Earth
	// From float: Mars
	// Invalid index produces error
}

// Example_databaseIntegration demonstrates database Scanner/Valuer interfaces.
// This shows how enums can be stored and retrieved from databases.
func Example_databaseIntegration() {
	planet := solarsystem.Planets.JUPITER

	// Value() method for storing in database
	value, err := planet.Value()
	if err == nil {
		fmt.Println("Database value:", value)
	}

	// Scan() method for reading from database
	var scanned solarsystem.Planet
	err = scanned.Scan("Saturn")
	if err == nil {
		fmt.Println("Scanned planet:", scanned.String())
		fmt.Println("Has rings:", scanned.Rings)
	}

	// Output: Database value: Jupiter
	// Scanned planet: Saturn
	// Has rings: true
}

// Example_textMarshaling demonstrates text marshaling and unmarshaling.
// This shows how enums work with text-based formats.
func Example_textMarshaling() {
	planet := solarsystem.Planets.VENUS

	// Marshal to text
	text, err := planet.MarshalText()
	if err == nil {
		fmt.Println("Text representation:", string(text))
	}

	// Unmarshal from text
	var unmarshaled solarsystem.Planet
	err = unmarshaled.UnmarshalText([]byte("Neptune"))
	if err == nil {
		fmt.Println("Unmarshaled:", unmarshaled.String())
		fmt.Println("Surface pressure:", unmarshaled.SurfacePressureBars, "bars")
	}

	// Output: Text representation: "Venus"
	// Unmarshaled: Neptune
	// Surface pressure: 1.5 bars
}

// Example_binaryMarshaling demonstrates binary marshaling and unmarshaling.
// This shows how enums can be serialized to binary formats.
func Example_binaryMarshaling() {
	planet := solarsystem.Planets.URANUS

	// Marshal to binary
	binary, err := planet.MarshalBinary()
	if err == nil {
		fmt.Println("Binary representation:", string(binary))
	}

	// Unmarshal from binary
	var unmarshaled solarsystem.Planet
	err = unmarshaled.UnmarshalBinary([]byte("Mercury"))
	if err == nil {
		fmt.Println("Unmarshaled:", unmarshaled.String())
		fmt.Println("Orbit days:", unmarshaled.OrbitDays)
	}

	// Output: Binary representation: "Uranus"
	// Unmarshaled: Mercury
	// Orbit days: 88
}
