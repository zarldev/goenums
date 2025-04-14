// The strings package provides specialized string handling utilities for enum generation.
//
// This package wraps standard library string functions and adds custom functionality
// specifically tailored for enum code generation, including:
//   - Case conversion (upper, lower, camel case)
//   - Intelligent pluralization with irregular word handling
//   - String parsing and manipulation
//
// By centralizing string operations here, we avoid code duplication and maintain
// consistent string handling across the codebase while preventing namespace
// collisions with the standard library with the wrapper functions.
package strings

import "strings"

// CamelCase converts a string to camel case by capitalizing the first letter
// while preserving the rest of the string. This is useful for generating
// idiomatic Go identifiers from various input formats.
func CamelCase(in string) string {
	first := ToUpper(in[:1])
	rest := in[1:]
	return first + rest
}

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

// GetPlural returns both lowercase and camel case versions of a plural form
// for a given type name. It handles irregular plurals, compound words with
// underscores, and applies standard English pluralization rules.
func GetPlural(iotaType string) (string, string) {
	if l := len(iotaType); l == 0 {
		return "", ""
	}
	lower, camel := ToLower(iotaType), CamelCase(iotaType)
	if plural, ok := irregular[lower]; ok {
		if iotaType == lower {
			return plural, CamelCase(plural)
		} else {
			if ToUpper(iotaType) == iotaType {
				return ToUpper(plural), ToUpper(plural)
			} else {
				return plural, CamelCase(plural)
			}
		}
	}

	// Check for compound words (with underscore)
	if Contains(lower, "_") {
		parts := Split(lower, "_")
		lastPart := parts[len(parts)-1]
		pluralLastPart := ""

		// Pluralize the last part
		if plural, ok := irregular[lastPart]; ok {
			pluralLastPart = plural
		} else {
			pluralLastPart = regularPlural(lastPart)
		}

		// Reconstruct with pluralized last part
		parts[len(parts)-1] = pluralLastPart
		pluralLower := Join(parts, "_")
		return pluralLower, CamelCase(pluralLower)
	}

	return regularPlural(lower), regularPlural(camel)
}

// regularPlural applies standard English pluralization rules to a word,
// handling special cases like words ending in 'y', 's', 'x', 'z', 'o',
// 'ch', 'sh', or 'ss'.
func regularPlural(word string) string {
	l := len(word)
	if l == 0 {
		return word
	}
	if l > 1 && word[l-1] == 'y' {
		prev := word[l-2]
		if prev != 'a' && prev != 'e' && prev != 'i' && prev != 'o' && prev != 'u' {
			return word[:l-1] + "ies"
		}
	}
	lastChar := word[l-1]
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
// This is a wrapper around ToLower.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper converts a string to uppercase.
// This is a wrapper around ToUpper.
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// Contains reports whether substr is within s.
// This is a wrapper around Contains.
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TrimSpace returns a slice of the string with all leading
// and trailing white space removed.
// This is a wrapper around TrimSpace.
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Split slices s into all substrings separated by sep and returns them.
// This is a wrapper around Split.
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// TrimLeft returns a slice of the string with all leading
// characters contained in cutset removed.
// This is a wrapper around TrimLeft.
func TrimLeft(s, cutset string) string {
	return strings.TrimLeft(s, cutset)
}

// HasPrefix tests whether the string starts with prefix.
// This is a wrapper around HasPrefix.
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix tests whether the string ends with suffix.
// This is a wrapper around HasSuffix.
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Index returns the index of the first instance of sep in s,
// or -1 if sep is not present in s.
// This is a wrapper around Index.
func Index(s, sep string) int {
	return strings.Index(s, sep)
}

// Count counts the number of non-overlapping instances of sep in s.
// This is a wrapper around Count.
func Count(s, sep string) int {
	return strings.Count(s, sep)
}

func Join(s1 []string, s2 string) string {
	return strings.Join(s1, s2)
}
