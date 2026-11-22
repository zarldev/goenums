package contracts_test

import (
	"errors"
	"testing"

	"github.com/zarldev/goenums/internal/testdata/contracts/lenientzero"
	"github.com/zarldev/goenums/internal/testdata/contracts/strictzero"
)

func TestBinarySparseValidZero(t *testing.T) {
	t.Run("strict", func(t *testing.T) {
		for _, tc := range []struct {
			input string
			want  strictzero.State
		}{
			{"below", strictzero.States.BELOW},
			{`"below"`, strictzero.States.BELOW},
			{"zero", strictzero.States.ZERO},
			{`"zero"`, strictzero.States.MINUSONE}, // Raw alias wins over legacy ZERO.
			{"minusone", strictzero.States.MINUSONE},
			{`"minusone"`, strictzero.States.MINUSONE},
		} {
			t.Run(tc.input, func(t *testing.T) {
				got := strictzero.States.MINUSONE
				if err := got.UnmarshalBinary([]byte(tc.input)); err != nil || got != tc.want {
					t.Errorf("UnmarshalBinary(%q) = %v, %v; want %v", tc.input, got, err, tc.want)
				}
			})
		}
		for _, input := range []string{"missing", `"missing"`, `"below`, `below"`, `""below""`, `"\u0062elow"`, ""} {
			t.Run(input, func(t *testing.T) {
				got := strictzero.States.BELOW
				if err := got.UnmarshalBinary([]byte(input)); !errors.Is(err, strictzero.ErrParseState) {
					t.Errorf("UnmarshalBinary(%q) = %v, want ErrParseState", input, err)
				}
				if got != strictzero.States.BELOW {
					t.Errorf("invalid decode changed receiver to %v", got)
				}
			})
		}
	})
	t.Run("lenient", func(t *testing.T) {
		for _, tc := range []struct {
			input string
			want  lenientzero.State
		}{
			{"below", lenientzero.States.BELOW},
			{`"below"`, lenientzero.States.BELOW},
			{"zero", lenientzero.States.ZERO},
			{`"zero"`, lenientzero.States.MINUSONE}, // Raw alias wins over legacy ZERO.
			{"minusone", lenientzero.States.MINUSONE},
			{`"minusone"`, lenientzero.States.MINUSONE},
		} {
			t.Run(tc.input, func(t *testing.T) {
				got := lenientzero.States.MINUSONE
				if err := got.UnmarshalBinary([]byte(tc.input)); err != nil || got != tc.want {
					t.Errorf("UnmarshalBinary(%q) = %v, %v; want %v", tc.input, got, err, tc.want)
				}
			})
		}
		// Keep the existing lenient fallback, even when its numeric value is valid ZERO.
		want, err := lenientzero.ParseState("missing")
		if err != nil {
			t.Fatal(err)
		}
		for _, input := range []string{"missing", `"missing"`, `"below`, `below"`, `""below""`, `"\u0062elow"`, ""} {
			t.Run(input, func(t *testing.T) {
				got := lenientzero.States.BELOW
				if err := got.UnmarshalBinary([]byte(input)); err != nil || got != want {
					t.Errorf("UnmarshalBinary(%q) = %v, %v; want existing fallback %v", input, got, err, want)
				}
			})
		}
	})
}
