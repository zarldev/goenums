package discount_test

import (
	"testing"

	discount "github.com/zarldev/goenums/internal/testdata/values_only"
)

func TestDiscountTypesOnlyIteration(t *testing.T) {
	// Test that iteration includes all valid discount types
	var collected []discount.DiscountType
	for discountType := range discount.DiscountTypes.All() {
		collected = append(collected, discountType)
	}

	// Should include all 4 discount types
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
	expected := []discount.DiscountType{
		discount.DiscountTypes.SALE, discount.DiscountTypes.PERCENTAGE, 
		discount.DiscountTypes.AMOUNT, discount.DiscountTypes.GIVEAWAY,
	}
	for i, discountType := range collected {
		if discountType != expected[i] {
			t.Errorf("Iterator[%d]: expected %v, got %v", i, expected[i], discountType)
		}
	}
}

func TestDiscountTypesOnlyStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected discount.DiscountType
		hasError bool
	}{
		// Valid discount types should parse successfully
		{"sale", discount.DiscountTypes.SALE, false},
		{"percentage", discount.DiscountTypes.PERCENTAGE, false},
		{"amount", discount.DiscountTypes.AMOUNT, false},
		{"giveaway", discount.DiscountTypes.GIVEAWAY, false},
		
		// Non-existent should fail
		{"invalid", discount.DiscountType{}, true},
	}

	for _, test := range tests {
		got, err := discount.ParseDiscountType(test.input)
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

func TestDiscountTypesOnlyValidity(t *testing.T) {
	// All discount types should be valid (this is a values-only enum)
	validDiscountTypes := []discount.DiscountType{
		discount.DiscountTypes.SALE, discount.DiscountTypes.PERCENTAGE,
		discount.DiscountTypes.AMOUNT, discount.DiscountTypes.GIVEAWAY,
	}
	
	for _, discountType := range validDiscountTypes {
		if !discountType.IsValid() {
			t.Errorf("DiscountType %v should be valid", discountType)
		}
	}
}