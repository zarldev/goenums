// Package strings provides specialized string handling utilities for enum generation.
//
// This package wraps standard library string functions and adds custom functionality
// specifically tailored for enum code generation, including:
//   - Case conversion (upper, lower, camel case)
//   - Intelligent pluralization with irregularToPlurals word handling
//   - String parsing and manipulation
//
// By centralizing string operations here, we avoid code duplication and maintain
// consistent string handling across the codebase while preventing namespace
// collisions with the standard library.
package strings

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// irregularToPlural contains mappings for words that don't follow standard English
// pluralization rules, ensuring correct pluralization for special cases.
var irregularToPlural = map[string]string{
	"man":        "men",
	"woman":      "women",
	"child":      "children",
	"foot":       "feet",
	"tooth":      "teeth",
	"goose":      "geese",
	"mouse":      "mice",
	"ox":         "oxen",
	"person":     "people",
	"index":      "indices",
	"matrix":     "matrices",
	"vertex":     "vertices",
	"datum":      "data",
	"medium":     "media",
	"analysis":   "analyses",
	"crisis":     "crises",
	"status":     "statuses",
	"alias":      "aliases",
	"basis":      "bases",
	"criterion":  "criteria",
	"phenomenon": "phenomena",
	"syllabus":   "syllabi",
	"thesis":     "theses",
	"bus":        "buses",
	"glass":      "glasses",
}

var irregularPluralsToSingular = map[string]string{
	"men":       "man",
	"women":     "woman",
	"children":  "child",
	"feet":      "foot",
	"teeth":     "tooth",
	"geese":     "goose",
	"mice":      "mouse",
	"oxen":      "ox",
	"people":    "person",
	"indices":   "index",
	"matrices":  "matrix",
	"vertices":  "vertex",
	"data":      "datum",
	"media":     "medium",
	"analyses":  "analysis",
	"crises":    "crisis",
	"statuses":  "status",
	"aliases":   "alias",
	"bases":     "basis",
	"criteria":  "criterion",
	"phenomena": "phenomenon",
	"syllabi":   "syllabus",
	"theses":    "thesis",
	"buses":     "bus",
	"glasses":   "glass",
}

func SplitBySpace(input string) (string, string) {
	if strings.Contains(input, "\"") {
		inQuote := false
		for i, char := range input {
			if char == '"' {
				inQuote = !inQuote
			} else if char == ' ' && !inQuote {
				return input[:i], input[i+1:]
			}
		}
		return input, ""
	}
	before, after, found := strings.Cut(input, " ")
	if !found {
		return before, ""
	}
	return before, after
}

// detectCase returns a function that applies original case from src to the target string
func detectCase(src string) func(string) string {
	if src == strings.ToUpper(src) {
		// ALL UPPER
		return strings.ToUpper
	}
	if src == strings.ToLower(src) {
		// all lower
		return strings.ToLower
	}
	// Capitalized (Title Case)
	if len(src) > 0 && unicode.IsUpper(rune(src[0])) && src[1:] == strings.ToLower(src[1:]) {
		return func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
		}
	}
	isUpper := make([]bool, len(src))
	for i, r := range src {
		if unicode.IsUpper(r) {
			isUpper[i] = true
		}
	}
	return func(s string) string {
		for i, r := range s {
			if i < len(isUpper) && isUpper[i] {
				s = s[:i] + strings.ToUpper(string(r)) + s[i+1:]
			}
		}
		return s
	}
}

