package gofile

import (
	"fmt"
	"strconv"
	"text/template"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/strings"
)

type parseFunctionData struct {
	WrapperName string
	FailFast    bool
	Enums       []enum.Enum
}

var (
	parseFunctionStr = `
{{- if .FailFast}}
var ErrParse{{.WrapperName}} = errors.New("invalid input provided to parse to {{.WrapperName}}")
{{- end}}
// Parse{{.WrapperName}} parses the input value into an enum value.
// It returns the parsed enum value or an error if the input is invalid.
// It is a convenience function that can be used to parse enum values from
// various input types, such as strings, byte slices, or other enum types.
func Parse{{.WrapperName}}(input any) ({{.WrapperName}}, error) {
	switch v := input.(type) {
	case {{.WrapperName}}:
		return v, nil
	case string:
		if result := stringTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case fmt.Stringer:
		if result := stringTo{{.WrapperName}}(v.String()); result != nil {
			return *result, nil
		}
	case []byte:
		if result := stringTo{{.WrapperName}}(string(v)); result != nil {
			return *result, nil
		}
	case int:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case int8:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case int16:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case int32:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case int64:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case uint:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case uint8:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case uint16:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case uint32:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case uint64:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case float32:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	case float64:
		if result := numberTo{{.WrapperName}}(v); result != nil {
			return *result, nil
		}
	default:
		return invalid{{.WrapperName}}, fmt.Errorf("invalid type %T", input)
	}
	{{- if .FailFast}}
	return invalid{{.WrapperName}}, fmt.Errorf("%w: invalid value %v", ErrParse{{.WrapperName}}, input)
	{{- else}}
	return invalid{{.WrapperName}}, nil
	{{- end}}
}
`
	parseFunctionTemplate = template.Must(template.New("parseFunction").Parse(parseFunctionStr))
)

func (g *Writer) writeParseFunction(rep enum.GenerationRequest) {
	g.writeTemplate(parseFunctionTemplate, parseFunctionData{
		WrapperName: wrapperName(rep.EnumIota.Type),
		Enums:       rep.EnumIota.Enums,
		FailFast:    rep.Configuration.Failfast,
	})
}

type parseStringFunctionData struct {
	EnumNameMap     string
	WrapperName     string
	EnumType        string
	Enums           []enumDefinition
	CaseInsensitive bool
}

type enumDefinition struct {
	EnumNameIdentifier string
	EnumType           string
	IotaType           string
	EnumName           string
	Value              any
	Fields             []enum.Field
	Aliases            []string
	QuotedAliases      []string
	Valid              bool
}

var (
	parseStringFunctionStr = `
// {{ .EnumNameMap }} is a map of enum values to their {{.WrapperName}} representation
// It is used to convert string representations of enum values into their {{.WrapperName}} representation.
var {{.EnumNameMap}} = map[string]{{.WrapperName}}{
{{- range .Enums }}
    {{- $enum := . }}
    {{- range .QuotedAliases }}
    {{ . }}: {{ $.EnumType }}.{{ $enum.EnumNameIdentifier }},
    {{- end }}
{{- end }}
}

// stringTo{{.WrapperName}} converts a string representation of an enum value into its {{.WrapperName}} representation
// It returns a pointer to the {{.WrapperName}} representation of the enum value if the string is valid
// Otherwise, it returns nil
func stringTo{{.WrapperName}}(s string) *{{.WrapperName}} {
{{- if .CaseInsensitive }}
    s = strings.ToLower(s)
{{- end }}
    if t, ok := {{.EnumNameMap}}[s]; ok {
        return &t
    }
    return nil
}
`
	parseStringFunctionTemplate = template.Must(template.New("parseStringFunction").Parse(parseStringFunctionStr))
)

func (g *Writer) writeStringParsingMethod(rep renderRequest) {
	if err := validateAliasCollisions(rep.AllEnums, rep.Configuration.Insensitive); err != nil {
		g.templateErr = err
		return
	}
	enums := rep.ParseEnums
	g.writeTemplate(parseStringFunctionTemplate, parseStringFunctionData{
		WrapperName:     wrapperName(rep.EnumIota.Type),
		EnumNameMap:     enumNameMap(rep.EnumIota.Type),
		EnumType:        enumType(rep.GenerationRequest),
		Enums:           enums,
		CaseInsensitive: rep.Configuration.Insensitive,
	})
}

type parseNumberFunctionData struct {
	Constraints bool
	WrapperName string
	EnumType    string
	Enums       []enumDefinition
}

var (
	parseIntegerGenericFunctionTemplate = template.Must(template.New("parseIntegerGenericFunction").Parse(`

// numberTo{{.WrapperName}} converts a numeric value to a {{.WrapperName}}
// It returns a pointer to the {{.WrapperName}} representation of the enum value if the numeric value is valid
// Otherwise, it returns nil
{{- if .Constraints }}
func numberTo{{.WrapperName}}[T interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64
}](num T) *{{.WrapperName}} {
{{- else }}
func numberTo{{.WrapperName}}[T constraints.Integer | constraints.Float](num T) *{{.WrapperName}} {
{{- end }}
    f := float64(num)
    if math.Floor(f) != f {
        return nil
    }
    switch f {
    {{- range .Enums }}
    case {{ .Value }}:
        result := {{ $.EnumType }}.{{ .EnumNameIdentifier }}
        if !result.IsValid() {
            return nil
        }
        return &result
    {{- end }}
    default:
        return nil
    }
}

`))
)

func enumNameMap(enumType string) string {
	return fmt.Sprintf("%sNameMap", strings.Pluralise(enumType))
}

func quotedAliases(aliases []string) []string {
	quoted := make([]string, len(aliases))
	for i, alias := range aliases {
		quoted[i] = strconv.Quote(alias)
	}
	return quoted
}

func validateAliasCollisions(enums []enumDefinition, insensitive bool) error {
	owners := make(map[string]string)
	for _, enumDef := range enums {
		for _, alias := range enumDef.Aliases {
			key := alias
			if insensitive {
				key = strings.ToLower(alias)
			}
			if owner, ok := owners[key]; ok && owner != enumDef.EnumNameIdentifier {
				return fmt.Errorf("duplicate enum alias %q used by both %s and %s", alias, owner, enumDef.EnumNameIdentifier)
			}
			owners[key] = enumDef.EnumNameIdentifier
		}
	}
	return nil
}

func lowercaseAliases(aliases []string) []string {
	lowercase := make([]string, len(aliases))
	for i, alias := range aliases {
		lowercase[i] = strings.ToLower(alias)
	}
	return lowercase
}
