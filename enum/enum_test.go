package enum_test

import (
	"errors"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/zarldev/goenums/enum"
)

func TestParseValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		input      string
		defaultVal any
		want       any
		err        error
	}{
		// Boolean tests
		{"bool true", "true", false, true, nil},
		{"bool false", "false", false, false, nil},
		{"bool invalid", "maybe", false, false, enum.ErrParseValue},
		{"bool 1", "1", false, true, nil},
		{"bool 0", "0", false, false, nil},

		// Integer tests
		{"int valid", "42", 0, 42, nil},
		{"int negative", "-42", 0, -42, nil},
		{"int invalid", "not-a-number", 0, 0, enum.ErrParseValue},
		{"int64 valid", "9223372036854775807", int64(0), int64(9223372036854775807), nil},
		{"int64 negative", "-9223372036854775808", int64(0), int64(-9223372036854775808), nil},
		{"int32 valid", "2147483647", int32(0), int32(2147483647), nil},
		{"int32 negative", "-2147483648", int32(0), int32(-2147483648), nil},
		{"int16 valid", "32767", int16(0), int16(32767), nil},
		{"int16 negative", "-32768", int16(0), int16(-32768), nil},
		{"int8 valid", "127", int8(0), int8(127), nil},
		{"int8 negative", "-128", int8(0), int8(-128), nil},
		{"int overflow", "9223372036854775808", int64(0), int64(0), enum.ErrParseValue},
		{"int unicode", "\U00045a2f", int64(0), int64(0), enum.ErrParseValue},

		// Unsigned integer tests
		{"uint valid", "42", uint(0), uint(42), nil},
		{"uint64 valid", "18446744073709551615", uint64(0), uint64(18446744073709551615), nil},
		{"uint32 valid", "4294967295", uint32(0), uint32(4294967295), nil},
		{"uint16 valid", "65535", uint16(0), uint16(65535), nil},
		{"uint8 valid", "255", uint8(0), uint8(255), nil},
		{"uint negative", "-1", uint(0), uint(0), enum.ErrParseValue},
		{"uint overflow", "18446744073709551616", uint64(0), uint64(0), enum.ErrParseValue},
		{"uint unicode", "\U00045a2f", uint64(0), uint64(0), enum.ErrParseValue},

		// Float tests
		{"float64 valid", "3.14", 0.0, 3.14, nil},
		{"float64 scientific", "1.23e-4", 0.0, 1.23e-4, nil},
		{"float64 invalid", "not-a-float", 0.0, 0.0, enum.ErrParseValue},
		{"float32 valid", "3.14", float32(0), float32(3.14), nil},
		{"float32 invalid", "not-a-float", float32(0), float32(0), enum.ErrParseValue},

		// String tests
		{"string quoted", `"hello"`, "", "hello", nil},
		{"string unquoted", "hello", "", "hello", nil},
		{"string empty quoted", `""`, "", "", nil},
		{"string empty", "", "", "", nil},
		{"string with spaces", "hello world", "", "hello world", nil},

		// Time tests
		{"time.Duration valid", "1h30m", time.Duration(0), 90 * time.Minute, nil},
		{"time.Duration seconds", "45s", time.Duration(0), 45 * time.Second, nil},
		{"time.Duration invalid", "invalid-duration", time.Duration(0), time.Duration(0), enum.ErrParseValue},
		{
			"time.Time valid",
			"2023-01-01T00:00:00Z",
			time.Time{},
			time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			nil,
		},
		{"time.Time invalid", "not-a-time", time.Time{}, time.Time{}, enum.ErrParseValue},

		// Unsupported type
		{"unsupported type", "value", struct{}{}, struct{}{}, enum.ErrParseValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := enum.ParseValue(tt.input, tt.defaultVal)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("ParseValue() error = %v, wantErr %v", err, tt.err)
					return
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEnumAliases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single alias", "alias1", []string{"alias1"}},
		{"multiple aliases", "alias1,alias2,alias3", []string{"alias1", "alias2", "alias3"}},
		{"quoted aliases", `"alias1","alias2"`, []string{"alias1", "alias2"}},
		{"mixed quotes", `alias1,"alias2",alias3`, []string{"alias1", "alias2", "alias3"}},
		{"empty input", "", []string{""}},
		{"spaces", " alias1 , alias2 ", []string{"alias1", "alias2"}},
		{"trailing comma", "alias1,alias2,", []string{"alias1", "alias2"}},
		{"leading comma", ",alias1,alias2", []string{"alias1", "alias2"}},
		{"only commas", ",,", []string{}},
		{"quoted with spaces", `"alias with spaces","another alias"`, []string{"alias with spaces", "another alias"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := enum.ParseEnumAliases(tt.input)
			if !slices.Equal(got, tt.want) {
				t.Errorf("parseEnumAliases() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEnumFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		enumIota enum.EnumIota
		want     []enum.Field
		err      error
	}{
		{
			name:  "valid mixed fields",
			input: "42,true,hello",
			enumIota: enum.EnumIota{
				Fields: []enum.Field{
					{Name: "Number", Value: 0},
					{Name: "Flag", Value: nil},
					{Name: "Text", Value: ""},
				},
			},
			want: []enum.Field{
				{Name: "Number", Value: 42},
				{Name: "Text", Value: "hello"},
				{Name: "Flag", Value: "extra"},
			},
			err: enum.ErrParseValue,
		},
		{
			name:  "valid float fields",
			input: "3.14,2.71",
			enumIota: enum.EnumIota{
				Fields: []enum.Field{
					{Name: "Pi", Value: 0.0},
					{Name: "E", Value: 0.0},
				},
			},
			want: []enum.Field{
				{Name: "Pi", Value: 3.14},
				{Name: "E", Value: 2.71},
			},
		},
		{
			name:  "valid time duration",
			input: "1h30m",
			enumIota: enum.EnumIota{
				Fields: []enum.Field{
					{Name: "Duration", Value: time.Duration(0)},
				},
			},
			want: []enum.Field{
				{Name: "Duration", Value: 90 * time.Minute},
			},
		},
		{
			name:  "quoted strings",
			input: `"hello world","another string"`,
			enumIota: enum.EnumIota{
				Fields: []enum.Field{
					{Name: "Greeting", Value: ""},
					{Name: "Message", Value: ""},
				},
			},
			want: []enum.Field{
				{Name: "Greeting", Value: "hello world"},
				{Name: "Message", Value: "another string"},
			},
		},
		{
			name:  "empty field value",
			input: "42,,hello",
			enumIota: enum.EnumIota{
				Fields: []enum.Field{
					{Name: "Number", Value: 0},
					{Name: "Flag", Value: nil},
					{Name: "Text", Value: ""},
				},
			},
			err: enum.ErrFieldEmptyValue,
		},
		{
			name:  "invalid type conversion",
			input: "not-a-number",
			enumIota: enum.EnumIota{
				Fields: []enum.Field{
					{Name: "Number", Value: 0},
				},
			},
			err: enum.ErrParseValue,
		},
		{
			name:  "more fields than expected",
			input: "42,true,hello,extra",
			enumIota: enum.EnumIota{
				Fields: []enum.Field{
					{Name: "Number", Value: 0},
					{Name: "Flag", Value: nil},
				},
			},
			want: []enum.Field{
				{Name: "Number", Value: 42},
				{Name: "Flag", Value: enum.ErrParseValue},
				{Name: "Text", Value: "hello"},
				{Name: "Message", Value: "extra"},
			},
			err: enum.ErrParseValue,
		},
		{
			name:  "fewer fields than input",
			input: "42",
			enumIota: enum.EnumIota{
				Fields: []enum.Field{
					{Name: "Number", Value: 0},
					{Name: "Flag", Value: nil},
					{Name: "Text", Value: ""},
				},
			},
			want: []enum.Field{
				{Name: "Number", Value: 42},
			},
			err: enum.ErrFieldEmptyValue,
		},
		{
			name:     "empty input",
			input:    "",
			enumIota: enum.EnumIota{Fields: []enum.Field{}},
			want:     []enum.Field{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := enum.ParseEnumFields(tt.input, tt.enumIota)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("parseEnumFields() error = %v, wantErr %v", err, tt.err)
					return
				}
				return
			}
			if !slices.EqualFunc(got, tt.want,
				func(a, b enum.Field) bool {
					return a.Name == b.Name && reflect.DeepEqual(a.Value, b.Value)
				}) {
				t.Errorf("parseEnumFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		comment    string
		wantOpener string
		wantCloser string
		wantFields []enum.Field
	}{
		{
			name:       "empty comment",
			comment:    "",
			wantOpener: " ",
			wantCloser: " ",
			wantFields: []enum.Field{},
		},
		{
			name:       "simple fields",
			comment:    "Name string, Age int",
			wantOpener: " ",
			wantCloser: " ",
			wantFields: []enum.Field{
				{Name: "Name", Value: ""},
				{Name: "Age", Value: 0},
			},
		},
		{
			name:       "bracket notation",
			comment:    "Name[string], Age[int]",
			wantOpener: "[",
			wantCloser: "]",
			wantFields: []enum.Field{
				{Name: "Name", Value: ""},
				{Name: "Age", Value: 0},
			},
		},
		{
			name:       "parenthesis notation",
			comment:    "Name(string), Age(int)",
			wantOpener: "(",
			wantCloser: ")",
			wantFields: []enum.Field{
				{Name: "Name", Value: ""},
				{Name: "Age", Value: 0},
			},
		},
		{
			name:       "complex types",
			comment:    "Duration time.Duration, Timestamp time.Time",
			wantOpener: " ",
			wantCloser: " ",
			wantFields: []enum.Field{
				{Name: "Duration", Value: time.Duration(0)},
				{Name: "Timestamp", Value: time.Time{}},
			},
		},
		{
			name:       "single field no name",
			comment:    "string",
			wantOpener: " ",
			wantCloser: " ",
			wantFields: []enum.Field{
				{Name: "", Value: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotOpener, gotCloser, gotFields := enum.ExtractFields(tt.comment)
			if gotOpener != tt.wantOpener {
				t.Errorf("extractFields() opener = %v, want %v", gotOpener, tt.wantOpener)
			}
			if gotCloser != tt.wantCloser {
				t.Errorf("extractFields() closer = %v, want %v", gotCloser, tt.wantCloser)
			}
			if !slices.EqualFunc(gotFields, tt.wantFields,
				func(a, b enum.Field) bool {
					return a.Name == b.Name && reflect.DeepEqual(a.Value, b.Value)
				}) {
				t.Errorf("extractFields() fields = %v, want %v", gotFields, tt.wantFields)
			}
		})
	}
}

func TestFieldToType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		field string
		want  any
	}{
		{"bool", "bool", false},
		{"int", "int", 0},
		{"string", "string", ""},
		{"float64", "float64", 0.0},
		{"float32", "float32", float32(0.0)},
		{"time.Duration", "time.Duration", time.Duration(0)},
		{"time.Time", "time.Time", time.Time{}},
		{"int64", "int64", int64(0)},
		{"int32", "int32", int32(0)},
		{"int16", "int16", int16(0)},
		{"int8", "int8", int8(0)},
		{"uint64", "uint64", uint64(0)},
		{"uint32", "uint32", uint32(0)},
		{"uint16", "uint16", uint16(0)},
		{"uint8", "uint8", uint8(0)},
		{"uint", "uint", uint(0)},
		{"byte", "byte", byte(0)},
		{"rune", "rune", rune(0)},
		{"complex64", "complex64", complex64(0)},
		{"complex128", "complex128", complex128(0)},
		{"uintptr", "uintptr", uintptr(0)},
		{"unknown", "unknown", nil},
		{"empty", "", nil},
		{"with spaces", "  string  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := enum.FieldToType(tt.field)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fieldToType() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestOpenCloser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		field      string
		wantOpener string
		wantCloser string
	}{
		{"no brackets", "simple field", " ", " "},
		{"square brackets", "field[type]", "[", "]"},
		{"parentheses", "field(type)", "(", ")"},
		{"both brackets and parens", "field[type](something)", "[", "]"}, // Should prefer brackets
		{"empty", "", " ", " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotOpener, gotCloser := enum.OpenCloser(tt.field)
			if gotOpener != tt.wantOpener {
				t.Errorf("openCloser() opener = %v, want %v", gotOpener, tt.wantOpener)
			}
			if gotCloser != tt.wantCloser {
				t.Errorf("openCloser() closer = %v, want %v", gotCloser, tt.wantCloser)
			}
		})
	}
}

func TestExtractImports(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		enumIotas []enum.EnumIota
		want      []string
	}{
		{
			name:      "no imports",
			enumIotas: []enum.EnumIota{},
			want:      []string{},
		},
		{
			name: "time imports",
			enumIotas: []enum.EnumIota{
				{
					Fields: []enum.Field{
						{Name: "Duration", Value: time.Duration(0)},
						{Name: "Timestamp", Value: time.Time{}},
					},
				},
			},
			want: []string{"time"},
		},
		{
			name: "duplicate imports",
			enumIotas: []enum.EnumIota{
				{
					Fields: []enum.Field{
						{Name: "Duration1", Value: time.Duration(0)},
						{Name: "Duration2", Value: time.Duration(0)},
						{Name: "Timestamp", Value: time.Time{}},
					},
				},
			},
			want: []string{"time"},
		},
		{
			name: "no dot in type",
			enumIotas: []enum.EnumIota{
				{
					Fields: []enum.Field{
						{Name: "Number", Value: 0},
						{Name: "Text", Value: ""},
					},
				},
			},
			want: []string{},
		},
		{
			name: "empty fields",
			enumIotas: []enum.EnumIota{
				{
					Fields: []enum.Field{},
				},
			},
			want: []string{},
		},
		{
			name: "external imports",
			enumIotas: []enum.EnumIota{
				{
					Fields: []enum.Field{
						{Name: "Lock", Value: sync.Mutex{}},
					},
				},
			},
			want: []string{"sync"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := enum.ExtractImports(tt.enumIotas)
			if !slices.Equal(got, tt.want) {
				t.Errorf("extractImports() = %v, want %v", got, tt.want)
			}
		})
	}
}
