package contracts_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/zarldev/goenums/internal/testdata/contracts"
	crypto "github.com/zarldev/goenums/internal/testdata/negative"
)

func TestSignalEncodingRoundTrips(t *testing.T) {
	for signal := range contracts.Signals.All() {
		t.Run(signal.String(), func(t *testing.T) {
			for _, codec := range []struct {
				name      string
				marshal   func() ([]byte, error)
				unmarshal func(*contracts.Signal, []byte) error
			}{
				{"binary", signal.MarshalBinary, (*contracts.Signal).UnmarshalBinary},
				{"text", signal.MarshalText, (*contracts.Signal).UnmarshalText},
			} {
				t.Run(codec.name, func(t *testing.T) {
					data, err := codec.marshal()
					if err != nil {
						t.Fatal(err)
					}
					if string(data) != signal.String() {
						t.Errorf("canonical encoding = %q, want %q", data, signal.String())
					}
					var decoded contracts.Signal
					if err := codec.unmarshal(&decoded, data); err != nil {
						t.Fatal(err)
					}
					if decoded != signal {
						t.Errorf("round trip = %v, want %v", decoded, signal)
					}
				})
			}

			data, err := json.Marshal(signal)
			if err != nil {
				t.Fatal(err)
			}
			var decoded contracts.Signal
			if err := json.Unmarshal(data, &decoded); err != nil || decoded != signal {
				t.Fatalf("JSON round trip = %v, %v; want %v", decoded, err, signal)
			}

			original := map[contracts.Signal]int{signal: 1}
			data, err = json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var decodedMap map[contracts.Signal]int
			if err := json.Unmarshal(data, &decodedMap); err != nil {
				t.Fatal(err)
			}
			if len(decodedMap) != 1 || decodedMap[signal] != 1 {
				t.Errorf("JSON map-key round trip = %v, want %v", decodedMap, original)
			}
		})
	}
}

func TestBinaryLegacyDecoding(t *testing.T) {
	var strict contracts.Signal
	if err := strict.UnmarshalBinary([]byte(`"below"`)); err != nil || strict != contracts.Signals.BELOW {
		t.Errorf("legacy strict decoding = %v, %v", strict, err)
	}
	var lenient contracts.Mode
	if err := lenient.UnmarshalBinary([]byte(`"ready"`)); err != nil || lenient != contracts.Modes.READY {
		t.Errorf("legacy lenient decoding = %v, %v", lenient, err)
	}
	data, err := contracts.Modes.READY.MarshalBinary()
	if err != nil || string(data) != "ready" {
		t.Fatalf("lenient binary encoding = %q, %v", data, err)
	}
	if err := lenient.UnmarshalBinary(data); err != nil || lenient != contracts.Modes.READY {
		t.Errorf("lenient binary round trip = %v, %v", lenient, err)
	}
}

func TestInvalidParseIdentity(t *testing.T) {
	var invalid contracts.Signal // No declared zero value.
	if invalid.IsValid() {
		t.Fatal("zero Signal unexpectedly valid")
	}
	for _, tc := range []struct {
		name  string
		input any
	}{
		{"nil", nil},
		{"bool", true},
		{"struct", struct{}{}},
		{"uintptr", uintptr(0)},
		{"same typed", invalid},
		{"string", "missing"},
		{"stringer", invalidStringer{}},
		{"bytes", []byte("missing")},
		{"int", 0},
		{"int8", int8(0)},
		{"int16", int16(0)},
		{"int32", int32(0)},
		{"int64", int64(0)},
		{"uint", uint(0)},
		{"uint8", uint8(0)},
		{"uint16", uint16(0)},
		{"uint32", uint32(0)},
		{"uint64", uint64(0)},
		{"float32", float32(0)},
		{"float64", float64(0)},
		{"fraction", 7.5},
		{"NaN", math.NaN()},
		{"infinity", math.Inf(1)},
		{"sparse gap", -2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := contracts.ParseSignal(tc.input)
			if !errors.Is(err, contracts.ErrParseSignal) {
				t.Errorf("ParseSignal(%v) error = %v, want ErrParseSignal", tc.input, err)
			}
			if got.IsValid() {
				t.Errorf("ParseSignal(%v) returned valid %v", tc.input, got)
			}
		})
	}
}

type invalidStringer struct{}

func (invalidStringer) String() string { return "missing" }

var _ fmt.Stringer = invalidStringer{}

func TestInvalidBinaryDoesNotChangeReceiver(t *testing.T) {
	for _, input := range []string{"missing", `"missing"`, `"below`, `below"`, `""below""`, ""} {
		t.Run(input, func(t *testing.T) {
			got := contracts.Signals.ABOVE
			if err := got.UnmarshalBinary([]byte(input)); !errors.Is(err, contracts.ErrParseSignal) {
				t.Errorf("UnmarshalBinary(%q) = %v, want ErrParseSignal", input, err)
			}
			if got != contracts.Signals.ABOVE {
				t.Errorf("invalid decode changed receiver to %v", got)
			}
		})
	}
}

func TestParseDeclaredNumericValues(t *testing.T) {
	for _, tc := range []struct {
		input any
		want  contracts.Signal
	}{
		{-3, contracts.Signals.BELOW},
		{int8(-3), contracts.Signals.BELOW},
		{int16(-3), contracts.Signals.BELOW},
		{int32(-3), contracts.Signals.BELOW},
		{int64(-3), contracts.Signals.BELOW},
		{uint(7), contracts.Signals.ABOVE},
		{uint8(7), contracts.Signals.ABOVE},
		{uint16(7), contracts.Signals.ABOVE},
		{uint32(7), contracts.Signals.ABOVE},
		{uint64(7), contracts.Signals.ABOVE},
		{float32(-3), contracts.Signals.BELOW},
		{float64(42), contracts.Signals.DISTANT},
		{contracts.Signals.ABOVE, contracts.Signals.ABOVE},
	} {
		got, err := contracts.ParseSignal(tc.input)
		if err != nil || got != tc.want {
			t.Errorf("ParseSignal(%v) = %v, %v; want %v", tc.input, got, err, tc.want)
		}
	}
	var zero crypto.Algorithm // AES256 is deliberately valid at zero.
	for _, input := range []any{zero, 0} {
		got, err := crypto.ParseAlgorithm(input)
		if err != nil || got != crypto.Algorithms.AES256 || !got.IsValid() {
			t.Errorf("ParseAlgorithm(%v) = %v, %v; want valid AES256", input, got, err)
		}
	}
}

func TestInvalidDeclaredSameType(t *testing.T) {
	got, err := crypto.ParseAlgorithm(crypto.Algorithms.NONE)
	if !errors.Is(err, crypto.ErrParseAlgorithm) || got.IsValid() {
		t.Errorf("ParseAlgorithm(NONE) = %v, %v; want ErrParseAlgorithm", got, err)
	}
	// Declared invalid names remain parseable, matching the existing string contract.
	got, err = crypto.ParseAlgorithm("None")
	if err != nil || got != crypto.Algorithms.NONE {
		t.Errorf("ParseAlgorithm(None) = %v, %v; want NONE", got, err)
	}
}

func TestLenientInvalidSameType(t *testing.T) {
	var invalid contracts.Mode
	got, err := contracts.ParseMode(invalid)
	if err != nil || got.IsValid() {
		t.Fatalf("lenient invalid same-type parse = %v, %v", got, err)
	}
	// Match the existing lenient invalid-value result rather than returning the unchecked input.
	want, err := contracts.ParseMode("missing")
	if err != nil || got != want {
		t.Errorf("lenient invalid same-type parse = %v; want %v, %v", got, want, err)
	}
}
