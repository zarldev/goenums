package accessorstyle

import "testing"

func TestAccessorStyleGo(t *testing.T) {
	t.Parallel()
	values := []Status{
		Statuses.PendingApproval,
		Statuses.APIError,
		Statuses.HTTP2Enabled,
		Statuses.UserID,
	}
	want := []string{
		"pendingApproval",
		"apiError",
		"http2Enabled",
		"userID",
	}
	for i, value := range values {
		if got := value.String(); got != want[i] {
			t.Fatalf("value %d String() = %q, want %q", i, got, want[i])
		}
		parsed, err := ParseStatus(want[i])
		if err != nil {
			t.Fatalf("ParseStatus(%q): %v", want[i], err)
		}
		if parsed != value {
			t.Fatalf("ParseStatus(%q) = %v, want %v", want[i], parsed, value)
		}
	}
}
