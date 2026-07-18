package gofile

import (
	"text/template"
	"unicode"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/strings"
)

var (
	jsonMarshalStr = `
// MarshalJSON implements the json.Marshaler interface for {{ .WrapperName }}.
// It returns the JSON representation of the enum value as a byte slice.
func ({{ .Receiver }} {{ .WrapperName }}) MarshalJSON() ([]byte, error) {
	return json.Marshal({{ .Receiver }}.String())
}
	`
	jsonMarshalTemplate = template.Must(template.New("jsonMarshal").Parse(jsonMarshalStr))

	jsonUnmarshalStr = `
// UnmarshalJSON implements the json.Unmarshaler interface for {{ .WrapperName }}.
// It parses the JSON representation of the enum value from the byte slice.
// It returns an error if the input is not a valid JSON representation.
func ({{ .Receiver }} *{{ .WrapperName }}) UnmarshalJSON(by []byte) error {
	var value string
	if err := json.Unmarshal(by, &value); err == nil {
		new{{ .Receiver }}, err := Parse{{ .WrapperName }}(value)
		if err != nil {
			return err
		}
		*{{ .Receiver }} = new{{ .Receiver }}
		return nil
	}
	new{{ .Receiver }}, err := Parse{{ .WrapperName }}(by)
	if err != nil {
		return err
	}
	*{{ .Receiver }} = new{{ .Receiver }}
	return nil
}
`
	jsonUnmarshalTemplate = template.Must(template.New("jsonUnmarshal").Parse(jsonUnmarshalStr))
	textMarshalStr        = `
// MarshalText implements the encoding.TextMarshaler interface for {{ .WrapperName }}.
// It returns the string representation of the enum value as a byte slice
func ({{ .Receiver }} {{ .WrapperName }}) MarshalText() ([]byte, error) {
	{{- if .LegacyTextMarshal }}
	return []byte("\"" + {{ .Receiver }}.String() + "\""), nil
	{{- else }}
	return []byte({{ .Receiver }}.String()), nil
	{{- end }}
}
`
)

type interfaceFunctionData struct {
	Receiver          string
	WrapperName       string
	EnumName          string
	EnumType          string
	RawEnumType       string
	InvalidValue      int
	LegacyTextMarshal bool
}

func newInterfaceFunctionData(rep enum.GenerationRequest) interfaceFunctionData {
	invalidValue := -1
	if rep.EnumIota.StartIndex < 0 {
		invalidValue = rep.EnumIota.StartIndex + len(rep.EnumIota.Enums)
	}

	return interfaceFunctionData{
		Receiver:          receiver(rep.EnumIota.Type),
		WrapperName:       wrapperName(rep.EnumIota.Type),
		EnumName:          strings.ToUpper(rep.EnumIota.Type),
		EnumType:          enumType(rep),
		RawEnumType:       rep.EnumIota.Type,
		InvalidValue:      invalidValue,
		LegacyTextMarshal: rep.Configuration.LegacyTextMarshal,
	}
}

func receiver(enumType string) string {
	if strings.Contains(enumType, ".") {
		return strings.Split(enumType, ".")[0]
	}
	if len(enumType) == 0 {
		return "r"
	}
	firstChar := enumType[0]
	return string(unicode.ToLower(rune(firstChar)))
}

func (g *Writer) writeJSONMarshalMethod(rep enum.GenerationRequest) {
	g.writeTemplate(jsonMarshalTemplate, newInterfaceFunctionData(rep))
}

func (g *Writer) writeJSONUnmarshalMethod(rep enum.GenerationRequest) {
	g.writeTemplate(jsonUnmarshalTemplate, newInterfaceFunctionData(rep))
}

var (
	textMarshalTemplate = template.Must(template.New("textMarshal").Parse(textMarshalStr))

	textUnmarshalStr = `
// UnmarshalText implements the encoding.TextUnmarshaler interface for {{ .WrapperName }}.
// It parses the string representation of the enum value from the byte slice.
// It returns an error if the byte slice does not contain a valid enum value.
func ({{ .Receiver }} *{{ .WrapperName }}) UnmarshalText(by []byte) error {
	new{{ .Receiver }}, err := Parse{{ .WrapperName }}(by)
	if err != nil {
		return err
	}
	*{{ .Receiver }} = new{{ .Receiver }}
	return nil
}
`
	textUnmarshalTemplate = template.Must(template.New("textUnmarshal").Parse(textUnmarshalStr))
)

func (g *Writer) writeTextMarshalMethod(rep enum.GenerationRequest) {
	g.writeTemplate(textMarshalTemplate, newInterfaceFunctionData(rep))
}

func (g *Writer) writeTextUnmarshalMethod(rep enum.GenerationRequest) {
	g.writeTemplate(textUnmarshalTemplate, newInterfaceFunctionData(rep))
}

