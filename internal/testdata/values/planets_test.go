package solarsystemsimple_test

import (
	"testing"

	solarsystemsimple "github.com/zarldev/goenums/internal/testdata/values"
)

func TestPlanetsIteration(t *testing.T) {
	// Test that iteration only includes valid planets (no "unknown")
	var collected []solarsystemsimple.Planet
	for planet := range solarsystemsimple.Planets.All() {
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
}

func TestPlanetsStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected solarsystemsimple.Planet
		hasError bool
	}{
		// Valid planets should parse successfully
		{"Mercury", solarsystemsimple.Planets.MERCURY, false},
		{"Venus", solarsystemsimple.Planets.VENUS, false},
		{"Earth", solarsystemsimple.Planets.EARTH, false},
		{"Mars", solarsystemsimple.Planets.MARS, false},
		{"Jupiter", solarsystemsimple.Planets.JUPITER, false},
		{"Saturn", solarsystemsimple.Planets.SATURN, false},
		{"Uranus", solarsystemsimple.Planets.URANUS, false},
		{"Neptune", solarsystemsimple.Planets.NEPTUNE, false},

		// Invalid planet should also parse (but will be marked as invalid)
		{"unknown", solarsystemsimple.Planets.UNKNOWN, false},

		// Non-existent should fail
		{"InvalidPlanet", solarsystemsimple.Planet{}, true},
	}

	for _, test := range tests {
		got, err := solarsystemsimple.ParsePlanet(test.input)
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
	// Invalid planet should return false
	if solarsystemsimple.Planets.UNKNOWN.IsValid() {
		t.Error("UNKNOWN planet should be invalid")
	}

	// Valid planets should return true
	validPlanets := []solarsystemsimple.Planet{
		solarsystemsimple.Planets.MERCURY, solarsystemsimple.Planets.VENUS, solarsystemsimple.Planets.EARTH, solarsystemsimple.Planets.MARS,
		solarsystemsimple.Planets.JUPITER, solarsystemsimple.Planets.SATURN, solarsystemsimple.Planets.URANUS, solarsystemsimple.Planets.NEPTUNE,
	}

	for _, planet := range validPlanets {
		if !planet.IsValid() {
			t.Errorf("Planet %v should be valid", planet)
		}
	}
}
