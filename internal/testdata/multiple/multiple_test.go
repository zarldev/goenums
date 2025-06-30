package multipleenums_test

import (
	"testing"

	multipleenums "github.com/zarldev/goenums/internal/testdata/multiple"
)

func TestMultipleEnumsNotMixed(t *testing.T) {
	// Test that order enums only contain order constants
	var orderEnums []multipleenums.Order
	for e := range multipleenums.Orders.All() {
		orderEnums = append(orderEnums, e)
	}
	orderCount := 7 // created, approved, processing, readyToShip, shipped, delivered, cancelled

	if len(orderEnums) != orderCount {
		t.Errorf("Expected %d order enums, got %d", orderCount, len(orderEnums))
	}

	// Check that no status constants are in the order container
	for _, e := range orderEnums {
		name := e.String()
		statusNames := []string{"FAILED", "PASSED", "SKIPPED", "SCHEDULED", "RUNNING", "BOOKED"}
		for _, statusName := range statusNames {
			if name == statusName {
				t.Errorf("Status enum %q found in order enums", statusName)
			}
		}
	}

	// Test that status enums only contain status constants
	var statusEnums []multipleenums.Status
	for e := range multipleenums.Statuses.All() {
		statusEnums = append(statusEnums, e)
	}
	statusCount := 6 // failed, passed, skipped, scheduled, running, booked

	if len(statusEnums) != statusCount {
		t.Errorf("Expected %d status enums, got %d", statusCount, len(statusEnums))
	}

	// Check that no order constants are in the status container
	for _, e := range statusEnums {
		name := e.String()
		orderNames := []string{"CREATED", "APPROVED", "PROCESSING", "READY_TO_SHIP", "SHIPPED", "DELIVERED", "CANCELLED"}
		for _, orderName := range orderNames {
			if name == orderName {
				t.Errorf("Order enum %q found in status enums", orderName)
			}
		}
	}
}
