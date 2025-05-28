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
		wantErr    bool
	}{
		// Boolean tests
		{"bool true", "true", false, true, false},
		{"bool false", "false", false, false, false},
		{"bool invalid", "maybe", false, false, true},
		{"bool 1", "1", false, true, false},
		{"bool 0", "0", false, false, false},

		// Integer tests
		{"int valid", "42", 0, 42, false},
		{"int negative", "-42", 0, -42, false},
		{"int invalid", "not-a-number", 0, 0, true},
		{"int64 valid", "9223372036854775807", int64(0), int64(9223372036854775807), false},
		{"int32 valid", "2147483647", int32(0), int32(2147483647), false},
		{"int16 valid", "32767", int16(0), int16(32767), false},
		{"int8 valid", "127", int8(0), int8(127), false},

		// Unsigned integer tests
		{"uint valid", "42", uint(0), uint(42), false},
		{"uint64 valid", "18446744073709551615", uint64(0), uint64(18446744073709551615), false},
		{"uint32 valid", "4294967295", uint32(0), uint32(4294967295), false},
		{"uint16 valid", "65535", uint16(0), uint16(65535), false},
		{"uint8 valid", "255", uint8(0), uint8(255), false},
		{"uint negative", "-1", uint(0), uint(0), true},

		// Float tests
		{"float64 valid", "3.14", 0.0, 3.14, false},
		{"float64 scientific", "1.23e-4", 0.0, 1.23e-4, false},
		{"float64 invalid", "not-a-float", 0.0, 0.0, true},
		{"float32 valid", "3.14", float32(0), float32(3.14), false},
		{"float32 invalid", "not-a-float", float32(0), float32(0), true},

		// String tests
		{"string quoted", `"hello"`, "", "hello", false},
		{"string unquoted", "hello", "", "hello", false},
		{"string empty quoted", `""`, "", "", false},
		{"string empty", "", "", "", false},
		{"string with spaces", "hello world", "", "hello world", false},

		// Time tests
		{"time.Duration valid", "1h30m", time.Duration(0), 90 * time.Minute, false},
		{"time.Duration seconds", "45s", time.Duration(0), 45 * time.Second, false},
		{"time.Duration invalid", "invalid-duration", time.Duration(0), time.Duration(0), true},
		{
			"time.Time valid",
			"2023-01-01T00:00:00Z",
			time.Time{},
			time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			false,
		},
		{"time.Time invalid", "not-a-time", time.Time{}, time.Time{}, true},

		// Unsupported type
		{"unsupported type", "value", struct{}{}, struct{}{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := enum.ParseValue(tt.input, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
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
					{Name: "Flag", Value: false},
					{Name: "Text", Value: ""},
				},
			},
			want: []enum.Field{
				{Name: "Number", Value: 42},
				{Name: "Flag", Value: true},
				{Name: "Text", Value: "hello"},
			},
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
					{Name: "Flag", Value: false},
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
					{Name: "Flag", Value: false},
				},
			},
			want: []enum.Field{
				{Name: "Number", Value: 42},
				{Name: "Flag", Value: true},
			},
		},
		{
			name:  "fewer fields than input",
			input: "42",
			enumIota: enum.EnumIota{
				Fields: []enum.Field{
					{Name: "Number", Value: 0},
					{Name: "Flag", Value: false},
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
			if err != nil && !errors.Is(err, tt.err) {
				t.Errorf("parseEnumFields() error = %v, wantErr %v", err, tt.err)
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
