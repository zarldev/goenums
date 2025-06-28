package skipvalues_test

import (
	"testing"

	skipvalues "github.com/zarldev/goenums/internal/testdata/skipvalues"
)

func TestVersionsIteration(t *testing.T) {
	// Test that iteration includes all valid versions (skipped values are excluded)
	var collected []skipvalues.Version
	for version := range skipvalues.Versions.All() {
		collected = append(collected, version)
	}

	// Should include 4 valid versions (V1, V3, V4, V7) - V2, V5, V6 are skipped
	if len(collected) != 4 {
		t.Errorf("Expected 4 valid versions in iteration, got %d", len(collected))
	}

	// Verify all versions in iteration are valid
	for _, version := range collected {
		if !version.IsValid() {
			t.Errorf("Invalid version %v found in iteration", version)
		}
	}

	// Verify expected version sequence (only non-skipped values)
	expected := []skipvalues.Version{skipvalues.Versions.V1, skipvalues.Versions.V3, skipvalues.Versions.V4, skipvalues.Versions.V7}
	for i, version := range collected {
		if version != expected[i] {
			t.Errorf("Iterator[%d]: expected %v, got %v", i, expected[i], version)
		}
	}
}

func TestVersionsStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected skipvalues.Version
		hasError bool
	}{
		// Valid versions should parse successfully
		{"V1", skipvalues.Versions.V1, false},
		{"V3", skipvalues.Versions.V3, false},
		{"V4", skipvalues.Versions.V4, false},
		{"V7", skipvalues.Versions.V7, false},
		
		// Skipped versions should also parse (but will be marked as invalid)
		// Note: V2, V5, V6 don't exist as they were skipped with _
		
		// Non-existent should fail
		{"V99", skipvalues.Version{}, true},
	}

	for _, test := range tests {
		got, err := skipvalues.ParseVersion(test.input)
		hasError := err != nil

		if hasError != test.hasError {
			t.Errorf("ParseVersion(%q): expected error=%v, got error=%v", test.input, test.hasError, hasError)
			continue
		}

		if !hasError && got != test.expected {
			t.Errorf("ParseVersion(%q): expected %v, got %v", test.input, test.expected, got)
		}
	}
}

func TestVersionsValidity(t *testing.T) {
	// Valid versions should return true
	validVersions := []skipvalues.Version{skipvalues.Versions.V1, skipvalues.Versions.V3, skipvalues.Versions.V4, skipvalues.Versions.V7}
	
	for _, version := range validVersions {
		if !version.IsValid() {
			t.Errorf("Version %v should be valid", version)
		}
	}

	// Note: Skipped versions (V2, V5, V6) don't exist as enum constants since they were skipped with _
}