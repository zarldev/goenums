package gofile

import (
	"slices"
	"text/template"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/strings"
)

type containerDefinition struct {
	WrapperName   string
	ContainerName string
	ContainerType string
	EnumDefs      []enumDefinition
}

var (
	containerDefinitionStr = `
// {{.ContainerName}} is a main entry point using the {{.WrapperName}} type.
// It is a container for all enum values and provides a convenient way to access all enum values and perform// operations, with convenience methods for common use cases.
var {{.ContainerName}} = {{.ContainerType}}{
{{- range .EnumDefs }}
	{{.EnumNameIdentifier}}: {{.EnumType}} {
		{{.IotaType}}: {{.EnumName}},
		{{- range .Fields }}
		{{.Name}}: {{.Value}},
		{{- end }}
	},
{{- end }}
}
`
	containerDefinitionTemplate = template.Must(template.New("containerDefinition").Parse(containerDefinitionStr))
)

func (g *Writer) writeContainerDefinition(rep renderRequest) {
	edefs := rep.AllEnums
	cdef := containerDefinition{
		WrapperName:   wrapperName(rep.EnumIota.Type),
		ContainerType: containerType(rep.GenerationRequest),
		ContainerName: strings.Pluralise(strings.Camel(rep.EnumIota.Type)),
		EnumDefs:      edefs,
	}
	g.writeTemplate(containerDefinitionTemplate, cdef)
}

func buildEnumDefinitions(rep enum.GenerationRequest) []enumDefinition {
	edefs := make([]enumDefinition, 0, len(rep.EnumIota.Enums))
	for _, e := range rep.EnumIota.Enums {
		fields := e.Fields
		ffields := make([]enum.Field, len(fields))
		for j, f := range fields {
			ffields[j] = enum.Field{
				Name:  f.Name,
				Value: strings.Ify(f.Value),
			}
		}
		aliases := enumAliases(e)
		edefs = append(edefs, enumDefinition{
			EnumName:           e.Name,
			Value:              e.Value,
			EnumNameIdentifier: strings.ToUpper(e.Name),
			EnumType:           wrapperName(rep.EnumIota.Type),
			Fields:             ffields,
			IotaType:           rep.EnumIota.Type,
			Aliases:            aliases,
			QuotedAliases:      quotedAliases(aliases),
			Valid:              e.Valid,
		})
	}
	return edefs
}

type allFunctionData struct {
	Legacy        bool
	Receiver      string
	ContainerType string
	ContainerName string
	WrapperName   string
	EnumDefs      []enumDefinition
}

var (
	allFunctionStr = `
// allSlice returns a slice of all enum values.
// This method is useful for iterating over all enum values in a loop.
func ({{.Receiver}} {{.ContainerType}}) allSlice() []{{.WrapperName}} {
    return []{{.WrapperName}}{
        {{-  range .EnumDefs}}
        {{$.ContainerName}}.{{.EnumNameIdentifier}},
        {{- end}}
    }
}
{{- if .Legacy}}
// All returns a slice of all enum values.
// This method is useful for iterating over all enum values in a loop.
func ({{.Receiver}} {{.ContainerType}}) All() []{{.WrapperName}} {
    return {{.Receiver}}.allSlice()
}
{{- else}}
// All returns an iterator over all enum values.
// This method is useful for iterating over all enum values in a loop.
func ({{.Receiver}} {{.ContainerType}}) All() iter.Seq[{{.WrapperName}}] {
    return func(yield func({{.WrapperName}}) bool) {
        for _, v := range {{.Receiver}}.allSlice() {
            if !yield(v) {
                return
            }
        }
    }
}
{{- end}}
    `
	allFunctionTemplate = template.Must(template.New("allFunction").Parse(allFunctionStr))
)

func (g *Writer) writeAllFunction(rep renderRequest) {
	allData := allFunctionData{
		Receiver:      receiver(rep.EnumIota.Type),
		ContainerType: containerType(rep.GenerationRequest),
		ContainerName: strings.Pluralise(strings.Camel(rep.EnumIota.Type)),
		WrapperName:   wrapperName(rep.EnumIota.Type),
		EnumDefs:      rep.ValidEnums,
		Legacy:        rep.Configuration.Legacy,
	}
	g.writeTemplate(allFunctionTemplate, allData)
}

type matchFunctionData struct {
	EnumType    string
	MatcherName string
	WrapperName string
	Enums       []matchHandlerDefinition
}

type matchHandlerDefinition struct {
	EnumNameIdentifier string
	MethodName         string
}

var (
	matchFunctionStr = `
type {{ .MatcherName }} interface {
    {{- range .Enums }}
    {{ .MethodName }}()
    {{- end }}
}

// Match{{.WrapperName}} dispatches to the matcher method for the given enum value.
// The {{ .MatcherName }} interface provides compile-time exhaustiveness: when enum
// values are added, removed, or renamed, existing matcher implementations stop
// satisfying this interface until they are updated.
func Match{{.WrapperName}}(en {{.WrapperName}}, matcher {{ .MatcherName }}) error {
    if matcher == nil {
        return fmt.Errorf("nil {{ .MatcherName }}")
    }
    switch en {
    {{- range .Enums }}
    case {{ $.EnumType }}.{{ .EnumNameIdentifier }}:
        matcher.{{ .MethodName }}()
    {{- end }}
    default:
        return fmt.Errorf("unhandled {{ .WrapperName }}: %v", en)
    }
    return nil
}
`
	mustMatchFunctionStr = `
// MustMatch{{.WrapperName}} dispatches to the matcher method for the given enum value.
// It panics if matcher is nil or the enum value is not handled.
func MustMatch{{.WrapperName}}(en {{.WrapperName}}, matcher {{ .MatcherName }}) {
    if err := Match{{.WrapperName}}(en, matcher); err != nil {
        panic(err)
    }
}
`
	matchFunctionTemplate     = template.Must(template.New("matchFunction").Parse(matchFunctionStr))
	mustMatchFunctionTemplate = template.Must(template.New("mustMatchFunction").Parse(mustMatchFunctionStr))
)

func (g *Writer) writeMatchFunction(rep renderRequest) {
	enums := rep.ValidEnums
	handlers := make([]matchHandlerDefinition, len(enums))
	for i, e := range enums {
		handlers[i] = matchHandlerDefinition{
			EnumNameIdentifier: e.EnumNameIdentifier,
			MethodName:         strings.Camel(e.EnumName),
		}
	}
	d := matchFunctionData{
		EnumType:    enumType(rep.GenerationRequest),
		MatcherName: matcherName(rep.EnumIota.Type),
		WrapperName: wrapperName(rep.EnumIota.Type),
		Enums:       handlers,
	}
	g.writeTemplate(matchFunctionTemplate, d)
	g.writeTemplate(mustMatchFunctionTemplate, d)
}

func matcherName(enumType string) string {
	return wrapperName(enumType) + "Matcher"
}

func enumAliases(e enum.Enum) []string {
	if len(e.Aliases) == 0 {
		return []string{e.Name}
	}
	aliases := make([]string, 0, len(e.Aliases))
	for _, alias := range e.Aliases {
		if slices.Contains(aliases, alias) {
			continue
		}
		aliases = append(aliases, alias)
	}
	return aliases
}
