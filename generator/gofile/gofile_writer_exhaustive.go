package gofile

import (
	"text/template"
)

var (
	exhaustiveStr = `
// Exhaustive{{ .EnumType }} iterates over all enum values and calls the provided function for each value.
// This function is useful for performing operations on all valid enum values in a loop.
func Exhaustive{{ .EnumType }}(f func({{ .WrapperName }})) {
    for _, p := range {{ .EnumType }}.allSlice() {
        f(p)
    }
}
`
	exhaustiveTemplate = template.Must(template.New("exhaustive").Parse(exhaustiveStr))
)

type exhaustiveFunctionData struct {
	EnumType    string
	WrapperName string
	Enums       []enumDefinition
}

func (g *Writer) writeExhaustiveFunction(rep renderRequest) {
	edefs := rep.ValidEnums
	exhaustiveData := exhaustiveFunctionData{
		WrapperName: wrapperName(rep.EnumIota.Type),
		EnumType:    enumType(rep.GenerationRequest),
		Enums:       edefs,
	}
	g.writeTemplate(exhaustiveTemplate, exhaustiveData)
}
