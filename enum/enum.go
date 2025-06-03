// Package enum defines the core interfaces and data structures for enum representation.
//
// This package forms the foundation of the goenums system by defining the domain model
// for enum types and values, along with the interfaces that components must implement
// to participate in the enum generation pipeline.
//
// The interfaces follow a clear separation of concerns:
//   - Parser: Extracts enum definitions from source content
//   - Writer: Generates output artifacts from enum representations
//   - Source: Provides raw content for parsing
//
// This design enables a modular system where different input formats and output targets
// can be supported without modifying the core workflow.
package enum

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/zarldev/goenums/generator/config"
)

// Parser defines the contract for components that convert source content into
// enum representations. Implementations of this interface analyze source code or other
// input formats and extract structured information about enum types and values.
// Different implementations can support various input languages or formats while
// producing the same standardized Representation output.
type Parser interface {
	// Parse analyzes the content from a source and returns structured enum representations.
	// It transforms input data into a format-agnostic model that can be used
	// for code generation. The context allows for cancellation and timeout control.
	Parse(ctx context.Context) ([]GenerationRequest, error)
}

// GenerationRequest represents a request to generate an enum implementation.
// It contains all the information needed to generate the implementation,
// including the package name, imports, enum type and value information,
// and configuration options.
type GenerationRequest struct {
	Package        string
	Imports        []string
	EnumIota       EnumIota
	Version        string
	SourceFilename string
	OutputFilename string
	Configuration  config.Configuration
}

func (e *GenerationRequest) IsValid() bool {
	return e.Package != "" &&
		e.EnumIota.Type != "" &&
		e.Version != "" &&
		e.SourceFilename != ""
}

// Handlers represents the configuration options for interface generation.
// It controls which standard Go interfaces will be implemented by the generated enum types.
//
// Each boolean flag enables generation of the corresponding interface implementation:
//   - JSON: encoding/json.Marshaler and encoding/json.Unmarshaler
//   - Text: encoding.TextMarshaler and encoding.TextUnmarshaler
//   - YAML: gopkg.in/yaml.v3 marshaling interfaces
//   - SQL: database/sql/driver.Scanner and database/sql/driver.Valuer
//   - Binary: encoding.BinaryMarshaler and encoding.BinaryUnmarshaler
//
// Example usage:
//
//	handlers := Handlers{
//	    JSON: true,  // Enable JSON marshaling
//	    SQL:  true,  // Enable database integration
//	    Text: false, // Disable text marshaling
//	}
type Handlers struct {
	// JSON enables JSON marshaling and unmarshaling interface implementations
	JSON bool
	// Text enables text marshaling and unmarshaling interface implementations
	Text bool
	// YAML enables YAML marshaling and unmarshaling interface implementations
	YAML bool
	// SQL enables database Scanner and Valuer interface implementations
	SQL bool
	// Binary enables binary marshaling and unmarshaling interface implementations
	Binary bool
}

// Command returns the command string for the enum generation process.
// It constructs the command string based on the configuration options.
func (r GenerationRequest) Command() string {
	var b bytes.Buffer
	b.WriteString("goenums")

	// Add flags
	if r.Configuration.Failfast {
		b.WriteString(" -f")
	}
	if r.Configuration.Legacy {
		b.WriteString(" -l")
	}
	if r.Configuration.Insensitive {
		b.WriteString(" -i")
	}
	if r.Configuration.Constraints {
		b.WriteString(" -c")
	}
	if r.Configuration.Verbose {
		b.WriteString(" -vv")
	}
	if r.Configuration.OutputFormat != "" && r.Configuration.OutputFormat != "go" {
		b.WriteString(" -o ")
		b.WriteString(r.Configuration.OutputFormat)
	}

	// Add source filename
	if r.SourceFilename != "" {
		b.WriteString(" ")
		b.WriteString(r.SourceFilename)
	}

	return b.String()
}

