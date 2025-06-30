package validationstrings_test

import (
	"testing"

	validationstrings "github.com/zarldev/goenums/internal/testdata/invalid_alias"
)

func TestStatusAliasIteration(t *testing.T) {
	// Test that iteration only includes valid statuses (no "invalid")
	var collected []validationstrings.Status
	for status := range validationstrings.Statuses.All() {
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
}

func TestStatusAliasStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected validationstrings.Status
		hasError bool
	}{
		// Valid statuses should parse successfully
		{"PASSED", validationstrings.Statuses.PASSED, false},
		{"SKIPPED", validationstrings.Statuses.SKIPPED, false},
		{"SCHEDULED", validationstrings.Statuses.SCHEDULED, false},
		{"RUNNING", validationstrings.Statuses.RUNNING, false},
		{"BOOKED", validationstrings.Statuses.BOOKED, false},

		// Failed status should also parse
		{"FAILED", validationstrings.Statuses.FAILED, false},

		// Non-existent should fail
		{"NonExistent", validationstrings.Status{}, true},
	}

	for _, test := range tests {
		got, err := validationstrings.ParseStatus(test.input)
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

func TestStatusAliasValidity(t *testing.T) {
	// Valid statuses should return true
	validStatuses := []validationstrings.Status{
		validationstrings.Statuses.PASSED, validationstrings.Statuses.SKIPPED, validationstrings.Statuses.SCHEDULED,
		validationstrings.Statuses.RUNNING, validationstrings.Statuses.BOOKED,
	}

	for _, status := range validStatuses {
		if !status.IsValid() {
			t.Errorf("Status %v should be valid", status)
		}
	}

	// Failed status should return false
	if validationstrings.Statuses.FAILED.IsValid() {
		t.Error("FAILED status should be invalid")
	}
}
