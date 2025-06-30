package sale_test

import (
	"testing"

	sale "github.com/zarldev/goenums/internal/testdata/plural"
)

func TestDiscountTypesIteration(t *testing.T) {
	// Test that iteration includes all valid discount types
	var collected []sale.DiscountType
	for discountType := range sale.DiscountTypes.All() {
		collected = append(collected, discountType)
	}

	// Should include all valid discount types
	if len(collected) != 4 {
		t.Errorf("Expected 4 valid discount types in iteration, got %d", len(collected))
	}

	// Verify all discount types in iteration are valid
	for _, discountType := range collected {
		if !discountType.IsValid() {
			t.Errorf("Invalid discount type %v found in iteration", discountType)
		}
	}

	// Verify expected discount type sequence
	expected := []sale.DiscountType{
		sale.DiscountTypes.SALE, sale.DiscountTypes.PERCENTAGE,
		sale.DiscountTypes.AMOUNT, sale.DiscountTypes.GIVEAWAY,
	}
	for i, discountType := range collected {
		if discountType != expected[i] {
			t.Errorf("Iterator[%d]: expected %v, got %v", i, expected[i], discountType)
		}
	}
}

func TestDiscountTypesStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected sale.DiscountType
		hasError bool
	}{
		// Valid discount types should parse successfully
		{"sale", sale.DiscountTypes.SALE, false},
		{"percentage", sale.DiscountTypes.PERCENTAGE, false},
		{"amount", sale.DiscountTypes.AMOUNT, false},
		{"giveaway", sale.DiscountTypes.GIVEAWAY, false},

		// Non-existent should fail
		{"invalid", sale.DiscountType{}, true},
	}

	for _, test := range tests {
		got, err := sale.ParseDiscountType(test.input)
		hasError := err != nil

		if hasError != test.hasError {
			t.Errorf("ParseDiscountType(%q): expected error=%v, got error=%v", test.input, test.hasError, hasError)
			continue
		}

		if !hasError && got != test.expected {
			t.Errorf("ParseDiscountType(%q): expected %v, got %v", test.input, test.expected, got)
		}
	}
}

func TestDiscountTypesValidity(t *testing.T) {
	// All discount types should be valid
	validDiscountTypes := []sale.DiscountType{
		sale.DiscountTypes.SALE, sale.DiscountTypes.PERCENTAGE,
		sale.DiscountTypes.AMOUNT, sale.DiscountTypes.GIVEAWAY,
	}

	for _, discountType := range validDiscountTypes {
		if !discountType.IsValid() {
			t.Errorf("DiscountType %v should be valid", discountType)
		}
	}
}
