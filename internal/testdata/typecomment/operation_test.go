package typecomment_test

import (
	"testing"

	typecomment "github.com/zarldev/goenums/internal/testdata/typecomment"
)

// TestDescriptiveTypeCommentGenerates guards against regression of the bug
// where a non-field type comment (e.g. "//operations") was parsed as a field
// and produced an uncompilable "<nil>" struct member. The fact that this test
// package compiles at all is the primary assertion; the checks below confirm
// the enum is usable as a simple field-less enum.
func TestDescriptiveTypeCommentGenerates(t *testing.T) {
	var count int
	for range typecomment.OperationTypes.All() {
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 enum values, got %d", count)
	}

	got, err := typecomment.ParseOperationType("update")
	if err != nil {
		t.Fatalf("ParseOperationType(\"update\") error: %v", err)
	}
	if got != typecomment.OperationTypes.UPDATE {
		t.Errorf("ParseOperationType(\"update\") = %v, want UPDATE", got)
	}
}
