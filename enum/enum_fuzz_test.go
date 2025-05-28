package enum_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/zarldev/goenums/enum"
)

func FuzzParseValue_String(f *testing.F) {
	// Seed with some interesting cases
	f.Add("hello")
	f.Add("")
	f.Add("hello world")
	f.Add(`"quoted string"`)
	f.Add("string with\nnewlines")
	f.Add("unicode: 🚀")

	f.Fuzz(func(t *testing.T, input string) {
		_, err := enum.ParseValue(input, "")
		if err != nil && !errors.Is(err, enum.ErrParseValue) {
			t.Errorf("parse value(%q, \"\") returned error: %v", input, err)
		}
	})
}

func FuzzParseValue_Int(f *testing.F) {
	// Seed with edge cases
	f.Add("0")
	f.Add("42")
	f.Add("-42")
	f.Add("9223372036854775807")  // max int64
	f.Add("-9223372036854775808") // min int64
	f.Add("abc")                  // invalid
	f.Add("\U00045a2f")

	f.Fuzz(func(t *testing.T, input string) {
		_, err := enum.ParseValue(input, 0)
		if err != nil && !errors.Is(err, enum.ErrParseValue) {
			t.Errorf("parse value(%q, 0) returned error: %v", input, err)
		}
	})
}

func FuzzParseValue_Bool(f *testing.F) {
	// Seed with known cases
	f.Add("true")
	f.Add("false")
	f.Add("1")
	f.Add("0")
	f.Add("invalid")

	f.Fuzz(func(t *testing.T, input string) {
		_, err := enum.ParseValue(input, false)
		if err != nil && !errors.Is(err, enum.ErrParseValue) {
			t.Errorf("parse value(%q, false) returned error: %v", input, err)
		}
	})
}

func FuzzParseValue_Float64(f *testing.F) {
	// Seed with various float formats
	f.Add("3.14")
	f.Add("-3.14")
	f.Add("1.23e-4")
	f.Add("1.23E+10")
	f.Add("0.0")
	f.Add("invalid")

	f.Fuzz(func(t *testing.T, input string) {
		_, err := enum.ParseValue(input, 0.0)
		if err != nil && !errors.Is(err, enum.ErrParseValue) {
			t.Errorf("parse value(%q, 0.0) returned error: %v", input, err)
		}
	})
}

func FuzzParseValue_Duration(f *testing.F) {
	// Seed with time duration formats
	f.Add("1h")
	f.Add("30m")
	f.Add("45s")
	f.Add("1h30m45s")
	f.Add("invalid")

	f.Fuzz(func(t *testing.T, input string) {
		_, err := enum.ParseValue(input, time.Duration(0))
		if err != nil && !errors.Is(err, enum.ErrParseValue) {
			t.Errorf("parse value(%q, time.Duration(0)) returned error: %v", input, err)
		}
	})
}

func FuzzParseEnumAliases(f *testing.F) {
	// Seed with various alias formats
	f.Add("alias1")
	f.Add("alias1,alias2")
	f.Add(`"quoted","unquoted"`)
	f.Add("alias1,alias2,alias3")
	f.Add("")
	f.Add(",")
	f.Add(",,")
	f.Add(" spaced , aliases ")

	f.Fuzz(func(t *testing.T, input string) {
		got := enum.ParseEnumAliases(input)
		if got == nil {
			t.Error("ParseEnumAliases returned nil slice")
		}
	})
}

func FuzzParseEnumFields(f *testing.F) {
	// Create a consistent enumIota for testing
	enumIota := enum.EnumIota{
		Fields: []enum.Field{
			{Name: "Number", Value: 0},
			{Name: "Flag", Value: false},
			{Name: "Text", Value: ""},
			{Name: "Float", Value: 0.0},
		},
	}

	// Seed with various field formats
	f.Add("42,true,hello,3.14")
	f.Add("1,false,world,2.71")
	f.Add("")
	f.Add("42")
	f.Add("42,true")
	f.Add("invalid,true,hello,3.14")

	f.Fuzz(func(t *testing.T, input string) {
		got, err := enum.ParseEnumFields(input, enumIota)
		if err != nil {
			if !errors.Is(err, enum.ErrFieldEmptyValue) && !errors.Is(err, enum.ErrParseValue) {
				t.Errorf("ParseEnumFields returned unexpected error: %v", err)
				return
			}
		}

		// Should never return nil
		if got == nil {
			t.Error("ParseEnumFields returned nil slice")
			return
		}

		// Each field should have the correct type
		for i, field := range got {
			if i >= len(enumIota.Fields) {
				break
			}

			expectedType := reflect.TypeOf(enumIota.Fields[i].Value)
			actualType := reflect.TypeOf(field.Value)

			if expectedType != actualType {
				t.Errorf("Field %d: got type %v, want type %v", i, actualType, expectedType)
			}
		}
	})
}

func FuzzExtractFields(f *testing.F) {
	// Seed with various comment formats
	f.Add("Name string, Age int")
	f.Add("Name[string], Age[int]")
	f.Add("Name(string), Age(int)")
	f.Add("")
	f.Add("string")
	f.Add("Duration time.Duration")
	f.Add("invalid format")

	f.Fuzz(func(t *testing.T, comment string) {
		opener, closer, fields := enum.ExtractFields(comment)

		// Should never panic
		if fields == nil {
			t.Error("ExtractFields returned nil fields slice")
		}

		// Opener and closer should be consistent
		validOpeners := []string{" ", "[", "("}
		validClosers := []string{" ", "]", ")"}

		openerValid := false
		closerValid := false

		for _, valid := range validOpeners {
			if opener == valid {
				openerValid = true
				break
			}
		}

		for _, valid := range validClosers {
			if closer == valid {
				closerValid = true
				break
			}
		}

		if !openerValid {
			t.Errorf("Invalid opener: %q", opener)
		}

		if !closerValid {
			t.Errorf("Invalid closer: %q", closer)
		}

		// If we have brackets, they should match
		if opener == "[" && closer != "]" {
			t.Error("Mismatched brackets: opener '[' but closer is not ']'")
		}
		if opener == "(" && closer != ")" {
			t.Error("Mismatched parentheses: opener '(' but closer is not ')'")
		}
	})
}
