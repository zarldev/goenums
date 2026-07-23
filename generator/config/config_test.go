package config_test

import (
	"testing"

	"github.com/zarldev/goenums/generator/config"
)

func TestAccessorStyle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		style      config.AccessorStyle
		wantNormal config.AccessorStyle
		wantValid  bool
	}{
		{
			name:       "zero value normalizes to upper",
			style:      "",
			wantNormal: config.AccessorStyleUpper,
			wantValid:  false,
		},
		{
			name:       "upper",
			style:      config.AccessorStyleUpper,
			wantNormal: config.AccessorStyleUpper,
			wantValid:  true,
		},
		{
			name:       "go",
			style:      config.AccessorStyleGo,
			wantNormal: config.AccessorStyleGo,
			wantValid:  true,
		},
		{
			name:       "unknown",
			style:      "camel",
			wantNormal: "camel",
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.style.Normalized(); got != tt.wantNormal {
				t.Fatalf("Normalized() = %q, want %q", got, tt.wantNormal)
			}
			if got := tt.style.Valid(); got != tt.wantValid {
				t.Fatalf("Valid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}