var (
	binaryMarshalStr = `
// MarshalBinary implements the encoding.BinaryMarshaler interface for {{ .WrapperName }}.
// It returns the binary representation of the enum value as a byte slice.
func ({{ .Receiver }} {{ .WrapperName }}) MarshalBinary() ([]byte, error) {
	return []byte("\"" + {{ .Receiver }}.String() + "\""), nil
}
`
	binaryMarshalTemplate = template.Must(template.New("binaryMarshal").Parse(binaryMarshalStr))

	binaryUnmarshalStr = `
// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for {{ .WrapperName }}.
// It parses the binary representation of the enum value from the byte slice.
// It returns an error if the byte slice does not contain a valid enum value.
func ({{ .Receiver }} *{{ .WrapperName }}) UnmarshalBinary(by []byte) error {
	new{{ .Receiver }}, err := Parse{{ .WrapperName }}(by)
	if err != nil {
		return err
	}
	*{{ .Receiver }} = new{{ .Receiver }}
	return nil
}
`
	binaryUnmarshalTemplate = template.Must(template.New("binaryUnmarshal").Parse(binaryUnmarshalStr))
)

func (g *Writer) writeBinaryMarshalMethod(rep enum.GenerationRequest) {
	g.writeTemplate(binaryMarshalTemplate, newInterfaceFunctionData(rep))
}

func (g *Writer) writeBinaryUnmarshalMethod(rep enum.GenerationRequest) {
	g.writeTemplate(binaryUnmarshalTemplate, newInterfaceFunctionData(rep))
}

var (
	yamlMarshalStr = `
// MarshalYAML implements the yaml.Marshaler interface for {{ .WrapperName }}.
// It returns the string representation of the enum value.
func ({{ .Receiver }} {{ .WrapperName }}) MarshalYAML() ([]byte, error) {
	return []byte({{ .Receiver }}.String()), nil}
`
	yamlMarshalTemplate = template.Must(template.New("yamlMarshal").Parse(yamlMarshalStr))

	yamlUnmarshalStr = `
// UnmarshalYAML implements the yaml.Unmarshaler interface for Planet.
// It parses the byte slice representation of the enum value and returns an error
// if the YAML byte slice does not contain a valid enum value.
func ({{ .Receiver }} *{{ .WrapperName }}) UnmarshalYAML(by []byte) error {	new{{ .Receiver }}, err := Parse{{ .WrapperName }}(by)
	if err != nil {
		return err
	}
	*{{ .Receiver }} = new{{ .Receiver }}
	return nil
}
`
	yamlUnmarshalTemplate = template.Must(template.New("yamlUnmarshal").Parse(yamlUnmarshalStr))
)

func (g *Writer) writeYAMLMarshalMethod(rep enum.GenerationRequest) {
	g.writeTemplate(yamlMarshalTemplate, newInterfaceFunctionData(rep))
}

func (g *Writer) writeYAMLUnmarshalMethod(rep enum.GenerationRequest) {
	g.writeTemplate(yamlUnmarshalTemplate, newInterfaceFunctionData(rep))
}

var (
	scanStr = `
// Scan implements the database/sql.Scanner interface for {{ .WrapperName }}.
// It parses the string representation of the enum value from the database row.
// It returns an error if the row does not contain a valid enum value.
func ({{ .Receiver }} *{{ .WrapperName }}) Scan(value any) error {
	new{{ .Receiver }}, err := Parse{{ .WrapperName }}(value)
	if err != nil {
		return err
	}
	*{{ .Receiver }} = new{{ .Receiver }}
	return nil
}
`
	scanTemplate = template.Must(template.New("scan").Parse(scanStr))

	valueStr = `
// Value implements the database/sql/driver.Valuer interface for {{ .WrapperName }}.
// It returns the string representation of the enum value.
func ({{ .Receiver }} {{ .WrapperName }}) Value() (driver.Value, error) {
	return {{ .Receiver }}.String(), nil
}
`
	valueTemplate = template.Must(template.New("value").Parse(valueStr))
)

func (g *Writer) writeScanMethod(rep enum.GenerationRequest) {
	g.writeTemplate(scanTemplate, newInterfaceFunctionData(rep))
}

func (g *Writer) writeValueMethod(rep enum.GenerationRequest) {
	g.writeTemplate(valueTemplate, newInterfaceFunctionData(rep))
}

var (
	compileCheckStr = `
// Compile-time check that all enum values are valid.
// This function is used to ensure that all enum values are defined and valid.
// It is called by the compiler to verify that the enum values are valid.
func _() {
    // An "invalid array index" compiler error signifies that the constant values have changed.
    // Re-run the goenums command to generate them again.
    // Does not identify newly added constant values unless order changes
    var x [{{ .ArraySize }}]struct{}
    {{- range .Enums }}
    _ = x[{{ .IndexExpr }}]
    {{- end }}
}
    `
	compileCheckTemplate = template.Must(template.New("compileCheck").Parse(compileCheckStr))
)

type compileCheckEnum struct {
	Name      string
	Value     int
	IndexExpr string
}

type compileCheckData struct {
	ArraySize int
	Enums     []compileCheckEnum
}
