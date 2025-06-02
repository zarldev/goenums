package gofile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"text/template"
	"time"
	"unicode"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/strings"
)

var _ enum.Writer = &Writer{}

var (
	// ErrWriteGoFile is returned when an error occurs while writing the go file.
	ErrWriteGoFile = errors.New("error writing go file")
)

// Writer implements enum.Writer for go source files.
// It writes enum definitions to a file on provided filesystem,
// with the specified configuration.
type Writer struct {
	Configuration config.Configuration
	w             io.Writer
	fs            file.ReadCreateWriteFileFS
}

// WriterOption is a function that configures a Writer.
type WriterOption func(*Writer)

// WithFileSystem sets the filesystem to use for writing files.
func WithFileSystem(fs file.ReadCreateWriteFileFS) func(*Writer) {
	return func(w *Writer) {
		w.fs = fs
	}
}

// WithWriterConfiguration sets the configuration for the writer.
func WithWriterConfiguration(configuration config.Configuration) func(*Writer) {
	return func(w *Writer) {
		w.Configuration = configuration
	}
}

// NewWriter creates a new go file writer with the specified configuration and filesystem.
// The writer will write enum definitions to the provided filesystem.
// When no options are provided, it will write to stdout.
func NewWriter(opts ...WriterOption) *Writer {
	w := Writer{
		Configuration: config.Configuration{},
		fs:            &file.OSReadWriteFileFS{},
		w:             os.Stdout,
	}
	for _, opt := range opts {
		opt(&w)
	}
	return &w
}

func (g *Writer) Write(ctx context.Context,
	reqs []enum.GenerationRequest) error {
	for _, req := range reqs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !req.IsValid() {
			return fmt.Errorf("invalid enum: %s", req.SourceFilename)
		}
		dirPath := filepath.Dir(req.SourceFilename)
		if !filepath.IsLocal(dirPath) {
			return fmt.Errorf("invalid path: %s", dirPath)
		}
		outFilename := fmt.Sprintf("%s_enums.go", req.OutputFilename)
		if strings.Contains(outFilename, " ") || strings.Contains(outFilename, "/") {
			return fmt.Errorf("%w: '%s' contains invalid characters", ErrWriteGoFile, outFilename)
		}
		fullPath := filepath.Clean(filepath.Join(dirPath, outFilename))
		err := file.WriteToFileAndFormatFS(ctx, g.fs, fullPath, true,
			func(w io.Writer) error {
				g.w = w
				g.writeEnumGenerationRequest(req)
				return nil
			})
		if err != nil {
			return fmt.Errorf("%w: %s: %w", ErrWriteGoFile, fullPath, err)
		}
	}
	return nil
}

func (g *Writer) writeEnumGenerationRequest(req enum.GenerationRequest) {
	g.writeGeneratedComments(req)
	g.writePackageAndImports(req)
	g.writeWrapperDefinition(req)
	g.writeContainerDefinition(req)
	g.writeInvalidEnumDefinition(req)
	g.writeAllFunction(req)
	g.writeParseFunction(req)
	g.writeStringParsingMethod(req)
	g.writeNumberParsingMethods(req)
	g.writeExhaustiveFunction(req)
	g.writeIsValidFunction(req)
	if req.Configuration.Handlers.JSON {
		g.writeJSONMarshalMethod(req)
		g.writeJSONUnmarshalMethod(req)
	}
	if req.Configuration.Handlers.Text {
		g.writeTextMarshalMethod(req)
		g.writeTextUnmarshalMethod(req)
	}
	if req.Configuration.Handlers.SQL {
		g.writeScanMethod(req)
		g.writeValueMethod(req)
	}
	if req.Configuration.Handlers.Binary {
		g.writeBinaryMarshalMethod(req)
		g.writeBinaryUnmarshalMethod(req)
	}
	if req.Configuration.Handlers.YAML {
		g.writeYAMLMarshalMethod(req)
		g.writeYAMLUnmarshalMethod(req)
	}
	if req.Configuration.Constraints {
		g.writeConstraints(req)
	}
	g.writeStringMethod(req)
	g.writeCompileCheck(req)
}

var (
	constraintsStr = `
	 
	type float interface {
		float32 | float64
	}
	type integer interface {
		int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr
	}
	type number interface {
		integer | float
	}
`
	constraintsTemplate = template.Must(template.New("constraints").Parse(constraintsStr))
)

