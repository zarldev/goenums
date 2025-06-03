// strings/strings_test.go
package strings_test

import (
	"testing"
	"time"

	"github.com/zarldev/goenums/strings"
)

func TestSingularise(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plural",
			input:    "dogs",
			expected: "dog",
		},
		{
			name:     "singular",
			input:    "dog",
			expected: "dog",
		},
		{
			name:     "plural with spaces",
			input:    "dog houses",
			expected: "dog house",
		},
		{
			name:     "plural with hyphens",
			input:    "dog-houses",
			expected: "dog-house",
		},
		{
			name:     "plural with underscores",
			input:    "dog_houses",
			expected: "dog_house",
		},
		{
			name:     "plural with dots",
			input:    "dog.houses",
			expected: "dog.house",
		},
		{
			name:     "plural with em-dashes",
			input:    "dog—houses",
			expected: "dog—house",
		},
		{
			name:     "plural with en-dashes",
			input:    "dog–houses",
			expected: "dog–house",
		},
		{
			name:     "plural with spaces",
			input:    "dog houses",
			expected: "dog house",
		},
		{
			name:     "plural with trailing spaces",
			input:    "dog houses ",
			expected: "dog house",
		},
		{
			name:     "plural with leading spaces",
			input:    " dog houses",
			expected: "dog house",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Singularise(tt.input)
			if got != tt.expected {
				t.Errorf("singularise(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSplitBySpace(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		input          string
		expectedBefore string
		expectedAfter  string
	}{
		{
			name:           "simple split",
			input:          "hello world",
			expectedBefore: "hello",
			expectedAfter:  "world",
		},
		{
			name:           "no space",
			input:          "hello",
			expectedBefore: "hello",
			expectedAfter:  "",
		},
		{
			name:           "multiple spaces",
			input:          "hello world test",
			expectedBefore: "hello",
			expectedAfter:  "world test",
		},
		{
			name:           "quoted string with space",
			input:          `"hello world" test`,
			expectedBefore: `"hello world"`,
			expectedAfter:  "test",
		},
		{
			name:           "quoted string no space after",
			input:          `"hello world"`,
			expectedBefore: `"hello world"`,
			expectedAfter:  "",
		},
		{
			name:           "space inside quotes",
			input:          `"hello world"`,
			expectedBefore: `"hello world"`,
			expectedAfter:  "",
		},
		{
			name:           "multiple quotes",
			input:          `"hello" "world"`,
			expectedBefore: `"hello"`,
			expectedAfter:  `"world"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			before, after := strings.SplitBySpace(tt.input)
			if before != tt.expectedBefore {
				t.Errorf("SplitBySpace(%q) before = %q, want %q", tt.input, before, tt.expectedBefore)
			}
			if after != tt.expectedAfter {
				t.Errorf("SplitBySpace(%q) after = %q, want %q", tt.input, after, tt.expectedAfter)
			}
		})
	}
}

func TestIsPlural(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "regular plural",
			input:    "dogs",
			expected: true,
		},
		{
			name:     "irregular plural",
			input:    "men",
			expected: true,
		},
		{
			name:     "singular",
			input:    "dog",
			expected: false,
		},
		{
			name:     "ends with es",
			input:    "boxes",
			expected: true,
		},
		{
			name:     "uppercase irregular plural",
			input:    "MEN",
			expected: true,
		},
		{
			name:     "status (irregular to plural mapping)",
			input:    "status",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.IsPlural(tt.input)
			if got != tt.expected {
				t.Errorf("IsPlural(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSingulariseAdvanced(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "already singular",
			input:    "dog",
			expected: "dog",
		},
		{
			name:     "regular plural",
			input:    "dogs",
			expected: "dog",
		},
		{
			name:     "irregular plural",
			input:    "men",
			expected: "man",
		},
		{
			name:     "snake_case plural",
			input:    "dog_houses",
			expected: "dog_house",
		},
		{
			name:     "snake_case irregular",
			input:    "dog_feet",
			expected: "dog_foot",
		},
		{
			name:     "kebab-case plural",
			input:    "dog-houses",
			expected: "dog-house",
		},
		{
			name:     "uppercase plural",
			input:    "DOGS",
			expected: "DOG",
		},
		{
			name:     "uppercase irregular",
			input:    "MEN",
			expected: "MAN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Singularise(tt.input)
			if got != tt.expected {
				t.Errorf("Singularise(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEnumBuilder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "WriteString and String",
			testFunc: func(t *testing.T) {
				var b strings.EnumBuilder
				b.WriteString("hello")
				b.WriteString(" world")
				got := b.String()
				expected := "hello world"
				if got != expected {
					t.Errorf("EnumBuilder.String() = %q, want %q", got, expected)
				}
			},
		},
		{
			name: "Len",
			testFunc: func(t *testing.T) {
				var b strings.EnumBuilder
				b.WriteString("hello")
				got := b.Len()
				expected := 5
				if got != expected {
					t.Errorf("EnumBuilder.Len() = %d, want %d", got, expected)
				}
			},
		},
		{
			name: "Reset",
			testFunc: func(t *testing.T) {
				var b strings.EnumBuilder
				b.WriteString("hello")
				b.Reset()
				got := b.String()
				expected := ""
				if got != expected {
					t.Errorf("EnumBuilder.String() after Reset() = %q, want %q", got, expected)
				}
			},
		},
		{
			name: "Grow",
			testFunc: func(t *testing.T) {
				var b strings.EnumBuilder
				b.Grow(100)
				b.WriteString("hello")
				got := b.String()
				expected := "hello"
				if got != expected {
					t.Errorf("EnumBuilder.String() after Grow() = %q, want %q", got, expected)
				}
			},
		},
		{
			name: "WriteByte",
			testFunc: func(t *testing.T) {
				var b strings.EnumBuilder
				err := b.WriteByte('h')
				if err != nil {
					t.Errorf("EnumBuilder.WriteByte() error = %v", err)
				}
				got := b.String()
				expected := "h"
				if got != expected {
					t.Errorf("EnumBuilder.String() after WriteByte() = %q, want %q", got, expected)
				}
			},
		},
		{
			name: "Write",
			testFunc: func(t *testing.T) {
				var b strings.EnumBuilder
				n, err := b.Write([]byte("hello"))
				if err != nil {
					t.Errorf("EnumBuilder.Write() error = %v", err)
				}
				if n != 5 {
					t.Errorf("EnumBuilder.Write() returned %d, want 5", n)
				}
				got := b.String()
				expected := "hello"
				if got != expected {
					t.Errorf("EnumBuilder.String() after Write() = %q, want %q", got, expected)
				}
			},
		},
		{
			name: "nil builder operations",
			testFunc: func(t *testing.T) {
				var b strings.EnumBuilder
				// Test operations on nil builder
				if b.String() != "" {
					t.Errorf("nil EnumBuilder.String() = %q, want empty string", b.String())
				}
				if b.Len() != 0 {
					t.Errorf("nil EnumBuilder.Len() = %d, want 0", b.Len())
				}
				b.Reset() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestEnumWriter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "NewEnumWriter default",
			testFunc: func(t *testing.T) {
				w := strings.NewEnumWriter()
				if w == nil {
					t.Error("NewEnumWriter() returned nil")
				}
			},
		},
		{
			name: "NewEnumWriter with custom writer",
			testFunc: func(t *testing.T) {
				var buf []byte
				customWriter := &testWriter{buf: &buf}
				w := strings.NewEnumWriter(strings.WithWriter(customWriter))

				n, err := w.Write([]byte("hello"))
				if err != nil {
					t.Errorf("EnumWriter.Write() error = %v", err)
				}
				if n != 5 {
					t.Errorf("EnumWriter.Write() returned %d, want 5", n)
				}
				if string(*customWriter.buf) != "hello" {
					t.Errorf("EnumWriter wrote %q, want %q", string(*customWriter.buf), "hello")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

type testWriter struct {
	buf *[]byte
}

func (w *testWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

func TestAsType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string",
			input:    "hello",
			expected: "string",
		},
		{
			name:     "int",
			input:    42,
			expected: "int",
		},
		{
			name:     "bool",
			input:    true,
			expected: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.AsType(tt.input)
			if got != tt.expected {
				t.Errorf("AsType(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIfy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "rune slice",
			input:    []rune("hello"),
			expected: `"hello"`,
		},
		{
			name:     "byte slice",
			input:    []byte("hello"),
			expected: `"hello"`,
		},
		{
			name:     "bool true",
			input:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			input:    false,
			expected: "false",
		},
		{
			name:     "int",
			input:    42,
			expected: "42",
		},
		{
			name:     "float",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "zero duration",
			input:    time.Duration(0),
			expected: "",
		},
		{
			name:     "negative duration",
			input:    -1 * time.Hour,
			expected: "time.Hour * -1",
		},
		{
			name:     "complex struct",
			input:    struct{ value string }{value: "test"},
			expected: "{test}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Ify(tt.input)
			if got != tt.expected {
				t.Errorf("Ify(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIfyTime(t *testing.T) {
	t.Parallel()

	// Test time.Time
	t.Run("time.Time", func(t *testing.T) {
		t.Parallel()
		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		got := strings.Ify(testTime)
		expected := "2023-01-01T12:00:00Z"
		if got != expected {
			t.Errorf("Ify(time.Time) = %q, want %q", got, expected)
		}
	})

	// Test time.Duration - the Ify function converts everything to hours if > 0
	durationTests := []struct {
		name     string
		input    time.Duration
		contains string // We'll check if the result contains this string
	}{
		{
			name:     "hours",
			input:    2 * time.Hour,
			contains: "time.Hour",
		},
		{
			name:     "minutes converted to hours",
			input:    30 * time.Minute,
			contains: "time.Hour", // 30 minutes = 0.5 hours
		},
		{
			name:     "seconds converted to hours",
			input:    45 * time.Second,
			contains: "time.Hour", // 45 seconds = 0.0125 hours
		},
	}

	for _, tt := range durationTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Ify(tt.input)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("Ify(%v) = %q, expected to contain %q", tt.input, got, tt.contains)
			}
		})
	}
}

func TestIfyStringer(t *testing.T) {
	t.Parallel()

	// Test fmt.Stringer interface
	stringer := &testStringer{value: "test"}
	got := strings.Ify(stringer)
	expected := `"test"`
	if got != expected {
		t.Errorf("Ify(Stringer) = %q, want %q", got, expected)
	}
}

type testStringer struct {
	value string
}

func (ts *testStringer) String() string {
	return ts.value
}

func TestIfiableNumeric(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "small integer",
			testFunc: func(t *testing.T) {
				got := strings.IfiableNumeric(42)
				expected := "42"
				if got != expected {
					t.Errorf("IfiableNumeric(42) = %q, want %q", got, expected)
				}
			},
		},
		{
			name: "small float",
			testFunc: func(t *testing.T) {
				got := strings.IfiableNumeric(3.14)
				expected := "3.14"
				if got != expected {
					t.Errorf("IfiableNumeric(3.14) = %q, want %q", got, expected)
				}
			},
		},
		{
			name: "large number scientific notation",
			testFunc: func(t *testing.T) {
				got := strings.IfiableNumeric(1e7)
				if !strings.Contains(got, "e") {
					t.Errorf("IfiableNumeric(1e7) = %q, expected scientific notation", got)
				}
			},
		},
		{
			name: "very small number scientific notation",
			testFunc: func(t *testing.T) {
				got := strings.IfiableNumeric(1e-7)
				if !strings.Contains(got, "e") {
					t.Errorf("IfiableNumeric(1e-7) = %q, expected scientific notation", got)
				}
			},
		},
		{
			name: "zero",
			testFunc: func(t *testing.T) {
				got := strings.IfiableNumeric(0)
				expected := "0"
				if got != expected {
					t.Errorf("IfiableNumeric(0) = %q, want %q", got, expected)
				}
			},
		},
		{
			name: "negative number",
			testFunc: func(t *testing.T) {
				got := strings.IfiableNumeric(-42.5)
				expected := "-42.5"
				if got != expected {
					t.Errorf("IfiableNumeric(-42.5) = %q, want %q", got, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

func TestDetectCaseAndMatchCasing(t *testing.T) {
	t.Parallel()

	// These functions are internal but we can test them indirectly through Plural and Singularise
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "all uppercase",
			input:    "DOGS",
			expected: "DOG",
		},
		{
			name:     "all lowercase",
			input:    "dogs",
			expected: "dog",
		},
		{
			name:     "title case",
			input:    "Dogs",
			expected: "Dog",
		},
		{
			name:     "mixed case",
			input:    "DoGs",
			expected: "DoG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Singularise(tt.input)
			if got != tt.expected {
				t.Errorf("Singularise(%q) = %q, want %q (testing case detection)", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPluralise(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "as",
		},
		{
			name:     "already plural",
			input:    "dogs",
			expected: "dogs",
		},
		{
			name:     "irregular word",
			input:    "man",
			expected: "men",
		},
		{
			name:     "word ending in y",
			input:    "city",
			expected: "cities",
		},
		{
			name:     "word ending in s",
			input:    "bus",
			expected: "buses",
		},
		{
			name:     "word ending in ch",
			input:    "church",
			expected: "churches",
		},
		{
			name:     "regular word",
			input:    "cat",
			expected: "cats",
		},
		{
			name:     "word ending in x",
			input:    "box",
			expected: "boxes",
		},
		{
			name:     "word ending in sh",
			input:    "dish",
			expected: "dishes",
		},
		{
			name:     "word ending in z",
			input:    "quiz",
			expected: "quizzes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Pluralise(tt.input)
			if got != tt.expected {
				t.Errorf("Pluralise(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCamel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "lowercase word",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "already capitalized",
			input:    "Hello",
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Camel(tt.input)
			if got != tt.expected {
				t.Errorf("Camel(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLower1stCharacter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character",
			input:    "A",
			expected: "a",
		},
		{
			name:     "uppercase word",
			input:    "Hello",
			expected: "hello",
		},
		{
			name:     "already lowercase",
			input:    "hello",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Lower1stCharacter(tt.input)
			if got != tt.expected {
				t.Errorf("Lower1stCharacter(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStringWrappers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "ToLower",
			testFunc: func(t *testing.T) {
				got := strings.ToLower("HELLO")
				expected := "hello"
				if got != expected {
					t.Errorf("ToLower(%q) = %q, want %q", "HELLO", got, expected)
				}
			},
		},
		{
			name: "ToUpper",
			testFunc: func(t *testing.T) {
				got := strings.ToUpper("hello")
				expected := "HELLO"
				if got != expected {
					t.Errorf("ToUpper(%q) = %q, want %q", "hello", got, expected)
				}
			},
		},
		{
			name: "Contains",
			testFunc: func(t *testing.T) {
				got := strings.Contains("hello world", "world")
				if !got {
					t.Errorf("Contains(%q, %q) = %v, want %v", "hello world", "world", got, true)
				}
			},
		},
		{
			name: "TrimSpace",
			testFunc: func(t *testing.T) {
				got := strings.TrimSpace("  hello  ")
				expected := "hello"
				if got != expected {
					t.Errorf("TrimSpace(%q) = %q, want %q", "  hello  ", got, expected)
				}
			},
		},
		{
			name: "TrimPrefix",
			testFunc: func(t *testing.T) {
				got := strings.TrimPrefix("hello world", "hello ")
				expected := "world"
				if got != expected {
					t.Errorf("TrimPrefix(%q, %q) = %q, want %q", "hello world", "hello ", got, expected)
				}
			},
		},
		{
			name: "TrimSuffix",
			testFunc: func(t *testing.T) {
				got := strings.TrimSuffix("hello world", " world")
				expected := "hello"
				if got != expected {
					t.Errorf("TrimSuffix(%q, %q) = %q, want %q", "hello world", " world", got, expected)
				}
			},
		},
		{
			name: "Split",
			testFunc: func(t *testing.T) {
				got := strings.Split("a,b,c", ",")
				expected := []string{"a", "b", "c"}
				if len(got) != len(expected) {
					t.Errorf("Split length mismatch: got %d, want %d", len(got), len(expected))
				}
				for i, v := range expected {
					if got[i] != v {
						t.Errorf("Split[%d] = %q, want %q", i, got[i], v)
					}
				}
			},
		},
		{
			name: "TrimLeft",
			testFunc: func(t *testing.T) {
				got := strings.TrimLeft("!!!hello", "!")
				expected := "hello"
				if got != expected {
					t.Errorf("TrimLeft(%q, %q) = %q, want %q", "!!!hello", "!", got, expected)
				}
			},
		},
		{
			name: "TrimRight",
			testFunc: func(t *testing.T) {
				got := strings.TrimRight("hello!!!", "!")
				expected := "hello"
				if got != expected {
					t.Errorf("TrimRight(%q, %q) = %q, want %q", "hello!!!", "!", got, expected)
				}
			},
		},
		{
			name: "HasPrefix",
			testFunc: func(t *testing.T) {
				got := strings.HasPrefix("hello world", "hello")
				if !got {
					t.Errorf("HasPrefix(%q, %q) = %v, want %v", "hello world", "hello", got, true)
				}
			},
		},
		{
			name: "HasSuffix",
			testFunc: func(t *testing.T) {
				got := strings.HasSuffix("hello world", "world")
				if !got {
					t.Errorf("HasSuffix(%q, %q) = %v, want %v", "hello world", "world", got, true)
				}
			},
		},
		{
			name: "Index",
			testFunc: func(t *testing.T) {
				got := strings.Index("hello world", "world")
				expected := 6
				if got != expected {
					t.Errorf("Index(%q, %q) = %d, want %d", "hello world", "world", got, expected)
				}
			},
		},
		{
			name: "Count",
			testFunc: func(t *testing.T) {
				got := strings.Count("hello hello", "hello")
				expected := 2
				if got != expected {
					t.Errorf("Count(%q, %q) = %d, want %d", "hello hello", "hello", got, expected)
				}
			},
		},
		{
			name: "Join",
			testFunc: func(t *testing.T) {
				got := strings.Join([]string{"a", "b", "c"}, ",")
				expected := "a,b,c"
				if got != expected {
					t.Errorf("Join(%v, %q) = %q, want %q", []string{"a", "b", "c"}, ",", got, expected)
				}
			},
		},
		{
			name: "SplitN",
			testFunc: func(t *testing.T) {
				got := strings.SplitN("a,b,c,d", ",", 3)
				expected := []string{"a", "b", "c,d"}
				if len(got) != len(expected) {
					t.Errorf("SplitN length mismatch: got %d, want %d", len(got), len(expected))
				}
				for i, v := range expected {
					if got[i] != v {
						t.Errorf("SplitN[%d] = %q, want %q", i, got[i], v)
					}
				}
			},
		},
		{
			name: "Fields",
			testFunc: func(t *testing.T) {
				got := strings.Fields("hello   world  test")
				expected := []string{"hello", "world", "test"}
				if len(got) != len(expected) {
					t.Errorf("Fields length mismatch: got %d, want %d", len(got), len(expected))
				}
				for i, v := range expected {
					if got[i] != v {
						t.Errorf("Fields[%d] = %q, want %q", i, got[i], v)
					}
				}
			},
		},
		{
			name: "LastIndex",
			testFunc: func(t *testing.T) {
				got := strings.LastIndex("hello world hello", "hello")
				expected := 12
				if got != expected {
					t.Errorf("LastIndex(%q, %q) = %d, want %d", "hello world hello", "hello", got, expected)
				}
			},
		},
		{
			name: "ReplaceAll",
			testFunc: func(t *testing.T) {
				got := strings.ReplaceAll("hello world hello", "hello", "hi")
				expected := "hi world hi"
				if got != expected {
					t.Errorf("ReplaceAll(%q, %q, %q) = %q, want %q", "hello world hello", "hello", "hi", got, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}
