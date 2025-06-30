package sale_test

import (
	"testing"
	"time"

	sale "github.com/zarldev/goenums/internal/testdata/time"
)

func TestSalesIteration(t *testing.T) {
	// Test that iteration includes all valid sales
	var collected []sale.Sale
	for s := range sale.Sales.All() {
		collected = append(collected, s)
	}

	// Should include all 4 sales
	if len(collected) != 4 {
		t.Errorf("Expected 4 valid sales in iteration, got %d", len(collected))
	}

	// Verify all sales in iteration are valid
	for _, s := range collected {
		if !s.IsValid() {
			t.Errorf("Invalid sale %v found in iteration", s)
		}
	}

	// Verify expected sale sequence
	expected := []sale.Sale{
		sale.Sales.SALES, sale.Sales.PERCENTAGE, sale.Sales.AMOUNT, sale.Sales.GIVEAWAY,
	}
	for i, s := range collected {
		if s != expected[i] {
			t.Errorf("Iterator[%d]: expected %v, got %v", i, expected[i], s)
		}
	}
}

func TestSalesStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected sale.Sale
		hasError bool
	}{
		// Valid sales should parse successfully
		{"sales", sale.Sales.SALES, false},
		{"percentage", sale.Sales.PERCENTAGE, false},
		{"amount", sale.Sales.AMOUNT, false},
		{"giveaway", sale.Sales.GIVEAWAY, false},

		// Non-existent should fail
		{"invalid", sale.Sale{}, true},
	}

	for _, test := range tests {
		got, err := sale.ParseSale(test.input)
		hasError := err != nil

		if hasError != test.hasError {
			t.Errorf("ParseSale(%q): expected error=%v, got error=%v", test.input, test.hasError, hasError)
			continue
		}

		if !hasError && got != test.expected {
			t.Errorf("ParseSale(%q): expected %v, got %v", test.input, test.expected, got)
		}
	}
}

func TestSalesValidity(t *testing.T) {
	// All sales should be valid
	validSales := []sale.Sale{
		sale.Sales.SALES, sale.Sales.PERCENTAGE, sale.Sales.AMOUNT, sale.Sales.GIVEAWAY,
	}

	for _, s := range validSales {
		if !s.IsValid() {
			t.Errorf("Sale %v should be valid", s)
		}
	}
}

func TestSalesTimeAttributes(t *testing.T) {
	// Test time.Duration attributes
	tests := []struct {
		sale     sale.Sale
		expected time.Duration
	}{
		{sale.Sales.SALES, time.Hour * 168},     // 168 hours
		{sale.Sales.PERCENTAGE, time.Hour * 24}, // 24 hours
		{sale.Sales.AMOUNT, time.Hour * 48},     // 48 hours
		{sale.Sales.GIVEAWAY, time.Minute * 30}, // 30 minutes
	}

	for _, test := range tests {
		if test.sale.Duration != test.expected {
			t.Errorf("Expected %v duration to be %v, got %v", test.sale, test.expected, test.sale.Duration)
		}
	}

	// Test boolean attributes for giveaway sale
	giveaway := sale.Sales.GIVEAWAY
	if !giveaway.Available {
		t.Error("GIVEAWAY sale should be available")
	}
	if !giveaway.Started {
		t.Error("GIVEAWAY sale should be started")
	}
	if giveaway.Finished {
		t.Error("GIVEAWAY sale should not be finished")
	}
	if giveaway.Cancelled {
		t.Error("GIVEAWAY sale should not be cancelled")
	}
}