func alreadyPlural(s string) bool {
	if _, ok := irregularPluralsToSingular[s]; ok {
		return true
	}
	if _, ok := irregularPluralsToSingular[strings.ToLower(s)]; ok {
		return true
	}
	if _, ok := irregularToPlural[strings.ToLower(s)]; ok {
		return false
	}
	if strings.Contains(s, "_") || strings.Contains(s, "-") || strings.Contains(s, " ") {
		var lastPart string
		switch {
		case strings.Contains(s, "_"):
			parts := strings.Split(s, "_")
			lastPart = parts[len(parts)-1]
		case strings.Contains(s, "-"):
			parts := strings.Split(s, "-")
			lastPart = parts[len(parts)-1]
		case strings.Contains(s, " "):
			parts := strings.Split(s, " ")
			lastPart = parts[len(parts)-1]
		}
		lowerLast := strings.ToLower(lastPart)
		if _, ok := irregularPluralsToSingular[lowerLast]; ok {
			return true
		}
		if strings.HasSuffix(lowerLast, "s") && !strings.HasSuffix(lowerLast, "ss") &&
			!strings.HasSuffix(lowerLast, "us") && !strings.HasSuffix(lowerLast, "is") {
			return true
		}
		if strings.HasSuffix(lowerLast, "es") {
			return true
		}
		return false
	}

	// Check if it's a regular plural (ends with s but not ss, us, is)
	lower := strings.ToLower(s)
	if strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss") &&
		!strings.HasSuffix(lower, "us") && !strings.HasSuffix(lower, "is") {
		return true
	}
	if strings.HasSuffix(lower, "es") {
		return true
	}
	return false
}

func IsPlural(s string) bool {
	return alreadyPlural(s)
}

func Singularise(iotaType string) string {
	if iotaType == "" {
		return ""
	}
	applyCase := detectCase(iotaType)
	if !alreadyPlural(iotaType) {
		return iotaType
	}
	// Handle snake_case
	if strings.Contains(iotaType, "_") {
		parts := strings.Split(iotaType, "_")
		// singularize last part only
		last := parts[len(parts)-1]
		lowerLast := strings.ToLower(last)
		var singularLast string
		if s, ok := irregularPluralsToSingular[lowerLast]; ok {
			singularLast = s
		} else {
			singularLast = Singularize(last)
		}
		// Apply case to the singularized part
		lastCase := detectCase(last)
		parts[len(parts)-1] = lastCase(singularLast)
		return strings.Join(parts, "_")
	}
	// Handle spaces
	if strings.Contains(iotaType, " ") {
		parts := strings.Split(iotaType, " ")
		// singularize last part only
		last := parts[len(parts)-1]
		lowerLast := strings.ToLower(last)
		var singularLast string
		if s, ok := irregularPluralsToSingular[lowerLast]; ok {
			singularLast = s
		} else {
			singularLast = Singularize(last)
		}
		// Apply case to the singularized part
		lastCase := detectCase(last)
		parts[len(parts)-1] = lastCase(singularLast)
		return strings.Join(parts, " ")
	}
	// Handle kebab-case
	if strings.Contains(iotaType, "-") {
		parts := strings.Split(iotaType, "-")
		// singularize last part only
		last := parts[len(parts)-1]
		lowerLast := strings.ToLower(last)
		var singularLast string
		if s, ok := irregularPluralsToSingular[lowerLast]; ok {
			singularLast = s
		} else {
			singularLast = Singularize(last)
		}
		// Apply case to the singularized part
		lastCase := detectCase(last)
		parts[len(parts)-1] = lastCase(singularLast)
		return strings.Join(parts, "-")
	}
	// Handle normal case
	lowerIotaType := strings.ToLower(iotaType)
	if s, ok := irregularPluralsToSingular[lowerIotaType]; ok {
		return applyCase(s)
	}
	return applyCase(Singularize(lowerIotaType))
}

func Singularize(word string) string {
	word = strings.TrimSpace(word)
	if _, ok := irregularPluralsToSingular[word]; ok {
		return irregularPluralsToSingular[word]
	}
	if len(word) > 1 && word[len(word)-1] == 's' {
		return word[:len(word)-1]
	}
	if len(word) > 2 && word[len(word)-2:] == "es" {
		return word[:len(word)-2]
	}
	if len(word) > 2 && word[len(word)-2:] == "ies" {
		return word[:len(word)-3] + "y"
	}
	if len(word) > 1 && word[len(word)-1] == 'x' {
		return word + "es"
	}
	return word
}

