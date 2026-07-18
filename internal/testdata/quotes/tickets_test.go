package tickets_test

import (
	"encoding/json"
	"testing"

	tickets "github.com/zarldev/goenums/internal/testdata/quotes"
)

func TestTicketsIteration(t *testing.T) {
	// Test that iteration only includes valid tickets (no "unknown")
	var collected []tickets.Ticket
	for ticket := range tickets.Tickets.All() {
		collected = append(collected, ticket)
	}

	// Filter to only valid tickets for this test
	var validTickets []tickets.Ticket
	for _, ticket := range collected {
		if ticket.Validstate {
			validTickets = append(validTickets, ticket)
		}
	}

	// Should be 4 valid tickets
	if len(validTickets) != 4 {
		t.Errorf("Expected 4 valid tickets, got %d", len(validTickets))
	}

	// Verify expected valid tickets are present
	expected := []tickets.Ticket{tickets.Tickets.CREATED, tickets.Tickets.PENDING,
		tickets.Tickets.APPROVAL_PENDING, tickets.Tickets.APPROVAL_ACCEPTED}
	for i, ticket := range validTickets {
		if ticket != expected[i] {
			t.Errorf("ValidTicket[%d]: expected %v, got %v", i, expected[i], ticket)
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
	validTickets := []tickets.Ticket{
		tickets.Tickets.CREATED,
		tickets.Tickets.PENDING,
		tickets.Tickets.APPROVAL_PENDING,
		tickets.Tickets.APPROVAL_ACCEPTED,
		tickets.Tickets.COMPLETED,
		tickets.Tickets.REJECTED,
	}

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
	if tickets.Tickets.REJECTED.Validstate {
		t.Error("REJECTED ticket should be invalid")
	}
	if tickets.Tickets.COMPLETED.Validstate {
		t.Error("COMPLETED ticket should be invalid")
	}
}

func TestTicketSerializationContracts(t *testing.T) {
	t.Parallel()
	value := tickets.Tickets.APPROVAL_PENDING

	text, err := value.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(text), "Pending Approval"; got != want {
		t.Fatalf("MarshalText() = %q, want %q", got, want)
	}
	var fromText tickets.Ticket
	if err := fromText.UnmarshalText(text); err != nil {
		t.Fatal(err)
	}
	if fromText != value {
		t.Fatalf("text round trip = %v, want %v", fromText, value)
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `"Pending Approval"`; got != want {
		t.Fatalf("json.Marshal() = %s, want %s", got, want)
	}
	var fromJSON tickets.Ticket
	if err := json.Unmarshal(data, &fromJSON); err != nil {
		t.Fatal(err)
	}
	if fromJSON != value {
		t.Fatalf("JSON round trip = %v, want %v", fromJSON, value)
	}
	var fromRawJSON tickets.Ticket
	if err := fromRawJSON.UnmarshalJSON([]byte("Pending Approval")); err != nil {
		t.Fatal(err)
	}
	if fromRawJSON != value {
		t.Fatalf("raw UnmarshalJSON fallback = %v, want %v", fromRawJSON, value)
	}
	if err := json.Unmarshal([]byte(`123`), &fromJSON); err == nil {
		t.Fatal("UnmarshalJSON accepted a non-string value")
	}
	if err := json.Unmarshal([]byte(`"unterminated`), &fromJSON); err == nil {
		t.Fatal("UnmarshalJSON accepted malformed JSON")
	}
}
