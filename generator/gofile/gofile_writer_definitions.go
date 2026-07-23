package gofile

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/naming"
	"github.com/zarldev/goenums/strings"
)

func (g *Writer) writeCompileCheck(rep enum.GenerationRequest) {
	enums := make([]compileCheckEnum, len(rep.EnumIota.Enums))
	maxValue := 0

	// Find the maximum value to determine array size
	for _, e := range rep.EnumIota.Enums {
		if e.Value > maxValue {
			maxValue = e.Value
		}
	}

	for i, e := range rep.EnumIota.Enums {
		// Normalize all indices to 0 by subtracting the value
		indexExpr := e.Name
		if e.Value != 0 {
			if e.Value > 0 {
				indexExpr = fmt.Sprintf("%s-%s", e.Name, strconv.Itoa(e.Value))
			} else {
				indexExpr = fmt.Sprintf("%s+%s", e.Name, strconv.Itoa(-e.Value))
			}
		}

		enums[i] = compileCheckEnum{
			Name:      e.Name,
			Value:     e.Value,
			IndexExpr: indexExpr,
		}
	}
	g.writeTemplate(compileCheckTemplate, compileCheckData{
		ArraySize: maxValue + 1,
		Enums:     enums,
	})
}

var (
	stringMethodStr = `
// {{ .EnumLower }}Names is a constant string containing the canonical names for all enum values.
const {{ .EnumLower }}Names = "{{ .NameString }}"

// {{ .EnumLower }}NamesMap is a map of enum values to their canonical absolute
// name positions within the {{ .EnumLower }}Names string slice
var {{ .EnumLower }}NamesMap = map[{{ .WrapperName }}]string{    {{- range .EnumDefs }}
    {{ $.EnumType }}.{{ .EnumNameIdentifier }}: {{ $.EnumLower }}Names[{{ index $.NameOffsets .EnumNameIdentifier "start" }}:{{ index $.NameOffsets .EnumNameIdentifier "end" }}],
    {{- end }}
}

// String implements the Stringer interface.
// It returns the canonical absolute name of the enum value.
func ({{ .Receiver }} {{ .WrapperName }}) String() string {
    if str, ok := {{ .EnumLower }}NamesMap[{{ .Receiver }}]; ok {
        return str
    }
    return fmt.Sprintf("{{ .EnumLower }}(%d)", {{ .Receiver }}.{{ .EnumIota }})
}
`
	stringMethodTemplate = template.Must(template.New("stringMethod").Parse(stringMethodStr))
)

type stringMethodData struct {
	Receiver        string
	WrapperName     string
	EnumLower       string
	EnumIota        string
	EnumType        string
	NameString      string
	EnumDefs        []enumDefinition
	NameOffsets     map[string]map[string]int
	ContainerName   string
	CaseInsensitive bool
}

func (g *Writer) writeStringMethod(rep renderRequest) {
	edefs := rep.AllEnums
	var names bytes.Buffer
	type nameOffset struct {
		start, end int
	}
	nameOffsets := make(map[string]nameOffset)

	for _, e := range edefs {
		if len(e.Aliases) == 0 {
			e.Aliases = append(e.Aliases, e.EnumName)
		}
		name := e.Aliases[0]
		start := names.Len()
		names.WriteString(name)
		end := names.Len()
		nameOffsets[e.EnumNameIdentifier] = struct{ start, end int }{start, end}
	}
	nameOffsetsForTemplate := make(map[string]map[string]int)
	for id, offset := range nameOffsets {
		nameOffsetsForTemplate[id] = map[string]int{
			"start": offset.start,
			"end":   offset.end,
		}
	}
	d := stringMethodData{
		Receiver:        receiver(rep.EnumIota.Type),
		WrapperName:     wrapperName(rep.EnumIota.Type),
		EnumLower:       strings.ToLower(rep.EnumIota.Type),
		EnumIota:        rep.EnumIota.Type,
		EnumType:        enumType(rep.GenerationRequest),
		NameString:      names.String(),
		EnumDefs:        edefs,
		NameOffsets:     nameOffsetsForTemplate,
		CaseInsensitive: rep.Configuration.Insensitive,
	}
	g.writeTemplate(stringMethodTemplate, d)
}