func (g *Writer) writeConstraints(rep enum.GenerationRequest) {
	g.writeTemplate(constraintsTemplate, rep)
}

var (
	jsonMarshalStr = `
// MarshalJSON implements the json.Marshaler interface for {{ .WrapperName }}.
// It returns the JSON representation of the enum value as a byte slice.
func ({{ .Receiver }} {{ .WrapperName }}) MarshalJSON() ([]byte, error) {
	return []byte( "\"" + {{ .Receiver }}.String() + "\""), nil 
}
	`
	jsonMarshalTemplate = template.Must(template.New("jsonMarshal").Parse(jsonMarshalStr))

	jsonUnmarshalStr = `
// UnmarshalJSON implements the json.Unmarshaler interface for {{ .WrapperName }}.
// It parses the JSON representation of the enum value from the byte slice.
// It returns an error if the input is not a valid JSON representation.
func ({{ .Receiver }} *{{ .WrapperName }}) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, "\""), "\"")
	new{{ .Receiver }}, err := Parse{{ .WrapperName }}(b)
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
	return []byte("\"" + {{ .Receiver }}.String() + "\""), nil
}
`
)

type interfaceFunctionData struct {
	Receiver    string
	WrapperName string
	EnumName    string
	EnumType    string
}

func newInterfaceFunctionData(rep enum.GenerationRequest) interfaceFunctionData {
	return interfaceFunctionData{
		Receiver:    receiver(rep.EnumIota.Type),
		WrapperName: wrapperName(rep.EnumIota.Type),
		EnumName:    strings.ToUpper(rep.EnumIota.Type),
		EnumType:    enumType(rep),
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
func ({{ .Receiver }} *{{ .WrapperName }}) UnmarshalText(b []byte) error {
	new{{ .Receiver }}, err := Parse{{ .WrapperName }}(b)
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
func ({{ .Receiver }} *{{ .WrapperName }}) UnmarshalBinary(b []byte) error {
	new{{ .Receiver }}, err := Parse{{ .WrapperName }}(b)
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
// It returns an error if the enum value is invalid.				
func ({{ .Receiver }} {{ .WrapperName }}) MarshalYAML() ([]byte, error) {
	return []byte({{ .Receiver }}.String()), nil
}
`
	yamlMarshalTemplate = template.Must(template.New("yamlMarshal").Parse(yamlMarshalStr))

	yamlUnmarshalStr = `
// UnmarshalYAML implements the yaml.Unmarshaler interface for {{ .WrapperName }}.
// It parses the string representation of the enum value from the YAML node.
// It returns an error if the YAML node does not contain a valid enum value.
func ({{ .Receiver }} *{{ .WrapperName }}) UnmarshalYAML(b []byte) error {
	new{{ .Receiver }}, err := Parse{{ .WrapperName }}(b)
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
	var x [{{len .Enums}}]struct{}
	{{- range .Enums }}
	_ = x[{{ .Name }}-{{ .Index }}]
	{{- end }}
}
	`
	compileCheckTemplate = template.Must(template.New("compileCheck").Parse(compileCheckStr))
)

type compileCheckData struct {
	Enums []enum.Enum
}

func (g *Writer) writeCompileCheck(rep enum.GenerationRequest) {
	g.writeTemplate(compileCheckTemplate, compileCheckData{
		Enums: rep.EnumIota.Enums,
	})
}

var (
	stringMethodStr = `
// {{ .EnumLower }}Names is a constant string slice containing all enum values cononical absolute names
const {{ .EnumLower }}Names = "{{ .NameString }}"

// {{ .EnumLower }}NamesMap is a map of enum values to their canonical absolute 
// name positions within the {{ .EnumLower }}Names string slice
var {{ .EnumLower }}NamesMap = map[{{ .WrapperName }}]string{
    {{- range .EnumDefs }}
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

func (g *Writer) writeStringMethod(rep enum.GenerationRequest) {
	edefs := enumDefinitions(rep)
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
		EnumType:        enumType(rep),
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

func (g *Writer) writeIsValidFunction(rep enum.GenerationRequest) {
	g.writeTemplate(isValidTemplate, isValidFunctionData{
		Receiver:    receiver(rep.EnumIota.Type),
		EnumType:    enumType(rep),
		WrapperName: wrapperName(rep.EnumIota.Type),
		Enums:       enumDefinitions(rep),
	})
}

func (g *Writer) writeNumberParsingMethods(rep enum.GenerationRequest) {
	g.writeTemplate(parseIntegerGenericFunctionTemplate, parseNumberFunctionData{
		Constraints:   rep.Configuration.Constraints,
		HasStartIndex: rep.EnumIota.StartIndex > 0,
		StartIndex:    rep.EnumIota.StartIndex,
		WrapperName:   wrapperName(rep.EnumIota.Type),
		EnumType:      enumType(rep),
	})
}

func enumType(rep enum.GenerationRequest) string {
	return strings.Pluralise(strings.Camel(rep.EnumIota.Type))
}

var (
	invalidEnumStr = `
	// invalid{{ .WrapperName }} is an invalid sentinel value for {{ .WrapperName }}
	var invalid{{ .WrapperName }} = {{ .WrapperName }}{}
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
// It combines the core information about the enum constant and it's defined fields.
type {{ .WrapperName }} struct {
  {{ .EnumType }}
  {{- range .Fields }}
  {{ .Name }} {{ .Type }}
  {{- end }}
}

// {{ .EnumContainerName }} is the container for all enum values.
// It is private and should not be used directly use the public methods on the {{.WrapperName}} type.
type {{ .EnumContainerName }} struct {
  {{- range .Enums }}
  {{ .Name }} {{ .EnumType }}
  {{- end }}
}
`
	wrapperDefinitionTemplate = template.Must(
		template.New("wrapperDefinition").Parse(wrapperDefinitionStr))
)

func (g *Writer) writeWrapperDefinition(enum enum.GenerationRequest) {
	var (
		fields = make([]field, len(enum.EnumIota.Fields)) // wrapper fields
		cenums = make([]cenum, len(enum.EnumIota.Enums))  // container enums
		wName  = wrapperName(enum.EnumIota.Type)          // wrapper name
		wType  = wrapperType(enum.EnumIota.Type)          // wrapper type
	)
	for i, f := range enum.EnumIota.Fields {
		fields[i] = field{
			Name: f.Name,
			Type: strings.AsType(f.Value),
		}
	}
	for i, e := range enum.EnumIota.Enums {
		cenums[i] = cenum{
			Name:     strings.ToUpper(e.Name),
			EnumType: wName,
		}
	}

	d := wrapperDefinition{
		WrapperName:       wName,
		WrapperType:       wType,
		Enums:             cenums,
		EnumType:          enum.EnumIota.Type,
		Fields:            fields,
		EnumContainerName: containerType(enum),
	}
	g.writeTemplate(wrapperDefinitionTemplate, d)
}

func wrapperName(enum string) string {
	if strings.IsPlural(enum) {
		enum = strings.Singularise(enum)
		strings.Camel(enum)
	}
	return strings.Camel(enum)
}

func wrapperType(enum string) string {
	return strings.Camel(enum)
}

func containerType(enum enum.GenerationRequest) string {
	cName := strings.Lower1stCharacter(enum.EnumIota.Type)
	cName = strings.Pluralise(cName)
	return cName + "Container"
}

type generatedComment struct {
	Version        string
	Time           string
	Command        string
	SourceFilename string
}

var (
	generatedCommentStr = `
// DO NOT EDIT.	
// code generated by goenums {{.Version}} at {{.Time}}. 
// 
// github.com/zarldev/goenums
//
// using the command:
// {{ .Command }}
	`
	generatedCommentTemplate = template.Must(
		template.New("generatedComment").Parse(generatedCommentStr))
)

func (g *Writer) writeTemplate(t *template.Template, d any) {
	err := t.Execute(g.w, d)
	if err != nil {
		slog.Default().Error("error writing template", "template", t.Name(), "error", err)
	}
}

func (g *Writer) writeGeneratedComments(rep enum.GenerationRequest) {
	g.writeTemplate(generatedCommentTemplate, generatedComment{
		Version:        rep.Version,
		Time:           time.Now().Format(time.Stamp),
		Command:        rep.Command(),
		SourceFilename: rep.SourceFilename,
	})
}

type packageImport struct {
	PackageName     string
	Imports         []string
	ExternalImports []string
}

var (
	packageImportStr = `
package {{ .PackageName }}

import (
{{- range .Imports }}
	"{{ . }}"
{{- end }}
{{ if .ExternalImports }}
{{ range .ExternalImports }}
	"{{ . }}"
{{ end }}
{{ end }}
	)
`
	packageImportTemplate = template.Must(template.New("packageImport").Parse(packageImportStr))
)

func (g *Writer) writePackageAndImports(rep enum.GenerationRequest) {
	externalImports := []string{}
	imports := []string{"fmt", "bytes", "database/sql/driver", "math"}
	imports = append(imports, rep.Imports...)
	if !rep.Configuration.Legacy {
		imports = append(imports, "iter")
	}
	if !rep.Configuration.Constraints {
		if slices.Contains(imports, "golang.org/x/exp/constraints") {
			imports = slices.DeleteFunc(imports, func(s string) bool {
				return s == "golang.org/x/exp/constraints"
			})
			slices.Sort(imports)
		}
		externalImports = append(externalImports, "golang.org/x/exp/constraints")
	}
	slices.Sort(imports)
	g.writeTemplate(packageImportTemplate, packageImport{
		PackageName:     rep.Package,
		Imports:         imports,
		ExternalImports: externalImports,
	})
}

type containerDefinition struct {
	WrapperName   string
	ContainerName string
	ContainerType string
	EnumDefs      []enumDefinition
}

var (
	containerDefinitionStr = `
// {{.ContainerName}} is a main entry point using the {{.WrapperName}} type.
// It it a container for all enum values and provides a convenient way to access all enum values and perform
// operations, with convenience methods for common use cases.
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

func (g *Writer) writeContainerDefinition(rep enum.GenerationRequest) {
	edefs := enumDefinitions(rep)
	cdef := containerDefinition{
		WrapperName:   wrapperName(rep.EnumIota.Type),
		ContainerType: containerType(rep),
		ContainerName: strings.Pluralise(strings.Camel(rep.EnumIota.Type)),
		EnumDefs:      edefs,
	}
	g.writeTemplate(containerDefinitionTemplate, cdef)
}

func enumDefinitions(rep enum.GenerationRequest) []enumDefinition {
	edefs := make([]enumDefinition, 0)
	for _, e := range rep.EnumIota.Enums {
		if len(rep.EnumIota.Fields) > 0 &&
			len(e.Fields) == 0 {
			continue
		}
		fields := e.Fields
		ffields := make([]enum.Field, len(fields))
		for j, f := range fields {
			ffields[j] = enum.Field{
				Name:  f.Name,
				Value: strings.Ify(f.Value),
			}
		}
		aliases := e.Aliases
		if rep.Configuration.Insensitive {
			for _, a := range e.Aliases {
				lwr := strings.ToLower(a)
				if lwr == a {
					continue
				}
				if slices.Contains(aliases, lwr) {
					continue
				}
				aliases = append(aliases, strings.ToLower(a))
			}
		}
		edefs = append(edefs, enumDefinition{
			EnumName:           e.Name,
			EnumNameIdentifier: strings.ToUpper(e.Name),
			EnumType:           wrapperName(rep.EnumIota.Type),
			Fields:             ffields,
			IotaType:           rep.EnumIota.Type,
			Aliases:            aliases,
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

func (g *Writer) writeAllFunction(rep enum.GenerationRequest) {
	allData := allFunctionData{
		Receiver:      receiver(rep.EnumIota.Type),
		ContainerType: containerType(rep),
		ContainerName: strings.Pluralise(strings.Camel(rep.EnumIota.Type)),
		WrapperName:   wrapperName(rep.EnumIota.Type),
		EnumDefs:      enumDefinitions(rep),
		Legacy:        rep.Configuration.Legacy,
	}
	g.writeTemplate(allFunctionTemplate, allData)
}

type parseFunctionData struct {
	WrapperName string
	FailFast    bool
	Enums       []enum.Enum
}

var (
	parseFunctionStr = `
// Parse{{.WrapperName}} parses the input value into an enum value.
// It returns the parsed enum value or an error if the input is invalid.
// It is a convenience function that can be used to parse enum values from
// various input types, such as strings, byte slices, or other enum types.
func Parse{{.WrapperName}}(input any) ({{.WrapperName}}, error) {
	var res = invalid{{.WrapperName}}
	switch v := input.(type) {
	case {{.WrapperName}}:
		return v, nil
	case string:
		res = stringTo{{.WrapperName}}(v)
	case fmt.Stringer:
		res = stringTo{{.WrapperName}}(v.String())
	case []byte:
		res = stringTo{{.WrapperName}}(string(v))
	case int:
		res = numberTo{{.WrapperName}}(v)
	case int8:
		res = numberTo{{.WrapperName}}(v)
	case int16:
		res = numberTo{{.WrapperName}}(v)
	case int32:
		res = numberTo{{.WrapperName}}(v)
	case int64:
		res = numberTo{{.WrapperName}}(v)
	case uint:
		res = numberTo{{.WrapperName}}(v)
	case uint8:
		res = numberTo{{.WrapperName}}(v)
	case uint16:
		res = numberTo{{.WrapperName}}(v)
	case uint32:
		res = numberTo{{.WrapperName}}(v)
	case uint64:
		res = numberTo{{.WrapperName}}(v)
	case float32:
		res = numberTo{{.WrapperName}}(v)
	case float64:
		res = numberTo{{.WrapperName}}(v)
	default:
		return res, fmt.Errorf("invalid type %T", input)
	}
	{{- if .FailFast}}
	if res == invalid{{.WrapperName}} {
	  return res, fmt.Errorf("invalid value %v", input)
	}
	{{- end}}
	return res, nil
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
	Fields             []enum.Field
	Aliases            []string
	Valid              bool
}

var (
	parseStringFunctionStr = `
// {{ .EnumNameMap }} is a map of enum values to their {{.WrapperName}} representation
// It is used to convert string representations of enum values into their {{.WrapperName}} representation.
var {{.EnumNameMap}} = map[string]{{.WrapperName}}{
{{- range .Enums }}
  {{- $enum := . }}
  {{- range .Aliases }}
    "{{ . }}": {{ $.EnumType }}.{{ $enum.EnumNameIdentifier }},
  {{- end }}
{{- end }}
}

// stringTo{{.WrapperName}} converts a string representation of an enum value into its {{.WrapperName}} representation
// It returns the {{.WrapperName}} representation of the enum value if the string is valid
// Otherwise, it returns invalid{{.WrapperName}}
func stringTo{{.WrapperName}}(s string) {{.WrapperName}} {
  if t, ok := {{.EnumNameMap}}[s]; ok {
    return t
  }
  return invalid{{.WrapperName}}
}
`
	parseStringFunctionTemplate = template.Must(template.New("parseStringFunction").Parse(parseStringFunctionStr))
)

func (g *Writer) writeStringParsingMethod(rep enum.GenerationRequest) {
	g.writeTemplate(parseStringFunctionTemplate, parseStringFunctionData{
		WrapperName:     wrapperName(rep.EnumIota.Type),
		EnumNameMap:     enumNameMap(rep.EnumIota.Type),
		EnumType:        enumType(rep),
		Enums:           enumDefinitions(rep),
		CaseInsensitive: rep.Configuration.Insensitive,
	})
}

type parseNumberFunctionData struct {
	Constraints   bool
	WrapperName   string
	EnumType      string
	StartIndex    int
	HasStartIndex bool
}

var (
	parseIntegerGenericFunctionTemplate = template.Must(template.New("parseIntegerGenericFunction").Parse(`

// numberTo{{.WrapperName}} converts a numeric value to a {{.WrapperName}}
// It returns the {{.WrapperName}} representation of the enum value if the numeric value is valid
// Otherwise, it returns invalid{{.WrapperName}}
{{- if .Constraints }}
	func numberTo{{.WrapperName}}[T number](num T) {{.WrapperName}} {
{{- else }}
func numberTo{{.WrapperName}}[T constraints.Integer | constraints.Float](num T) {{.WrapperName}} {
{{- end }}
	f := float64(num)
    if math.Floor(f) != f {
        return invalid{{.WrapperName}}
    }
	i := int(f)
	if i <= 0 || i > len({{.EnumType}}.allSlice()) {
		return invalid{{.WrapperName}}
	}
	{{- if .StartIndex }}
	return {{.EnumType}}.allSlice()[i-{{.StartIndex}}]
	{{- else }}
	return {{.EnumType}}.allSlice()[i]
	{{- end }}
}

`))
)

func enumNameMap(enumType string) string {
	return strings.Pluralise(enumType) + "NameMap"
}

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

func (g *Writer) writeExhaustiveFunction(rep enum.GenerationRequest) {
	edefs := enumDefinitions(rep)
	exhaustiveData := exhaustiveFunctionData{
		WrapperName: wrapperName(rep.EnumIota.Type),
		EnumType:    enumType(rep),
		Enums:       edefs,
	}
	g.writeTemplate(exhaustiveTemplate, exhaustiveData)
}