// Plural pluralizes a word or snake_case word with case preservation
func Plural(iotaType string) string {
	if iotaType == "" {
		return ""
	}

	applyCase := detectCase(iotaType)

	if alreadyPlural(iotaType) {
		return iotaType
	}

	// Handle snake_case
	if strings.Contains(iotaType, "_") {
		parts := strings.Split(iotaType, "_")
		// pluralize last part only
		last := parts[len(parts)-1]

		lowerLast := strings.ToLower(last)
		var pluralLast string
		if p, ok := irregularToPlural[lowerLast]; ok {
			pluralLast = p
		} else {
			pluralLast = regularPlural(lowerLast)
		}
		parts[len(parts)-1] = applyCase(pluralLast)

		return strings.Join(parts, "_")
	}

	lower := strings.ToLower(iotaType)
	var plural string

	if p, ok := irregularToPlural[lower]; ok {
		plural = p
	} else {
		plural = regularPlural(lower)
	}

	return applyCase(plural)
}

// splitWords splits a CamelCase or PascalCase or ALLCAPS string into components
func splitWords(s string) []string {
	var words []string
	var current strings.Builder

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && (i+1 < len(s) && unicode.IsLower(rune(s[i+1]))) {
			words = append(words, current.String())
			current.Reset()
		}
		current.WriteRune(r)
	}
	words = append(words, current.String())
	return words
}

// matchCasing copies the casing pattern from src to dst
func matchCasing(src, dst string) string {
	if src == strings.ToUpper(src) {
		return strings.ToUpper(dst)
	}
	if src == strings.ToLower(src) {
		return strings.ToLower(dst)
	}
	// Copy casing per character
	srcRunes := []rune(src)
	dstRunes := []rune(dst)
	for i := 0; i < len(srcRunes) && i < len(dstRunes); i++ {
		if unicode.IsUpper(srcRunes[i]) {
			dstRunes[i] = unicode.ToUpper(dstRunes[i])
		} else {
			dstRunes[i] = unicode.ToLower(dstRunes[i])
		}
	}
	return string(dstRunes)
}

func Singular(input string) string {
	words := splitWords(input)
	if len(words) == 0 {
		return input
	}

	last := words[len(words)-1]
	lower := strings.ToLower(last)

	var singular string
	for s, p := range irregularToPlural {
		if lower == p {
			singular = s
			break
		}
	}
	if singular == "" && IsRegularPlural(lower) {
		singular = lower[:len(lower)-1]
	}
	if singular == "" {
		singular = lower
	}
	singularCased := matchCasing(last, singular)
	words[len(words)-1] = singularCased
	return strings.Join(words, "")
}

func IsRegularPlural(word string) bool {
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
	in = strings.TrimSpace(in)
	if strings.Contains(in, "_") {
		return camel(in, "_")
	}
	if strings.Contains(in, " ") {
		return camel(in, " ")
	}
	if strings.Contains(in, "-") {
		return camel(in, "-")
	}
	if strings.Contains(in, "-") {
		return camel(in, "-")
	}
	if strings.Contains(in, "—") {
		return camel(in, "—")
	}
	if strings.Contains(in, "–") {
		return camel(in, "–")
	}
	if strings.Contains(in, ".") {
		return camel(in, ".")
	}
	runes := []rune(in)
	if len(runes) > 0 {
		return string(unicode.ToUpper(runes[0])) + string(runes[1:])
	}
	return in
}

func camel(in string, sp string) string {
	parts := strings.Split(in, sp)
	var result strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		if len(runes) > 0 {
			result.WriteRune(unicode.ToUpper(runes[0]))
			result.WriteString(string(runes[1:]))
		}
	}
	return result.String()
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

func TrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}
func TrimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
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

