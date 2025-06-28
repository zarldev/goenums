package tickets_test

import (
	"testing"

	tickets "github.com/zarldev/goenums/internal/testdata/quotes"
)

func TestTicketsIteration(t *testing.T) {
	// Test that iteration only includes valid tickets (no "unknown")
	var collected []tickets.Ticket
	for ticket := range tickets.Tickets.All() {
		collected = append(collected, ticket)
	}

	// Should be 4 valid tickets (created, pending, approval_pending, approval_accepted are valid)
	if len(collected) != 4 {
		t.Errorf("Expected 4 valid tickets in iteration, got %d", len(collected))
	}

	// Verify no invalid tickets in iteration
	for _, ticket := range collected {
		if !ticket.IsValid() {
			t.Errorf("Invalid ticket %v found in iteration", ticket)
		}
	}

	// Verify expected valid tickets are present
	expected := []tickets.Ticket{tickets.Tickets.CREATED, tickets.Tickets.PENDING, 
		tickets.Tickets.APPROVAL_PENDING, tickets.Tickets.APPROVAL_ACCEPTED}
	for i, ticket := range collected {
		if ticket != expected[i] {
			t.Errorf("Iterator[%d]: expected %v, got %v", i, expected[i], ticket)
		}
	}
}

func TestTicketsStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected tickets.Ticket
		hasError bool
	}{
		// Valid tickets should parse successfully
		{"Created Successfully", tickets.Tickets.CREATED, false},
		{"In Progress", tickets.Tickets.PENDING, false},
		{"Pending Approval", tickets.Tickets.APPROVAL_PENDING, false},
		{"Fully Approved", tickets.Tickets.APPROVAL_ACCEPTED, false},
		{"Has Been Rejected", tickets.Tickets.REJECTED, false},
		{"Successfully Completed", tickets.Tickets.COMPLETED, false},
		
		// Invalid ticket should also parse (but will be marked as invalid)
		{"Not Found", tickets.Tickets.UNKNOWN, false},
		
		// Non-existent should fail
		{"InvalidTicket", tickets.Ticket{}, true},
	}

	for _, test := range tests {
		got, err := tickets.ParseTicket(test.input)
		hasError := err != nil

		if hasError != test.hasError {
			t.Errorf("ParseTicket(%q): expected error=%v, got error=%v", test.input, test.hasError, hasError)
			continue
		}

		if !hasError && got != test.expected {
			t.Errorf("ParseTicket(%q): expected %v, got %v", test.input, test.expected, got)
		}
	}
}

func TestTicketsValidity(t *testing.T) {
	// Valid tickets should return true
	validTickets := []tickets.Ticket{tickets.Tickets.CREATED, tickets.Tickets.PENDING,
		tickets.Tickets.APPROVAL_PENDING, tickets.Tickets.APPROVAL_ACCEPTED}
	
	for _, ticket := range validTickets {
		if !ticket.IsValid() {
			t.Errorf("Ticket %v should be valid", ticket)
		}
	}

	// Invalid ticket should return false
	if tickets.Tickets.UNKNOWN.IsValid() {
		t.Error("UNKNOWN ticket should be invalid")
	}
	
	// Also check that rejected and completed are invalid
	if tickets.Tickets.REJECTED.IsValid() {
		t.Error("REJECTED ticket should be invalid")
	}
	if tickets.Tickets.COMPLETED.IsValid() {
		t.Error("COMPLETED ticket should be invalid")
	}
}