// EnumIota represents a complete enum type definition extracted from source code.
// It contains all the information needed to generate enum implementations,
// including the type name, associated fields, and individual enum values.
//
// The struct captures both the metadata about the enum type (like custom fields
// and formatting preferences) and the actual enum values that will be generated.
//
// Example usage:
//
//	enumIota := EnumIota{
//	    Type: "Status",
//	    Comment: "Status represents the state of an operation",
//	    Fields: []Field{{Name: "description", Value: ""}},
//	    Enums: []Enum{
//	        {Name: "Active", Index: 0, Valid: true},
//	        {Name: "Inactive", Index: 1, Valid: true},
//	    },
//	}
type EnumIota struct {
	// Type is the name of the enum type (e.g., "Status", "Color")
	Type string
	// Comment is the documentation comment associated with the enum type
	Comment string
	// Fields defines custom fields that each enum value can have
	Fields []Field
	// Opener is the opening delimiter for field values (e.g., "[", "(")
	Opener string
	// Closer is the closing delimiter for field values (e.g., "]", ")")
	Closer string
	// StartIndex is the starting index for enum values (usually 0)
	StartIndex int
	// Enums contains all the individual enum values for this type
	Enums []Enum
}

// Field represents a custom field that can be associated with enum values.
// Fields allow enums to carry additional metadata beyond just their name and index.
//
// For example, an enum representing HTTP status codes might have fields for
// the status message and whether it represents an error condition.
//
// Example usage:
//
//	field := Field{
//	    Name:  "message",
//	    Value: "Operation completed successfully",
//	}
//
// The Value can be any Go type that can be parsed from string representation
// in source code comments, including strings, numbers, booleans, and time values.
type Field struct {
	// Name is the identifier for this field (e.g., "description", "code", "priority")
	Name string
	// Value is the actual value for this field, which can be any supported Go type
	Value any
}

func (f *Field) Valid() bool {
	return f.Name != "" && f.Value != nil
}

// Enum represents a single enum value within an enum type.
// It contains the value's name, position, any custom fields, and validation status.
//
// Each Enum corresponds to one constant in the original iota-based declaration,
// but enriched with additional metadata that enables type-safe operations
// and enhanced functionality.
//
// Example usage:
//
//	enum := Enum{
//	    Name:    "Active",
//	    Index:   0,
//	    Fields:  []Field{{Name: "description", Value: "Currently active"}},
//	    Aliases: []string{"Running", "Enabled"},
//	    Valid:   true,
//	}
type Enum struct {
	// Name is the identifier for this enum value (e.g., "Active", "Pending")
	Name string
	// Index is the numeric position of this enum in the sequence (0-based)
	Index int
	// Fields contains any custom field values associated with this enum
	Fields []Field
	// Aliases are alternative names that can be used to reference this enum
	Aliases []string
	// Valid indicates whether this enum value is valid in the set of enums
	Valid bool
}

// Source abstracts the origin of input content to be parsed for enum definitions.
// This interface decouples the parsing logic from the specific location or format
// of the input data, allowing for flexible input sources.
type Source interface {
	// Content returns the raw bytes to be parsed for enum definitions.
	// This method retrieves the complete content from whatever backing store
	// or input mechanism the implementation uses.
	Content() ([]byte, error)

	// Filename returns an identifier for the source, typically a file path.
	// Even for non-file sources, this should return a meaningful identifier
	Filename() string
}

// Writer defines the contract for components that transform enum representations
// into output artifacts. Implementations of this interface generate code or other
// output formats based on the standardized Representation model. This interface
// completes the pipeline from input source to output artifact.
type Writer interface {
	// Write generates output artifacts from enum representations.
	// It transforms the format-agnostic model into concrete output such as
	// source code files. The context allows for cancellation and timeout control.
	Write(ctx context.Context, enums []GenerationRequest) error
}

