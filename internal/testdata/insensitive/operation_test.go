package insensitive_test

import (
	"testing"

	insensitive "github.com/zarldev/goenums/internal/testdata/insensitive"
)

// TestCaseInsensitiveParsing guards against regression of issue #39:
// with -i, parsing must accept any casing of an enum name, even when the
// enum has no explicit string alias.
func TestCaseInsensitiveParsing(t *testing.T) {
	tests := []struct {
		input string
		want  insensitive.OperationType
	}{
		{"update", insensitive.OperationTypes.UPDATE},
		{"UPDATE", insensitive.OperationTypes.UPDATE},
		{"Update", insensitive.OperationTypes.UPDATE},
		{"uPdAtE", insensitive.OperationTypes.UPDATE},
		{"remove", insensitive.OperationTypes.REMOVE},
		{"REMOVE", insensitive.OperationTypes.REMOVE},
		{"Remove", insensitive.OperationTypes.REMOVE},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := insensitive.ParseOperationType(tt.input)
			if err != nil {
				t.Fatalf("ParseOperationType(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseOperationType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestCaseInsensitiveRejectsUnknown confirms case-insensitivity does not make
// unrelated strings parse successfully.
func TestCaseInsensitiveRejectsUnknown(t *testing.T) {
	if _, err := insensitive.ParseOperationType("nonsense"); err == nil {
		t.Error("ParseOperationType(\"nonsense\") = nil error, want error")
	}
}
