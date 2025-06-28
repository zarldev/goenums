package validation_test

import (
	"testing"

	validation "github.com/zarldev/goenums/internal/testdata/invalid"
)

func TestStatusIteration(t *testing.T) {
	// Test that iteration only includes valid statuses (no "invalid")
	var collected []validation.Status
	for status := range validation.Statuses.All() {
		collected = append(collected, status)
	}

	// Should be 5 valid statuses (excluding FAILED which is marked as invalid)
	if len(collected) != 5 {
		t.Errorf("Expected 5 valid statuses in iteration, got %d", len(collected))
	}

	// Verify no invalid statuses in iteration
	for _, status := range collected {
		if !status.IsValid() {
			t.Errorf("Invalid status %v found in iteration", status)
		}
	}

	// Verify expected valid statuses are present
	expected := []validation.Status{
		validation.Statuses.PASSED, validation.Statuses.SKIPPED, validation.Statuses.SCHEDULED, 
		validation.Statuses.RUNNING, validation.Statuses.BOOKED,
	}
	for i, status := range collected {
		if status != expected[i] {
			t.Errorf("Iterator[%d]: expected %v, got %v", i, expected[i], status)
		}
	}
}

func TestStatusStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected validation.Status
		hasError bool
	}{
		// Valid statuses should parse successfully
		{"passed", validation.Statuses.PASSED, false},
		{"skipped", validation.Statuses.SKIPPED, false},
		{"scheduled", validation.Statuses.SCHEDULED, false},
		{"running", validation.Statuses.RUNNING, false},
		{"booked", validation.Statuses.BOOKED, false},
		
		// Failed status should also parse (but will be marked as invalid)
		{"failed", validation.Statuses.FAILED, false},
		
		// Non-existent should fail
		{"NonExistent", validation.Status{}, true},
	}

	for _, test := range tests {
		got, err := validation.ParseStatus(test.input)
		hasError := err != nil

		if hasError != test.hasError {
			t.Errorf("ParseStatus(%q): expected error=%v, got error=%v", test.input, test.hasError, hasError)
			continue
		}

		if !hasError && got != test.expected {
			t.Errorf("ParseStatus(%q): expected %v, got %v", test.input, test.expected, got)
		}
	}
}

func TestStatusValidity(t *testing.T) {
	// Valid statuses should return true
	validStatuses := []validation.Status{
		validation.Statuses.PASSED, validation.Statuses.SKIPPED, validation.Statuses.SCHEDULED,
		validation.Statuses.RUNNING, validation.Statuses.BOOKED,
	}
	
	for _, status := range validStatuses {
		if !status.IsValid() {
			t.Errorf("Status %v should be valid", status)
		}
	}

	// Failed status should return false
	if validation.Statuses.FAILED.IsValid() {
		t.Error("FAILED status should be invalid")
	}
}