// TrimRight returns a slice of the string with all trailing
// characters contained in cutset removed.
// This is a wrapper around strings.TrimRight.
func TrimRight(s, cutset string) string {
	return strings.TrimRight(s, cutset)
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

// LastIndex returns the index of the last instance of sep in s, or -1 if sep is not present in s.
// This is a wrapper around strings.LastIndex.
func LastIndex(s, sep string) int {
	return strings.LastIndex(s, sep)
}

// ReplaceAll returns a copy of the string s with all non-overlapping instances of old replaced by new.
// If old is empty, it matches at the beginning of the string and after each UTF-8 sequence, yielding up to k+1 replacements for a len(s) = k string.
// If new is empty, it removes all instances of old from s.
// This is a wrapper around strings.ReplaceAll.
func ReplaceAll(s, o, n string) string {
	return strings.ReplaceAll(s, o, n)
}

// EnumBuilder is a specialized string builder for constructing enum-related content.
// It wraps the standard strings.Builder with enum-specific functionality and provides
// a consistent interface for building strings during enum generation.
//
// EnumBuilder is designed to be efficient for the common patterns used in enum
// code generation, such as building method bodies, type definitions, and documentation.
//
// Example usage:
//
//	var builder EnumBuilder
//	builder.WriteString("func (e Status) String() string {\n")
//	builder.WriteString("    return statusNames[e]\n")
//	builder.WriteString("}")
//	code := builder.String()
type EnumBuilder struct {
	b *strings.Builder
}

func (b *EnumBuilder) Write(p []byte) (int, error) {
	if b.b == nil {
		b.b = &strings.Builder{}
	}
	return b.b.Write(p)
}

// EnumWriter wraps an io.Writer to provide enum-specific writing functionality.
// It serves as an adapter that can be configured with different output destinations
// while maintaining a consistent interface for enum code generation.
//
// The EnumWriter is typically used in code generation pipelines where the output
// destination might vary (files, buffers, stdout, etc.) but the writing interface
// should remain the same.
//
// Example usage:
//
//	writer := NewEnumWriter(WithWriter(os.Stdout))
//	writer.Write([]byte("generated enum code"))
type EnumWriter struct {
	io.Writer
}

func NewEnumWriter(opts ...WriterOption) *EnumWriter {
	e := &EnumWriter{
		Writer: os.Stdout,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (w *EnumWriter) Write(p []byte) (int, error) {
	return w.Writer.Write(p)
}

type WriterOption func(*EnumWriter)

func WithWriter(w io.Writer) WriterOption {
	return func(e *EnumWriter) {
		e.Writer = w
	}
}

func AsType(v any) string {
	return fmt.Sprintf("%T", v)
}

// WriteString writes the string s to the EnumBuilder
// It is a wrapper around strings.Builder.WriteString.
func (b *EnumBuilder) WriteString(s string) {
	if b.b == nil {
		b.b = &strings.Builder{}
	}
	_, _ = b.b.WriteString(s)
}

// String returns the accumulated string.
// It is a wrapper around strings.Builder.String.
func (b *EnumBuilder) String() string {
	if b.b == nil {
		return ""
	}
	return b.b.String()
}

// Len returns the number of accumulated bytes.
// It is a wrapper around strings.Builder.Len.
func (b *EnumBuilder) Len() int {
	if b.b == nil {
		return 0
	}
	return b.b.Len()
}

// Reset resets the EnumBuilder to be empty.
// It is a wrapper around strings.Builder.Reset.
func (b *EnumBuilder) Reset() {
	if b.b == nil {
		return
	}
	b.b.Reset()
}

// Grow grows the EnumBuilder's capacity, if necessary, to guarantee space for
// another n bytes. It is a wrapper around strings.Builder.Grow.
func (b *EnumBuilder) Grow(n int) {
	if b.b == nil {
		b.b = &strings.Builder{}
	}
	b.b.Grow(n)
}

// WriteByte appends the byte c to the EnumBuilder.
// It is a wrapper around strings.Builder.WriteByte.
func (b *EnumBuilder) WriteByte(c byte) error {
	if b.b == nil {
		b.b = &strings.Builder{}
	}
	return b.b.WriteByte(c)
}

func Pluralise(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return s + "s"
	}

	// Check if already plural
	if isPlural(s) {
		return s
	}
	lower := strings.ToLower(s)
	if p, ok := irregularToPlural[lower]; ok {
		// Apply original case pattern
		if s == strings.ToUpper(s) {
			return strings.ToUpper(p)
		}
		if s == strings.ToLower(s) {
			return p
		}
		// Title case or mixed case
		return matchCasing(s, p)
	}

	// Use regular pluralization
	plural := regularPlural(lower)
	// Apply original case pattern
	if s == strings.ToUpper(s) {
		return strings.ToUpper(plural)
	}
	if s == strings.ToLower(s) {
		return plural
	}
	return matchCasing(s, plural)
}

func isPlural(s string) bool {
	if isIrregularPlural(s) {
		return true
	}
	return isRegularPlural(s)
}

func isIrregularPlural(s string) bool {
	_, ok := irregularPluralsToSingular[strings.ToLower(s)]
	return ok
}

func isRegularPlural(s string) bool {
	if len(s) < 2 {
		return false
	}

	lower := strings.ToLower(s)

	// Check if it's an irregular word that's not actually plural
	if _, ok := irregularToPlural[lower]; ok {
		return false
	}

	// Words ending in 'ss', 'us', 'is' are usually not plural
	if strings.HasSuffix(lower, "ss") || strings.HasSuffix(lower, "us") || strings.HasSuffix(lower, "is") {
		return false
	}

	return strings.HasSuffix(lower, "s") ||
		strings.HasSuffix(lower, "es") ||
		strings.HasSuffix(lower, "ies")
}

func Camel(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return strings.ToUpper(s)
	}
	c := unicode.ToUpper(rune(s[0]))
	return string(c) + s[1:]
}

func Lower1stCharacter(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return strings.ToLower(s)
	}
	c := unicode.ToLower(rune(s[0]))
	return string(c) + s[1:]
}