func ParseEnumAliases(s string) []string {
	if !strings.Contains(s, ",") {
		// Handle single case without slice allocation
		trimmed := strings.TrimSpace(s)
		if len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' {
			trimmed = trimmed[1 : len(trimmed)-1]
		}
		return []string{trimmed}
	}

	aliases := strings.Split(s, ",")
	// Process in-place to avoid second allocation
	j := 0
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		if len(alias) == 0 {
			continue
		}
		if len(alias) >= 2 && alias[0] == '"' && alias[len(alias)-1] == '"' {
			alias = alias[1 : len(alias)-1]
		}
		aliases[j] = alias
		j++
	}
	return aliases[:j] // Return slice of actual length
}

var (
	ErrFieldEmptyValue = errors.New("empty field value")
)

func ParseEnumFields(s string, enumIota EnumIota) ([]Field, error) {
	fieldValues := strings.Split(s, ",")
	if len(fieldValues) == 1 && fieldValues[0] == "" {
		return []Field{}, nil
	}

	fcount := len(fieldValues)
	minLen := min(fcount, len(enumIota.Fields))

	// Use capacity, append as we go - this is the big win
	enumFields := make([]Field, 0, minLen)

	for i := range minLen {
		valRaw := strings.TrimSpace(fieldValues[i])
		if valRaw == "" {
			return []Field{}, ErrFieldEmptyValue
		}

		val, err := ParseValue(valRaw, enumIota.Fields[i].Value)
		if err != nil {
			return []Field{}, err
		}

		if val == nil {
			return []Field{}, ErrFieldEmptyValue
		}

		if str, ok := val.(string); ok && str == "" {
			return []Field{}, ErrFieldEmptyValue
		}

		fie := Field{
			Name:  enumIota.Fields[i].Name,
			Value: val,
		}
		if fie.Valid() {
			enumFields = append(enumFields, fie) // Only append valid fields
		}
	}
	return enumFields, nil
}

var (
	ErrParseValue      = errors.New("failed to parse value")
	ErrParseSource     = errors.New("failed to parse source")
	ErrNoEnumsFound    = errors.New("no valid enums found")
	ErrWriteOutput     = errors.New("failed to write output")
	ErrUnsupportedType = errors.New("unsupported type")
)

