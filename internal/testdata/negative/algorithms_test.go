package crypto_test

import (
	"testing"

	crypto "github.com/zarldev/goenums/internal/testdata/negative"
)

func TestAlgorithmsIteration(t *testing.T) {
	// Test that iteration only includes valid algorithms (no "None")
	var collected []crypto.Algorithm
	for alg := range crypto.Algorithms.All() {
		collected = append(collected, alg)
	}

	// Should be 2 valid algorithms (excluding None)
	if len(collected) != 2 {
		t.Errorf("Expected 2 valid algorithms in iteration, got %d", len(collected))
	}

	// Verify no invalid algorithms in iteration
	for _, alg := range collected {
		if !alg.IsValid() {
			t.Errorf("Invalid algorithm %v found in iteration", alg)
		}
	}

	// Verify all expected valid algorithms are present
	expected := []crypto.Algorithm{crypto.Algorithms.AES256, crypto.Algorithms.CHACHA20}
	for i, alg := range collected {
		if alg != expected[i] {
			t.Errorf("Iterator[%d]: expected %v, got %v", i, expected[i], alg)
		}
	}
}

func TestAlgorithmsStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected crypto.Algorithm
		hasError bool
	}{
		// Valid algorithms should parse successfully
		{"AES256", crypto.Algorithms.AES256, false},
		{"ChaCha20", crypto.Algorithms.CHACHA20, false},
		
		// Invalid algorithm should also parse (but will be marked as invalid)
		{"None", crypto.Algorithms.NONE, false},
		
		// Non-existent should fail
		{"InvalidAlgorithm", crypto.Algorithm{}, true},
	}

	for _, test := range tests {
		got, err := crypto.ParseAlgorithm(test.input)
		hasError := err != nil

		if hasError != test.hasError {
			t.Errorf("ParseAlgorithm(%q): expected error=%v, got error=%v", test.input, test.hasError, hasError)
			continue
		}

		if !hasError && got != test.expected {
			t.Errorf("ParseAlgorithm(%q): expected %v, got %v", test.input, test.expected, got)
		}
	}
}

func TestAlgorithmsValidity(t *testing.T) {
	// Valid algorithms should return true
	validAlgorithms := []crypto.Algorithm{crypto.Algorithms.AES256, crypto.Algorithms.CHACHA20}
	
	for _, alg := range validAlgorithms {
		if !alg.IsValid() {
			t.Errorf("Algorithm %v should be valid", alg)
		}
	}

	// Invalid algorithm should return false
	if crypto.Algorithms.NONE.IsValid() {
		t.Error("NONE algorithm should be invalid")
	}
}

func TestAlgorithmsStringConversion(t *testing.T) {
	tests := []struct {
		algorithm crypto.Algorithm
		expected  string
	}{
		{crypto.Algorithms.AES256, "AES256"},
		{crypto.Algorithms.CHACHA20, "ChaCha20"},
		{crypto.Algorithms.NONE, "None"},
	}

	for _, test := range tests {
		if got := test.algorithm.String(); got != test.expected {
			t.Errorf("Expected %v.String() = %q, got %q", test.algorithm, test.expected, got)
		}
	}
}

func TestAlgorithmsNumericParsing(t *testing.T) {
	tests := []struct {
		input    int
		expected crypto.Algorithm
		hasError bool
	}{
		// Numeric parsing based on valid algorithm indices (not enum values)
		{1, crypto.Algorithms.AES256, false},   // First valid algorithm
		{2, crypto.Algorithms.CHACHA20, false}, // Second valid algorithm
		
		// Out of bounds should fail
		{0, crypto.Algorithm{}, true},  // Invalid index
		{3, crypto.Algorithm{}, true},  // Out of bounds
		{999, crypto.Algorithm{}, true},
	}

	for _, test := range tests {
		got, err := crypto.ParseAlgorithm(test.input)
		hasError := err != nil

		if hasError != test.hasError {
			t.Errorf("ParseAlgorithm(%d): expected error=%v, got error=%v", test.input, test.hasError, hasError)
			continue
		}

		if !hasError && got != test.expected {
			t.Errorf("ParseAlgorithm(%d): expected %v, got %v", test.input, test.expected, got)
		}
	}
}