// Package strings provides specialized string handling utilities for enum generation.
//
// This package wraps standard library string functions and adds custom functionality
// specifically tailored for enum code generation, including:
//   - Case conversion (upper, lower, camel case)
//   - Intelligent pluralization with irregular word handling
//   - String parsing and manipulation
//
// By centralizing string operations here, we avoid code duplication and maintain
// consistent string handling across the codebase while preventing namespace
// collisions with the standard library.
package strings

import "strings"

// irregular contains mappings for words that don't follow standard English
// pluralization rules, ensuring correct pluralization for special cases.
var irregular = map[string]string{
	"man":      "men",
	"woman":    "women",
	"child":    "children",
	"foot":     "feet",
	"tooth":    "teeth",
	"goose":    "geese",
	"mouse":    "mice",
	"ox":       "oxen",
	"person":   "people",
	"index":    "indices",
	"matrix":   "matrices",
	"vertex":   "vertices",
	"datum":    "data",
	"medium":   "media",
	"analysis": "analyses",
	"crisis":   "crises",
	"status":   "statuses",
}

// PluralAndCamelPlural returns both lowercase and camel case versions of a plural form
// for a given type name. It handles irregular plurals, compound words with
// underscores, and applies standard English pluralization rules.
func PluralAndCamelPlural(iotaType string) (string, string) {
	return Plural(iotaType), CamelCase(Plural(strings.ToLower(iotaType)))
}

// Plural returns the plural form of a given type name. It handles irregular plurals,
// compound words with underscores, and already pluralized words. It preserves the
// original casing (uppercase/lowercase) of the input.
func Plural(iotaType string) string {
	if len(iotaType) == 0 {
		return ""
	}
	lower := ToLower(iotaType)
	isUpper := ToUpper(iotaType) == iotaType
	for _, plural := range irregular {
		if lower == plural {
			if isUpper {
				return ToUpper(lower)
			}
			return lower
		}
	}
	if isRegularPlural(lower) {
		if isUpper {
			return ToUpper(lower)
		}
		return lower
	}
	if Contains(lower, "_") {
		parts := Split(lower, "_")
		lastIndex := len(parts) - 1
		lastPart := parts[lastIndex]

		alreadyPlural := false
		for _, plural := range irregular {
			if lastPart == plural {
				alreadyPlural = true
				break
			}
		}
		if isRegularPlural(lastPart) {
			alreadyPlural = true
		}
		if !alreadyPlural {
			if p, ok := irregular[lastPart]; ok {
				parts[lastIndex] = p
			} else {
				parts[lastIndex] = regularPlural(lastPart)
			}
		}
		result := Join(parts, "_")
		if isUpper {
			return ToUpper(result)
		}
		return result
	}
	if p, ok := irregular[lower]; ok {
		result := p
		if isUpper {
			return ToUpper(result)
		}
		return result
	}
	result := regularPlural(lower)
	if isUpper {
		return ToUpper(result)
	}
	return result
}

func isRegularPlural(word string) bool {
	if len(word) < 2 {
		return false
	}
	return HasSuffix(word, "s") && !HasSuffix(word, "ss") &&
		!HasSuffix(word, "us") && !HasSuffix(word, "is")
}

// CamelCase converts a string to camel case format by capitalizing the first letter
// of the input string or each segment after a separator like underscore. It removes
// separators and preserves proper casing for each segment.
//
// Examples:
//   - "hello_world" → "HelloWorld"
//   - "dog_house" → "DogHouse"
//   - "DOG_HOUSE" → "DogHouse"
func CamelCase(in string) string {
	if len(in) == 0 {
		return ""
	}
	if Contains(in, "_") {
		parts := Split(in, "_")
		var result strings.Builder
		for _, part := range parts {
			if part == "" {
				continue
			}
			lower := ToLower(part)
			if lower != "" {
				result.WriteString(ToUpper(lower[:1]))
				result.WriteString(lower[1:])
			}
		}
		return result.String()
	}
	lower := ToLower(in)
	if len(lower) == 0 {
		return ""
	}
	return ToUpper(lower[:1]) + lower[1:]
}

