// strings/strings_test.go
package strings_test

import (
	"testing"

	"github.com/zarldev/goenums/strings"
)

func TestPlural(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

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
			name:     "single letter",
			input:    "a",
			expected: "as",
		},
		{
			name:     "regular word",
			input:    "dog",
			expected: "dogs",
		},
		{
			name:     "word ending in y with consonant before",
			input:    "city",
			expected: "cities",
		},
		{
			name:     "word ending in y with vowel before",
			input:    "boy",
			expected: "boys",
		},
		{
			name:     "word ending in s",
			input:    "bus",
			expected: "buses",
		},
		{
			name:     "word ending in x",
			input:    "box",
			expected: "boxes",
		},
		{
			name:     "word ending in z",
			input:    "quiz",
			expected: "quizes",
		},
		{
			name:     "word ending in o",
			input:    "hero",
			expected: "heroes",
		},
		{
			name:     "word ending in ch",
			input:    "match",
			expected: "matches",
		},
		{
			name:     "word ending in sh",
			input:    "dish",
			expected: "dishes",
		},
		{
			name:     "word ending in ss",
			input:    "glass",
			expected: "glasses",
		},
		{
			name:     "irregular: man",
			input:    "man",
			expected: "men",
		},
		{
			name:     "irregular: woman",
			input:    "woman",
			expected: "women",
		},
		{
			name:     "irregular: status",
			input:    "status",
			expected: "statuses",
		},
		{
			name:     "compound word",
			input:    "dog_house",
			expected: "dog_houses",
		},
		{
			name:     "compound irregular word",
			input:    "dog_foot",
			expected: "dog_feet",
		},
		{
			name:     "uppercase word",
			input:    "DOG",
			expected: "DOGS",
		},
		{
			name:     "uppercase compound word",
			input:    "DOG_HOUSE",
			expected: "DOG_HOUSES",
		},
		{
			name:     "uppercase irregular word",
			input:    "MAN",
			expected: "MEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if ctx.Err() != nil {
				t.Skip("context cancelled")
			}
			got := strings.Plural(tt.input)
			if got != tt.expected {
				t.Errorf("for input %q: got %q, expected %q",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestCamelCase(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single letter",
			input:    "a",
			expected: "A",
		},
		{
			name:     "lowercase word",
			input:    "dog",
			expected: "Dog",
		},
		{
			name:     "uppercase word - remains uppercase - cannot tell which chars should be lowercase",
			input:    "DOG",
			expected: "DOG",
		},
		{
			name:     "already camel case",
			input:    "Dog",
			expected: "Dog",
		},
		{
			name:     "with underscores",
			input:    "dog_house",
			expected: "DogHouse",
		},
		{
			name:     "with dashes",
			input:    "dog-house",
			expected: "DogHouse",
		},
		{
			name:     "with dots",
			input:    "dog.house",
			expected: "DogHouse",
		},
		{
			name:     "with em-dashes",
			input:    "dog—house",
			expected: "DogHouse",
		},
		{
			name:     "with en-dashes",
			input:    "dog–house",
			expected: "DogHouse",
		},
		{
			name:     "with spaces",
			input:    "dog house",
			expected: "DogHouse",
		},
		{
			name:     "with trailing spaces",
			input:    "dog house ",
			expected: "DogHouse",
		},
		{
			name:     "with leading spaces",
			input:    " dog house",
			expected: "DogHouse",
		},
		{
			name:     "already plural regular",
			input:    "dogs",
			expected: "Dogs",
		},
		{
			name:     "already plural irregular",
			input:    "men",
			expected: "Men",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if ctx.Err() != nil {
				t.Skip("context cancelled")
			}

			got := strings.CamelCase(tt.input)
			if got != tt.expected {
				t.Errorf("for input %q: got %q, expected %q",
					tt.input, got, tt.expected)
			}
		})
	}
}

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
			got := strings.Singularize(tt.input)
			if got != tt.expected {
				t.Errorf("Singularize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