// Integer defines a type constraint for all integer types.
// This interface uses Go 1.18+ type constraints to represent any integer type,
// including both signed and unsigned variants of all sizes.
//
// It's used in generic functions that need to work with any integer type
// while maintaining type safety and avoiding runtime type assertions.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Float defines a type constraint for floating-point types.
// This interface represents both float32 and float64 types using Go's
// type constraint syntax.
//
// It enables generic functions to work with floating-point numbers
// without losing type information or requiring separate implementations.
type Float interface {
	~float32 | ~float64
}

// Number defines a type constraint that encompasses all numeric types.
// It combines both Integer and Float constraints to represent any numeric
// type that can be used in mathematical operations.
//
// This is particularly useful for generic functions that need to format
// or manipulate numeric values regardless of their specific type.
type Number interface {
	Integer | Float
}

func Ify(v any) string {
	switch v := v.(type) {
	case string:
		return `"` + v + `"`
	case []rune:
		return `"` + string(v) + `"`
	case []byte:
		return `"` + string(v) + `"`
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	case time.Duration:
		hrs := v.Hours()
		mins := v.Minutes()
		secs := v.Seconds()
		var b strings.Builder
		if hrs != 0 {
			b.WriteString("time.Hour * ")
			b.WriteString(strconv.FormatFloat(float64(hrs), 'f', -1, 64))
			return b.String()
		}
		if mins != 0 {
			b.WriteString("time.Minute * ")
			b.WriteString(strconv.FormatFloat(float64(mins), 'f', -1, 64))
			return b.String()
		}
		if secs != 0 {
			b.WriteString("time.Second * ")
			b.WriteString(strconv.FormatFloat(float64(secs), 'f', -1, 64))
			return b.String()
		}
		return b.String()
	case fmt.Stringer:
		return `"` + v.String() + `"`
	default:
		return fmt.Sprintf("%v", v)
	}
}

// IfiableNumeric is a generic version that can be used when type is known at compile time
func IfiableNumeric[T Number](n T) string {
	f := float64(n)
	absF := math.Abs(f)

	if absF >= 1e6 || (absF < 1e-6 && absF > 0) {
		return strconv.FormatFloat(f, 'e', 2, 64)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}