func ParseValue[T any](valRaw string, defaultVal T) (T, error) {
	var zero T
	switch any(defaultVal).(type) {
	case bool:
		val, err := strconv.ParseBool(valRaw)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(val).(T); ok {
			return v, nil
		}
	case float64:
		val, err := strconv.ParseFloat(valRaw, 64)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(val).(T); ok {
			return v, nil
		}
	case float32:
		val, err := strconv.ParseFloat(valRaw, 32)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(float32(val)).(T); ok {
			return v, nil
		}
	case int:
		val, err := strconv.Atoi(valRaw)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(val).(T); ok {
			return v, nil
		}
	case int64:
		val, err := strconv.ParseInt(valRaw, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(val).(T); ok {
			return v, nil
		}
	case int32:
		val, err := strconv.ParseInt(valRaw, 10, 32)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(int32(val)).(T); ok {
			return v, nil
		}
	case int16:
		val, err := strconv.ParseInt(valRaw, 10, 16)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(int16(val)).(T); ok {
			return v, nil
		}
	case int8:
		val, err := strconv.ParseInt(valRaw, 10, 8)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(int8(val)).(T); ok {
			return v, nil
		}
	case uint:
		val, err := strconv.ParseUint(valRaw, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(uint(val)).(T); ok {
			return v, nil
		}
	case uint64:
		val, err := strconv.ParseUint(valRaw, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(val).(T); ok {
			return v, nil
		}
	case uint32:
		val, err := strconv.ParseUint(valRaw, 10, 32)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(uint32(val)).(T); ok {
			return v, nil
		}
	case uint16:
		val, err := strconv.ParseUint(valRaw, 10, 16)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(uint16(val)).(T); ok {
			return v, nil
		}
	case uint8:
		val, err := strconv.ParseUint(valRaw, 10, 8)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(uint8(val)).(T); ok {
			return v, nil
		}
	case string:
		if len(valRaw) >= 2 && valRaw[0] == '"' && valRaw[len(valRaw)-1] == '"' {
			if v, ok := any(valRaw[1 : len(valRaw)-1]).(T); ok {
				return v, nil
			}
		}
		if v, ok := any(valRaw).(T); ok {
			return v, nil
		}
	case time.Time:
		val, err := time.Parse(time.RFC3339, valRaw)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(val).(T); ok {
			return v, nil
		}
	case time.Duration:
		val, err := time.ParseDuration(valRaw)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrParseValue, err)
		}
		if v, ok := any(val).(T); ok {
			return v, nil
		}
	default:
		return zero, fmt.Errorf("%w: %w", ErrParseValue, ErrUnsupportedType)
	}
	return zero, fmt.Errorf("%w: %w", ErrParseValue, ErrUnsupportedType)
}

func ExtractImports(enumIotas []EnumIota) []string {
	totalFields := 0
	for _, enumIota := range enumIotas {
		totalFields += len(enumIota.Fields)
	}
	imports := make([]string, 0, totalFields)
	for _, enumIota := range enumIotas {
		for _, field := range enumIota.Fields {
			str := fmt.Sprintf("%T", field.Value)
			if strings.Contains(str, ".") {
				imports = append(imports, strings.Split(str, ".")[0])
			}
		}
	}
	slices.Sort(imports)
	return slices.Compact(imports)
}

func ExtractFields(comment string) (string, string, []Field) {
	fields := make([]Field, 0)
	comment = strings.TrimSpace(comment)
	open, closer := " ", " "
	if comment == "" {
		return open, closer, fields
	}
	fieldVals := strings.Split(comment, ",")
	for _, val := range fieldVals {
		field := strings.TrimSpace(val)
		open, closer = OpenCloser(field)

		var nO, nC, tO, tC int
		var n, f string

		if open == " " {
			extra := strings.Split(field, " ")
			if len(extra) > 1 {
				n = extra[0]
				f = extra[1]
			} else {
				f = extra[0]
			}
			fields = append(fields, Field{
				Name:  n,
				Value: FieldToType(f),
			})
			continue
		}

		nO = strings.Index(field, open)
		if nO == -1 {
			continue
		}
		nC = strings.Index(field[nO:], closer) + nO
		if nC == -1 {
			continue
		}
		tO = nO + len(open)
		tC = nC

		if nO < 0 || tO > len(field) || tC > len(field) || tO > tC {
			continue
		}
		n = field[:nO]
		f = field[tO:tC]
		fields = append(fields, Field{
			Name:  n,
			Value: FieldToType(f),
		})
	}
	return open, closer, fields
}

func OpenCloser(field string) (string, string) {
	open := " "
	closer := " "
	if strings.Contains(field, "[") {
		open = "["
		closer = "]"
	} else if strings.Contains(field, "(") {
		open = "("
		closer = ")"
	}
	return open, closer
}

func FieldToType(field string) any {
	f := strings.TrimSpace(field)
	switch f {
	case "bool":
		return false
	case "int":
		return 0
	case "string":
		return ""
	case "time.Duration":
		return time.Duration(0)
	case "time.Time":
		return time.Time{}
	case "float64":
		return 0.0
	case "float32":
		return float32(0.0)
	case "int64":
		return int64(0)
	case "int32":
		return int32(0)
	case "int16":
		return int16(0)
	case "int8":
		return int8(0)
	case "uint64":
		return uint64(0)
	case "uint32":
		return uint32(0)
	case "uint16":
		return uint16(0)
	case "uint8":
		return uint8(0)
	case "uint":
		return uint(0)
	case "byte":
		return byte(0)
	case "rune":
		return rune(0)
	case "complex64":
		return complex64(0)
	case "complex128":
		return complex128(0)
	case "uintptr":
		return uintptr(0)
	default:
		return nil
	}
}