var (
	isValidStr = `
// valid{{ .EnumType }} is a map of enum values to their validity
var valid{{ .EnumType }} = map[{{ .WrapperName }}]bool{
	{{- range .Enums }}
	{{ $.EnumType }}.{{ .EnumNameIdentifier }}: {{ .Valid }},
	{{- end }}
}

// IsValid checks whether the {{ .EnumType }} value is valid.
// A valid value is one that is defined in the original enum and not marked as invalid.
func ({{ .Receiver }} {{ .WrapperName }}) IsValid() bool {
	return valid{{ .EnumType }}[{{ .Receiver }}]
}
`
	isValidTemplate = template.Must(template.New("isValid").Parse(isValidStr))
)

type isValidFunctionData struct {
	Receiver    string
	EnumType    string
	WrapperName string
	Enums       []enumDefinition
}

func (g *Writer) writeIsValidFunction(rep renderRequest) {
	g.writeTemplate(isValidTemplate, isValidFunctionData{
		Receiver:    receiver(rep.EnumIota.Type),
		EnumType:    enumType(rep.GenerationRequest),
		WrapperName: wrapperName(rep.EnumIota.Type),
		Enums:       rep.AllEnums,
	})
}

func (g *Writer) writeNumberParsingMethods(rep renderRequest) {
	g.writeTemplate(parseIntegerGenericFunctionTemplate, parseNumberFunctionData{
		Constraints: rep.Configuration.Constraints,
		WrapperName: wrapperName(rep.EnumIota.Type),
		EnumType:    enumType(rep.GenerationRequest),
		Enums:       rep.AllEnums,
	})
}

func enumType(rep enum.GenerationRequest) string {
	return naming.ContainerName(rep.EnumIota.Type)
}

var (
	invalidEnumStr = `
	// invalid{{ .WrapperName }} is an invalid sentinel value for {{ .WrapperName }}
	var invalid{{ .WrapperName }} = {{ .WrapperName }}{
		{{ .RawEnumType }}: {{ .InvalidValue }},
	}
	`
	invalidEnumTemplate = template.Must(template.New("invalidEnum").Parse(invalidEnumStr))
)

func (g *Writer) writeInvalidEnumDefinition(enum enum.GenerationRequest) {
	g.writeTemplate(invalidEnumTemplate, newInterfaceFunctionData(enum))
}

type wrapperDefinition struct {
	WrapperName string
	WrapperType string
	EnumType    string
	Fields      []field

	EnumContainerName string
	Enums             []cenum
}

type field struct {
	Name string
	Type string
}

type cenum struct {
	Name     string
	EnumType string
}

var (
	wrapperDefinitionStr = `
// {{ .WrapperName }} is a type that represents a single enum value.
// It combines the core information about the enum constant and its defined fields.
type {{ .WrapperName }} struct {
  {{ .EnumType }}
  {{- range .Fields }}
  {{ .Name }} {{ .Type }}
  {{- end }}
}

// {{ .EnumContainerName }} is the container for all enum values.
// It is private; use the exported {{ .WrapperName }} values and methods instead.
type {{ .EnumContainerName }} struct {
  {{- range .Enums }}
  {{ .Name }} {{ .EnumType }}
  {{- end }}
}
`
	wrapperDefinitionTemplate = template.Must(
		template.New("wrapperDefinition").Parse(wrapperDefinitionStr))
)

func (g *Writer) writeWrapperDefinition(req renderRequest) {
	var (
		fields = make([]field, len(req.EnumIota.Fields)) // wrapper fields
		cenums = buildContainerEnums(req.AllEnums)       // container enums
		wName  = wrapperName(req.EnumIota.Type)          // wrapper name
		wType  = wrapperType(req.EnumIota.Type)          // wrapper type
	)
	for i, f := range req.EnumIota.Fields {
		fields[i] = field{
			Name: f.Name,
			Type: strings.AsType(f.Value),
		}
	}

	d := wrapperDefinition{
		WrapperName:       wName,
		WrapperType:       wType,
		Enums:             cenums,
		EnumType:          req.EnumIota.Type,
		Fields:            fields,
		EnumContainerName: containerType(req.GenerationRequest),
	}
	g.writeTemplate(wrapperDefinitionTemplate, d)
}

func wrapperName(enumType string) string {
	return naming.WrapperName(enumType)
}

func wrapperType(enumType string) string {
	return naming.WrapperType(enumType)
}

func containerType(rep enum.GenerationRequest) string {
	return naming.ContainerType(rep.EnumIota.Type)
}