func regularPlural(word string) string {
	l := len(word)
	if l == 0 {
		return word
	}
	if l == 1 {
		return word + "s"
	}
	lastChar := word[l-1]
	secondLastChar := word[l-2]
	isVowel := func(c byte) bool {
		return c == 'a' || c == 'e' || c == 'i' || c == 'o' || c == 'u'
	}
	if lastChar == 'y' && !isVowel(secondLastChar) {
		return word[:l-1] + "ies"
	}
	if lastChar == 's' || lastChar == 'x' || lastChar == 'z' || lastChar == 'o' {
		return word + "es"
	}
	if l > 1 {
		lastTwo := word[l-2:]
		if lastTwo == "ch" || lastTwo == "sh" || lastTwo == "ss" {
			return word + "es"
		}
	}
	return word + "s"
}

// ToLower converts a string to lowercase.
// This is a wrapper around strings.ToLower.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper converts a string to uppercase.
// This is a wrapper around strings.ToUpper.
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// Contains reports whether substr is within s.
// This is a wrapper around strings.Contains.
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TrimSpace returns a slice of the string with all leading
// and trailing white space removed.
// This is a wrapper around strings.TrimSpace.
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Split slices s into all substrings separated by sep and returns them.
// This is a wrapper around strings.Split.
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// TrimLeft returns a slice of the string with all leading
// characters contained in cutset removed.
// This is a wrapper around strings.TrimLeft.
func TrimLeft(s, cutset string) string {
	return strings.TrimLeft(s, cutset)
}

// HasPrefix tests whether the string starts with prefix.
// This is a wrapper around strings.HasPrefix.
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix tests whether the string ends with suffix.
// This is a wrapper around strings.HasSuffix.
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Index returns the index of the first instance of sep in s,
// or -1 if sep is not present in s.
// This is a wrapper around strings.Index.
func Index(s, sep string) int {
	return strings.Index(s, sep)
}

// Count counts the number of non-overlapping instances of sep in s.
// This is a wrapper around strings.Count.
func Count(s, sep string) int {
	return strings.Count(s, sep)
}

// Join concatenates the elements of a string slice to create a single string
// with the specified separator between elements.
// This is a wrapper around strings.Join.
func Join(s1 []string, s2 string) string {
	return strings.Join(s1, s2)
}

// SplitN slices s into substrings separated by sep and returns them.
// The count determines the number of substrings to return.
// This is a wrapper around strings.SplitN.
func SplitN(s, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

// Fields splits the string s around each instance of one or more consecutive white space
// characters, as defined by unicode.IsSpace, returning a slice of substrings.
// This is a wrapper around strings.Fields.
func Fields(s string) []string {
	return strings.Fields(s)
}

// Trim returns a slice of the string s with all leading and trailing Unicode
// code points contained in cutset removed.
// This is a wrapper around strings.Trim.
func Trim(s, cutset string) string {
	return strings.Trim(s, s)
}

// TrimPrefix returns s without the provided leading prefix string.
// If s doesn't start with prefix, s is returned unchanged.
// This is a wrapper around strings.TrimPrefix.
func TrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// LastIndex returns the index of the last instance of sep in s, or -1 if sep is not present in s.
// This is a wrapper around strings.LastIndex.
func LastIndex(s, sep string) int {
	return strings.LastIndex(s, sep)
}

// ReplaceAll returns a copy of the string s with all non-overlapping instances of old replaced by new.
// If old is empty, it matches at the beginning of the string and after each UTF-8 sequence, yielding up to k+1 replacements for a len(s) = k string.
// If new is empty, it removes all instances of old from s.
// This is a wrapper around strings.ReplaceAll.
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

type Builder struct {
	b *strings.Builder
}

func (b *Builder) WriteString(s string) {
	if b.b == nil {
		b.b = &strings.Builder{}
	}
	_, _ = b.b.WriteString(s)
}

func (b *Builder) String() string {
	if b.b == nil {
		return ""
	}
	return b.b.String()
}

func (b *Builder) Len() int {
	if b.b == nil {
		return 0
	}
	return b.b.Len()
}

func (b *Builder) Reset() {
	if b.b == nil {
		return
	}
	b.b.Reset()
}

func (b *Builder) WriteByte(c byte) error {
	if b.b == nil {
		b.b = &strings.Builder{}
	}
	return b.b.WriteByte(c)
}
