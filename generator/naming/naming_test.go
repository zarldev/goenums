package naming_test

import (
	"errors"
	"testing"

	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/naming"
)

func TestAccessorIdentifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		style  config.AccessorStyle
		want   string
	}{
		{name: "zero style uses upper", source: "pendingApproval", style: "", want: "PENDINGAPPROVAL"},
		{name: "upper", source: "pending_approval", style: config.AccessorStyleUpper, want: "PENDING_APPROVAL"},
		{name: "go camel", source: "pendingApproval", style: config.AccessorStyleGo, want: "PendingApproval"},
		{name: "go snake", source: "pending_approval", style: config.AccessorStyleGo, want: "PendingApproval"},
		{name: "go initialism", source: "api_error", style: config.AccessorStyleGo, want: "APIError"},
		{name: "go initialism camel", source: "apiError", style: config.AccessorStyleGo, want: "APIError"},
		{name: "go initialism uppercase", source: "API_ERROR", style: config.AccessorStyleGo, want: "APIError"},
		{name: "go digit boundary", source: "http2Enabled", style: config.AccessorStyleGo, want: "HTTP2Enabled"},
		{name: "go digit snake", source: "http2_enabled", style: config.AccessorStyleGo, want: "HTTP2Enabled"},
		{name: "go false initialism prefix", source: "idempotent", style: config.AccessorStyleGo, want: "Idempotent"},
		{name: "go unicode", source: "éclairStatus", style: config.AccessorStyleGo, want: "ÉclairStatus"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := naming.AccessorIdentifier(tt.source, tt.style)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("AccessorIdentifier(%q, %q) = %q, want %q", tt.source, tt.style, got, tt.want)
			}
		})
	}
}

func TestAccessorIdentifierErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		style  config.AccessorStyle
	}{
		{name: "invalid style", source: "pending", style: "camel"},
		{name: "blank identifier", source: "_", style: config.AccessorStyleGo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := naming.AccessorIdentifier(tt.source, tt.style)
			if !errors.Is(err, naming.ErrInvalidAccessorIdentifier) {
				t.Fatalf("error = %v, want ErrInvalidAccessorIdentifier", err)
			}
		})
	}
}

func TestValidateAccessors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		candidates []naming.AccessorCandidate
	}{
		{name: "duplicate", candidates: []naming.AccessorCandidate{{Source: "pendingApproval", Accessor: "PendingApproval"}, {Source: "pending_approval", Accessor: "PendingApproval"}}},
		{name: "reserved", candidates: []naming.AccessorCandidate{{Source: "all", Accessor: "All"}}},
		{name: "initialism duplicate", candidates: []naming.AccessorCandidate{{Source: "apiError", Accessor: "APIError"}, {Source: "api_error", Accessor: "APIError"}}},
		{name: "digit duplicate", candidates: []naming.AccessorCandidate{{Source: "http2Enabled", Accessor: "HTTP2Enabled"}, {Source: "http2_enabled", Accessor: "HTTP2Enabled"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := naming.ValidateAccessors("status", tt.candidates, config.AccessorStyleGo)
			if !errors.Is(err, naming.ErrAccessorCollision) {
				t.Fatalf("error = %v, want ErrAccessorCollision", err)
			}
		})
	}
}

func TestGeneratedNames(t *testing.T) {
	t.Parallel()
	if got := naming.WrapperName("statuses"); got != "Status" {
		t.Fatalf("WrapperName(statuses) = %q, want Status", got)
	}
	if got := naming.WrapperType("status"); got != "Status" {
		t.Fatalf("WrapperType(status) = %q, want Status", got)
	}
	if got := naming.ContainerType("status"); got != "statusesContainer" {
		t.Fatalf("ContainerType(status) = %q, want statusesContainer", got)
	}
	if got := naming.ContainerName("status"); got != "Statuses" {
		t.Fatalf("ContainerName(status) = %q, want Statuses", got)
	}
	if got := naming.MatcherName("status"); got != "StatusMatcher" {
		t.Fatalf("MatcherName(status) = %q, want StatusMatcher", got)
	}
	if got := naming.AccessorMethodName("apiError"); got != "APIError" {
		t.Fatalf("AccessorMethodName(apiError) = %q, want APIError", got)
	}
	if got := naming.Receiver("status"); got != "s" {
		t.Fatalf("Receiver(status) = %q, want s", got)
	}
	if got := naming.Receiver("HTTPStatus"); got != "h" {
		t.Fatalf("Receiver(HTTPStatus) = %q, want h", got)
	}
	if got := naming.Receiver("État"); got != "é" {
		t.Fatalf("Receiver(État) = %q, want é", got)
	}
	if got := naming.Receiver(""); got != "r" {
		t.Fatalf("Receiver(empty) = %q, want r", got)
	}
	if got := naming.Receiver("pkg.Status"); got != "pkg" {
		t.Fatalf("Receiver(pkg.Status) = %q, want pkg", got)
	}
}
