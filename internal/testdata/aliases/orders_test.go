package orders_test

import (
	"testing"

	orders "github.com/zarldev/goenums/internal/testdata/aliases"
)

func TestOrdersIteration(t *testing.T) {
	// Test that iteration includes all valid orders
	var collected []orders.Order
	for order := range orders.Orders.All() {
		collected = append(collected, order)
	}

	// Should include all valid orders
	if len(collected) != 7 {
		t.Errorf("Expected 7 valid orders in iteration, got %d", len(collected))
	}

	// Verify all orders in iteration are valid
	for _, order := range collected {
		if !order.IsValid() {
			t.Errorf("Invalid order %v found in iteration", order)
		}
	}

	// Verify expected order sequence
	expected := []orders.Order{
		orders.Orders.CREATED, orders.Orders.APPROVED, orders.Orders.PROCESSING, orders.Orders.READYTOSHIP, 
		orders.Orders.SHIPPED, orders.Orders.DELIVERED, orders.Orders.CANCELLED,
	}
	for i, order := range collected {
		if order != expected[i] {
			t.Errorf("Iterator[%d]: expected %v, got %v", i, expected[i], order)
		}
	}
}

func TestOrdersStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected orders.Order
		hasError bool
	}{
		// Valid orders should parse successfully
		{"CREATED", orders.Orders.CREATED, false},
		{"APPROVED", orders.Orders.APPROVED, false},
		{"PROCESSING", orders.Orders.PROCESSING, false},
		{"READY_TO_SHIP", orders.Orders.READYTOSHIP, false},
		{"SHIPPED", orders.Orders.SHIPPED, false},
		{"DELIVERED", orders.Orders.DELIVERED, false},
		{"CANCELLED", orders.Orders.CANCELLED, false},
		
		// Non-existent should fail
		{"InvalidOrder", orders.Order{}, true},
	}

	for _, test := range tests {
		got, err := orders.ParseOrder(test.input)
		hasError := err != nil

		if hasError != test.hasError {
			t.Errorf("ParseOrder(%q): expected error=%v, got error=%v", test.input, test.hasError, hasError)
			continue
		}

		if !hasError && got != test.expected {
			t.Errorf("ParseOrder(%q): expected %v, got %v", test.input, test.expected, got)
		}
	}
}

func TestOrdersValidity(t *testing.T) {
	// All orders should be valid
	validOrders := []orders.Order{
		orders.Orders.CREATED, orders.Orders.APPROVED, orders.Orders.PROCESSING, orders.Orders.READYTOSHIP,
		orders.Orders.SHIPPED, orders.Orders.DELIVERED, orders.Orders.CANCELLED,
	}
	
	for _, order := range validOrders {
		if !order.IsValid() {
			t.Errorf("Order %v should be valid", order)
		}
	}
}