package planets_test

import (
	"testing"

	planets "github.com/zarldev/goenums/internal/testdata/attributes"
)

func TestPlanetsIteration(t *testing.T) {
	// Test that iteration only includes valid planets (no "unknown")
	var collected []planets.Planet
	for planet := range planets.Planets.All() {
		collected = append(collected, planet)
	}

	// Should be 8 valid planets (excluding unknown)
	if len(collected) != 8 {
		t.Errorf("Expected 8 valid planets in iteration, got %d", len(collected))
	}

	// Verify no invalid planets in iteration
	for _, planet := range collected {
		if !planet.IsValid() {
			t.Errorf("Invalid planet %v found in iteration", planet)
		}
	}

	// Verify all expected valid planets are present
	expected := []planets.Planet{
		planets.Planets.MERCURY, planets.Planets.VENUS, planets.Planets.EARTH, planets.Planets.MARS,
		planets.Planets.JUPITER, planets.Planets.SATURN, planets.Planets.URANUS, planets.Planets.NEPTUNE,
	}
	for i, planet := range collected {
		if planet != expected[i] {
			t.Errorf("Iterator[%d]: expected %v, got %v", i, expected[i], planet)
		}
	}
}

func TestPlanetsStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected planets.Planet
		hasError bool
	}{
		// Valid planets should parse successfully
		{"Mercury", planets.Planets.MERCURY, false},
		{"Venus", planets.Planets.VENUS, false},
		{"Earth", planets.Planets.EARTH, false},
		{"Mars", planets.Planets.MARS, false},
		{"Jupiter", planets.Planets.JUPITER, false},
		{"Saturn", planets.Planets.SATURN, false},
		{"Uranus", planets.Planets.URANUS, false},
		{"Neptune", planets.Planets.NEPTUNE, false},
		// Non-existent should fail
		{"InvalidPlanet", planets.Planet{}, true},
	}

	for _, test := range tests {
		got, err := planets.ParsePlanet(test.input)
		hasError := err != nil

		if hasError != test.hasError {
			t.Errorf("ParsePlanet(%q): expected error=%v, got error=%v", test.input, test.hasError, hasError)
			continue
		}

		if !hasError && got != test.expected {
			t.Errorf("ParsePlanet(%q): expected %v, got %v", test.input, test.expected, got)
		}
	}
}

func TestPlanetsValidity(t *testing.T) {
	// Valid planets should return true
	validPlanets := []planets.Planet{
		planets.Planets.MERCURY, planets.Planets.VENUS, planets.Planets.EARTH, planets.Planets.MARS,
		planets.Planets.JUPITER, planets.Planets.SATURN, planets.Planets.URANUS, planets.Planets.NEPTUNE,
	}

	for _, planet := range validPlanets {
		if !planet.IsValid() {
			t.Errorf("Planet %v should be valid", planet)
		}
	}
}

func TestPlanetsStringConversion(t *testing.T) {
	tests := []struct {
		planet   planets.Planet
		expected string
	}{
		{planets.Planets.MERCURY, "Mercury"},
		{planets.Planets.VENUS, "Venus"},
		{planets.Planets.EARTH, "Earth"},
		{planets.Planets.MARS, "Mars"},
		{planets.Planets.JUPITER, "Jupiter"},
		{planets.Planets.SATURN, "Saturn"},
		{planets.Planets.URANUS, "Uranus"},
		{planets.Planets.NEPTUNE, "Neptune"},
	}

	for _, test := range tests {
		if got := test.planet.String(); got != test.expected {
			t.Errorf("Expected %v.String() = %q, got %q", test.planet, test.expected, got)
		}
	}
}
