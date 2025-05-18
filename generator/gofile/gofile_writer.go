package gofile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"text/template"

	"time"

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
	enums []enum.GenerationRequest) error {
	for _, enum := range enums {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		dirPath := filepath.Dir(enum.SourceFilename)
		outFilename := fmt.Sprintf("%s_enums.go", enum.OutputFilename)
		if strings.Contains(outFilename, " ") || strings.Contains(outFilename, "/") {
			return fmt.Errorf("%w: '%s' contains invalid characters", ErrWriteGoFile, outFilename)
		}
		fullPath := filepath.Join(dirPath, outFilename)
		err := file.WriteToFileAndFormatFS(ctx, g.fs, fullPath, true,
			func(w io.Writer) error {
				g.w = w
				g.writeEnumGenerationRequest(enum)
				return nil
			})
		if err != nil {
			return fmt.Errorf("%w: %s: %w", ErrWriteGoFile, fullPath, err)
		}
	}
	return nil
}

// func (g *Writer) writeEnumFile(enum enum.Representation) {
// 	// Top-level file structure
// 	g.writeGeneratedComment(enum)
// 	g.writePackage(enum)
// 	g.writeImports(enum)

// 	// Type definitions
// 	g.writeWrapperType(enum)
// 	g.writeInvalidTypeDefinition(enum)

// 	// Methods and functions
// 	g.writeAllMethod(enum)
// 	g.writeParsingMethods(enum)
// 	g.writeExhaustiveMethod(enum)
// 	g.writeIsValidMethod(enum)

// 	// Database and JSON handling
// 	g.writeJSONMarshalMethod(enum)
// 	g.writeJSONUnmarshalMethod(enum)
// 	g.writeScanMethod(enum)
// 	g.writeValueMethod(enum)
// 	g.writeBinaryMarshalMethod(enum)
// 	g.writeBinaryUnmarshalMethod(enum)
// 	g.writeTextMarshalMethod(enum)
// 	g.writeTextUnmarshalMethod(enum)

// 	// Other utility code
// 	g.writeCompileCheck(enum)
// 	g.writeStringMethod(enum)
// }

// write writes a string to the output writer
// it is a wrapper around the io.Writer interface
func (g *Writer) write(s string) {
	_, _ = g.w.Write([]byte(s))
}

func (g *Writer) writeEnumGenerationRequest(enum enum.GenerationRequest) {
	g.writeGeneratedComments(enum)
	g.writePackageAndImports(enum)
	g.writeWrapperDefinition(enum)
	g.writeContainerDefinition(enum)
	g.writeInvalidEnumDefinition(enum)
	g.writeAllFunction(enum)
	g.writeParseFunction(enum)
	g.writeStringParsingMethod(enum)
	g.writeNumberParsingMethods(enum)
	g.writeExhaustiveFunction(enum)
	g.writeIsValidFunction(enum)
	if enum.Handlers.JSON {
		g.writeJSONMarshalMethod(enum)
		g.writeJSONUnmarshalMethod(enum)
	}
	if enum.Handlers.Text {
		g.writeTextMarshalMethod(enum)
		g.writeTextUnmarshalMethod(enum)
	}
	if enum.Handlers.SQL {
		g.writeScanMethod(enum)
		g.writeValueMethod(enum)
	}
	if enum.Handlers.Binary {
		g.writeBinaryMarshalMethod(enum)
		g.writeBinaryUnmarshalMethod(enum)
	}
	if enum.Handlers.YAML {
		g.writeYAMLMarshalMethod(enum)
		g.writeYAMLUnmarshalMethod(enum)
	}
	g.writeStringMethod(enum)
	g.writeCompileCheck(enum)
}

var (
	jsonMarshalStr = `
func (p {{ .EnumType }}) MarshalJSON() ([]byte, error) {
	return []byte( p.String()), nil 
}
	`
	jsonMarshalTemplate = template.Must(template.New("jsonMarshal").Parse(jsonMarshalStr))

	jsonUnmarshalStr = `
func (p *{{ .EnumType }}) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(bytes.Trim(b, "\""), "\"")
	newp, err := Parse{{ .EnumType }}(b)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}
`
	jsonUnmarshalTemplate = template.Must(template.New("jsonUnmarshal").Parse(jsonUnmarshalStr))
	textMarshalStr        = `
	func (p {{ .EnumType }}) MarshalText() ([]byte, error) {
		return []byte(p.String()), nil
	}
`
)

type jsonMarshalFunctionData struct {
	EnumType string
	EnumName string
}

func (g *Writer) writeJSONMarshalMethod(rep enum.GenerationRequest) {
	d := jsonMarshalFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
		EnumName: strings.ToUpper(rep.EnumIota.Type),
	}
	jsonMarshalTemplate.Execute(g.w, d)
}

func (g *Writer) writeJSONUnmarshalMethod(rep enum.GenerationRequest) {
	d := jsonMarshalFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
	}
	jsonUnmarshalTemplate.Execute(g.w, d)
}

var (
	textMarshalTemplate = template.Must(template.New("textMarshal").Parse(textMarshalStr))

	textUnmarshalStr = `
	func (p *{{ .EnumType }}) UnmarshalText(b []byte) error {
		newp, err := Parse{{ .EnumType }}(b)
		if err != nil {
			return err
		}
		*p = newp
		return nil
	}
`
	textUnmarshalTemplate = template.Must(template.New("textUnmarshal").Parse(textUnmarshalStr))
)

type textMarshalFunctionData struct {
	EnumType string
}

func (g *Writer) writeTextMarshalMethod(rep enum.GenerationRequest) {
	d := textMarshalFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
	}
	textMarshalTemplate.Execute(g.w, d)
}

func (g *Writer) writeTextUnmarshalMethod(rep enum.GenerationRequest) {
	d := textMarshalFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
	}
	textUnmarshalTemplate.Execute(g.w, d)
}

var (
	binaryMarshalStr = `
	func (p {{ .EnumType }}) MarshalBinary() ([]byte, error) {
		return []byte(p.String()), nil
	}
`
	binaryMarshalTemplate = template.Must(template.New("binaryMarshal").Parse(binaryMarshalStr))

	binaryUnmarshalStr = `
	func (p *{{ .EnumType }}) UnmarshalBinary(b []byte) error {
		newp, err := Parse{{ .EnumType }}(b)
		if err != nil {
			return err
		}
		*p = newp
		return nil
	}
`
	binaryUnmarshalTemplate = template.Must(template.New("binaryUnmarshal").Parse(binaryUnmarshalStr))
)

type binaryMarshalFunctionData struct {
	EnumType string
}

func (g *Writer) writeBinaryMarshalMethod(rep enum.GenerationRequest) {
	d := binaryMarshalFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
	}
	binaryMarshalTemplate.Execute(g.w, d)
}

func (g *Writer) writeBinaryUnmarshalMethod(rep enum.GenerationRequest) {
	d := binaryMarshalFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
	}
	binaryUnmarshalTemplate.Execute(g.w, d)
}

var (
	yamlMarshalStr = `
	func (p {{ .EnumType }}) MarshalYAML() (interface{}, error) {
		return p.String(), nil
	}
`
	yamlMarshalTemplate = template.Must(template.New("yamlMarshal").Parse(yamlMarshalStr))

	yamlUnmarshalStr = `
	func (p *{{ .EnumType }}) UnmarshalYAML(value *yaml.Node) error {
		newp, err := Parse{{ .EnumType }}(value.Value)
		if err != nil {
			return err
		}
		*p = newp
		return nil
	}
`
	yamlUnmarshalTemplate = template.Must(template.New("yamlUnmarshal").Parse(yamlUnmarshalStr))
)

type yamlMarshalFunctionData struct {
	EnumType string
}

func (g *Writer) writeYAMLMarshalMethod(rep enum.GenerationRequest) {
	d := yamlMarshalFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
	}
	yamlMarshalTemplate.Execute(g.w, d)
}

func (g *Writer) writeYAMLUnmarshalMethod(rep enum.GenerationRequest) {
	d := yamlMarshalFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
	}
	yamlUnmarshalTemplate.Execute(g.w, d)
}

var (
	scanStr = `
func (p *{{ .EnumType }}) Scan(value any) error {
	newp, err := Parse{{ .EnumType }}(value)
	if err != nil {
		return err
	}
	*p = newp
	return nil
}
`
	scanTemplate = template.Must(template.New("scan").Parse(scanStr))

	valueStr = `
func (p {{ .EnumType }}) Value() (driver.Value, error) {
	return p.String(), nil
}
`
	valueTemplate = template.Must(template.New("value").Parse(valueStr))
)

type scanFunctionData struct {
	EnumType string
}

func (g *Writer) writeScanMethod(rep enum.GenerationRequest) {
	d := scanFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
	}
	scanTemplate.Execute(g.w, d)
}

type valueFunctionData = scanFunctionData

func (g *Writer) writeValueMethod(rep enum.GenerationRequest) {
	d := valueFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
	}
	valueTemplate.Execute(g.w, d)
}

var (
	compileCheckStr = `
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
	d := compileCheckData{
		Enums: rep.EnumIota.Enums,
	}
	compileCheckTemplate.Execute(g.w, d)
}

// const statusesName = "FAILEDPASSEDSKIPPEDSCHEDULEDRUNNINGBOOKED"

// var statusesIdx = [...]uint16{0, 6, 12, 19, 28, 35, 41}

// // String returns the string representation of the Status value.
// // For valid values, it returns the name of the constant.
// // For invalid values, it returns a string in the format "statuses(N)",
// // where N is the numeric value.
// func (i status) String() string {
// 	if i < 0 || i > status(len(statusesIdx)-1)+-1 {
// 		return "statuses(" + (strconv.FormatInt(int64(i), 10) + ")")
// 	}
// 	return statusesName[statusesIdx[i]:statusesIdx[i+1]]
// }

var (
	stringMethodStr = `
	const {{ .EnumNames }}Name = "{{ .AllEnumNames }}" 
	var {{ .EnumNames }}Idx = [...]uint16{ {{ .AllEnumIdx }} }

	func (i {{ .EnumType }}) String() string {
		if i < {{ .EnumType }}(len({{ .EnumNames }}Idx)-{{.StartIndex}}) {
			return {{ .EnumNames }}Name[{{ .EnumNames }}Idx[i]:{{ .EnumNames }}Idx[i+{{.StartIndex}}]]	
		}
		return {{ .EnumNames }}Name[{{ .EnumNames }}Idx[i]:{{ .EnumNames }}Idx[i+{{.StartIndex}}]]
`
	stringMethodTemplate = template.Must(template.New("stringMethod").Parse(stringMethodStr))
)

type stringMethodData struct {
	EnumNames    string
	AllEnumNames string
	AllEnumIdx   string
	EnumType     string
	StartIndex   int
}

func (g *Writer) writeStringMethod(rep enum.GenerationRequest) {
	edefs := enumDefinitions(rep)
	var names bytes.Buffer
	var indexes bytes.Buffer
	var startIdx, endIdx int
	for _, e := range edefs {
		if e.Valid {
			startIdx = endIdx
			endIdx = startIdx + len(e.Aliases[0])
			names.WriteString(e.Aliases[0])
			indexes.WriteString(fmt.Sprintf("%d, ", startIdx))
		}
	}
	d := stringMethodData{
		EnumNames:    strings.ToUpper(rep.EnumIota.Type),
		AllEnumNames: names.String(),
		AllEnumIdx:   indexes.String(),
		EnumType:     strings.Camel(rep.EnumIota.Type),
		StartIndex:   startIdx,
	}

	stringMethodTemplate.Execute(g.w, d)

}

var (
	isValidStr = `
var valid{{ .EnumTypes }} = map[{{ .EnumType }}]bool{
	{{- range .Enums }}
	{{ $.EnumTypes }}.{{ .EnumNameIdentifier }}: {{ .Valid }},
	{{- end }}
}

// IsValid checks whether the {{ .EnumType }} value is valid.
// A valid value is one that is defined in the original enum and not marked as invalid.
func (p {{ .EnumType }}) IsValid() bool {
	return valid{{ .EnumTypes }}[p]
}
`
	isValidTemplate = template.Must(template.New("isValid").Parse(isValidStr))
)

type isValidFunctionData struct {
	EnumTypes string
	EnumType  string
	Enums     []enumDefinition
}

func (g *Writer) writeIsValidFunction(rep enum.GenerationRequest) {
	edefs := enumDefinitions(rep)
	eType := strings.Camel(rep.EnumIota.Type)
	isValidData := isValidFunctionData{
		EnumTypes: strings.Plural(eType),
		EnumType:  eType,
		Enums:     edefs,
	}
	isValidTemplate.Execute(g.w, isValidData)
}

func (g *Writer) writeNumberParsingMethods(rep enum.GenerationRequest) {
	data := parseNumberFunctionData{
		HasStartIndex: rep.EnumIota.StartIndex > 0,
		StartIndex:    rep.EnumIota.StartIndex,
		MapEnumType:   wrapperType(rep.EnumIota.Type),
		EnumType:      containerName(rep),
	}
	g.writeNumberParsingMethod(data)
}

type invalidEnumDefinition struct {
	EnumType string
}

var (
	invalidEnumStr = `
	var invalid{{ .EnumType }} = {{ .EnumType }}{}
	`
	invalidEnumTemplate = template.Must(template.New("invalidEnum").Parse(invalidEnumStr))
)

func (g *Writer) writeInvalidEnumDefinition(enum enum.GenerationRequest) {
	invalidEnumDefinition := invalidEnumDefinition{
		EnumType: wrapperType(enum.EnumIota.Type),
	}
	invalidEnumTemplate.Execute(g.w, invalidEnumDefinition)
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
type {{ .WrapperName }} struct {
  {{ .EnumType }}
  {{- range .Fields }}
  {{ .Name }} {{ .Type }}
  {{- end }}
}

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
			EnumType: wType,
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
	wrapperDefinitionTemplate.Execute(g.w, d)
}

func wrapperName(enum string) string {
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

func containerName(enum enum.GenerationRequest) string {
	cName := strings.Pluralise(enum.EnumIota.Type)
	cName = strings.Camel(cName)
	return cName
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
// code generated by goenums {.Version} at {.Time}.
// github.com/zarldev/goenums
//
// using the command:
// goenums {.Command} {.SourceFilename}
	`
	generatedCommentTemplate = template.Must(
		template.New("generatedComment").Parse(generatedCommentStr))
)

func (g *Writer) writeGeneratedComments(rep enum.GenerationRequest) {
	generatedCommentTemplate.Execute(g.w, generatedComment{
		Version:        rep.Version,
		Time:           time.Now().Format(time.RFC3339),
		Command:        rep.Command(),
		SourceFilename: rep.SourceFilename,
	})
}

type packageImport struct {
	PackageName string
	Imports     []string
}

var (
	packageImportStr = `
package {{ .PackageName }}

import (
{{- range .Imports }}
"{{ . }}"
{{- end }}
)
`
	packageImportTemplate = template.Must(template.New("packageImport").Parse(packageImportStr))
)

func (g *Writer) writePackageAndImports(rep enum.GenerationRequest) {
	imports := []string{"fmt", "strconv", "bytes", "database/sql/driver", "math"}
	if rep.CaseInsensitive {
		imports = append(imports, "strings")
	}
	if !rep.Legacy {
		imports = append(imports, "iter")
	}
	packageImportTemplate.Execute(g.w, packageImport{
		PackageName: rep.Package,
		Imports:     imports,
	})
}

type containerDefinition struct {
	ContainerName string
	ContainerType string
	EnumDefs      []enumDefinition
}

var (
	containerDefinitionStr = `
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
		ContainerType: containerType(rep),
		ContainerName: containerName(rep),
		EnumDefs:      edefs,
	}
	containerDefinitionTemplate.Execute(g.w, cdef)
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
		aliases = append(aliases, e.Name)
		if rep.CaseInsensitive {
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
			EnumType:           strings.Camel(rep.EnumIota.Type),
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
	EnumType      string
	EnumDefs      []enumDefinition
}

var (
	allFunctionStr = `
func ({{.Receiver}} {{.ContainerType}}) allSlice() []{{.EnumType}} {
	return []{{.EnumType}}{
		{{-  range .EnumDefs}}
		{{$.ContainerName}}.{{.EnumNameIdentifier}},
		{{- end}}
	}
}
{{- if .Legacy}}
func ({{.Receiver}} {{.ContainerType}}) All() []{{.EnumType}} {
	return {{.Receiver}}.allSlice()
}
{{- else}}
func ({{.Receiver}} {{.ContainerType}}) All() iter.Seq[{{.EnumType}}] {
	return func(yield func({{.EnumType}}) bool) {
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
	edefs := enumDefinitions(rep)
	r := strings.Lower1stCharacter(rep.EnumIota.Type)[0]
	allData := allFunctionData{
		Receiver:      string(r),
		ContainerType: containerType(rep),
		ContainerName: containerName(rep),
		EnumType:      strings.Camel(rep.EnumIota.Type),
		EnumDefs:      edefs,
		Legacy:        rep.Legacy,
	}
	allFunctionTemplate.Execute(g.w, allData)
}

type parseFunctionData struct {
	EnumType string
	Enums    []enum.Enum
}

var (
	parseFunctionStr = `
func Parse{{.EnumType}}(input any) ({{.EnumType}}, error) {
	var res = invalid{{.EnumType}}
	switch v := input.(type) {
	case {{.EnumType}}:
		return v, nil
	case string:
		res = stringTo{{.EnumType}}(v)
	case fmt.Stringer:
		res = stringTo{{.EnumType}}(v.String())
	case []byte:
		res = stringTo{{.EnumType}}(string(v))
	case int:
		res = numberTo{{.EnumType}}(v)
	case int8:
		res = numberTo{{.EnumType}}(v)
	case int16:
		res = numberTo{{.EnumType}}(v)
	case int32:
		res = numberTo{{.EnumType}}(v)
	case int64:
		res = numberTo{{.EnumType}}(v)
	case uint:
		res = numberTo{{.EnumType}}(v)
	case uint8:
		res = numberTo{{.EnumType}}(v)
	case uint16:
		res = numberTo{{.EnumType}}(v)
	case uint32:
		res = numberTo{{.EnumType}}(v)
	case uint64:
		res = numberTo{{.EnumType}}(v)
	case float32:
		res = numberTo{{.EnumType}}(v)
	case float64:
		res = numberTo{{.EnumType}}(v)
	default:
		return res, fmt.Errorf("invalid type %T", input)
	}
	return res, nil
}
`
	parseFunctionTemplate = template.Must(template.New("parseFunction").Parse(parseFunctionStr))
)

func (g *Writer) writeParseFunction(rep enum.GenerationRequest) error {
	data := parseFunctionData{
		EnumType: strings.Camel(rep.EnumIota.Type),
		Enums:    rep.EnumIota.Enums,
	}
	return parseFunctionTemplate.Execute(g.w, data)
}

type parseStringFunctionData struct {
	EnumNameMap     string
	MapEnumType     string
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
var {{.EnumNameMap}} = map[string]{{.MapEnumType}}{
{{- range .Enums }}
  {{- $enum := . }}
  {{- range .Aliases }}
    "{{ . }}": {{ $.EnumType }}.{{ $enum.EnumNameIdentifier }},
  {{- end }}
{{- end }}
}

func stringTo{{.MapEnumType}}(s string) {{.MapEnumType}} {
  if t, ok := {{.EnumNameMap}}[s]; ok {
    return t
  }
  return invalid{{.MapEnumType}}
}
`
	parseStringFunctionTemplate = template.Must(template.New("parseStringFunction").Parse(parseStringFunctionStr))
)

func (g *Writer) writeStringParsingMethod(rep enum.GenerationRequest) {
	edefs := enumDefinitions(rep)
	data := parseStringFunctionData{
		MapEnumType:     wrapperType(rep.EnumIota.Type),
		EnumNameMap:     enumNameMap(rep.EnumIota.Type),
		EnumType:        containerName(rep),
		Enums:           edefs,
		CaseInsensitive: rep.CaseInsensitive,
	}
	parseStringFunctionTemplate.Execute(g.w, data)
}

type parseNumberFunctionData struct {
	MapEnumType   string
	EnumType      string
	StartIndex    int
	HasStartIndex bool
}

var (
	parseIntegerGenericFunctionTemplate = template.Must(template.New("parseIntegerGenericFunction").Parse(`

type integer interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type float interface {
    ~float32 | ~float64
}

type number interface {
    integer | float
}

// To{{.MapEnumType}} converts a numeric value to a {{.MapEnumType}}
func numberTo{{.MapEnumType}}[T number](num T) {{.MapEnumType}} {
	f := float64(num)
    if math.Floor(f) != f {
        return invalid{{.MapEnumType}}
    }
	i := int(f)
	if i <= 0 || i > len({{.EnumType}}.allSlice()) {
		return invalid{{.MapEnumType}}
	}
	{{- if .StartIndex }}
	return {{.EnumType}}.allSlice()[i-{{.StartIndex}}]
	{{- else }}
	return {{.EnumType}}.allSlice()[i]
	{{- end }}
}

`))
)

func (g *Writer) writeNumberParsingMethod(data parseNumberFunctionData) {
	parseIntegerGenericFunctionTemplate.Execute(g.w, data)
}

func enumNameMap(enumType string) string {
	return strings.Pluralise(enumType) + "NameMap"
}

var (
	exhaustiveStr = `
	func Exhaustive{{ .EnumTypes }}(f func({{ .EnumType }})) {
		for _, p := range {{ .EnumTypes }}.allSlice() {
			f(p)
		}
	}
	`
	exhaustiveTemplate = template.Must(template.New("exhaustive").Parse(exhaustiveStr))
)

type exhaustiveFunctionData struct {
	EnumTypes string
	EnumType  string
	Enums     []enumDefinition
}

func (g *Writer) writeExhaustiveFunction(rep enum.GenerationRequest) {
	edefs := enumDefinitions(rep)
	eType := strings.Camel(rep.EnumIota.Type)
	exhaustiveData := exhaustiveFunctionData{
		EnumType:  eType,
		EnumTypes: strings.Plural(eType),
		Enums:     edefs,
	}
	exhaustiveTemplate.Execute(g.w, exhaustiveData)
}

// // New combined function for all parsing methods
// func (g *Writer) writeParsingMethods(enum enum.Representation) {
// 	g.writeParseFunction(enum)

// 	g.writeStringParsingMethod(enum)
// 	g.writeIntParsingMethod(enum)
// }

// func (g *Writer) writeParseFunction(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// Parse" + rep.TypeInfo.Camel + " converts various input types to a " +
// 		rep.TypeInfo.Camel + " value.\n")
// 	b.WriteString("// It accepts the following types:\n")
// 	b.WriteString("// - " + rep.TypeInfo.Camel + ": returns the value directly\n")
// 	b.WriteString("// - string: parses the string representation\n")
// 	b.WriteString("// - []byte: converts to string and parses\n")
// 	b.WriteString("// - fmt.Stringer: uses the String() result for parsing\n")
// 	b.WriteString("// - int/int32/int64: converts the integer to the corresponding enum value\n")
// 	if rep.Failfast {
// 		b.WriteString("//\n")
// 		b.WriteString("// If the input cannot be converted to a valid " + rep.TypeInfo.Camel +
// 			" value, it returns\n")
// 		b.WriteString("// the invalid" + rep.TypeInfo.Camel + " value and an error.\n")
// 	} else {
// 		b.WriteString("//\n")
// 		b.WriteString("// If the input cannot be converted to a valid " + rep.TypeInfo.Camel +
// 			" value, it returns\n")
// 		b.WriteString("// the invalid" + rep.TypeInfo.Camel + " value without an error.\n")
// 	}
// 	b.WriteString("func Parse" + rep.TypeInfo.Camel + "(a any) (" + rep.TypeInfo.Camel + ", error) {\n")
// 	b.WriteString("\tres := invalid" + rep.TypeInfo.Camel + "\n")
// 	b.WriteString("\tswitch v := a.(type) {\n")
// 	b.WriteString("\tcase " + rep.TypeInfo.Camel + ":\n")
// 	b.WriteString("\t\treturn v, nil\n")
// 	b.WriteString("\tcase []byte:\n")
// 	b.WriteString("\t\tres = stringTo" + rep.TypeInfo.Camel + "(string(v))\n")
// 	b.WriteString("\tcase string:\n")
// 	b.WriteString("\t\tres = stringTo" + rep.TypeInfo.Camel + "(v)\n")
// 	b.WriteString("\tcase fmt.Stringer:\n")
// 	b.WriteString("\t\tres = stringTo" + rep.TypeInfo.Camel + "(v.String())\n")
// 	b.WriteString("\tcase int:\n")
// 	b.WriteString("\t\tres = intTo" + rep.TypeInfo.Camel + "(v)\n")
// 	b.WriteString("\tcase int64:\n")
// 	b.WriteString("\t\tres = intTo" + rep.TypeInfo.Camel + "(int(v))\n")
// 	b.WriteString("\tcase int32:\n")
// 	b.WriteString("\t\tres = intTo" + rep.TypeInfo.Camel + "(int(v))\n")
// 	b.WriteString("\t}\n")
// 	if rep.Failfast {
// 		b.WriteString("\tif res == invalid" + rep.TypeInfo.Camel + " {\n")
// 		errorMsg := fmt.Sprintf("failed to parse %s value - invalid input: %%v", rep.TypeInfo.Camel)
// 		b.WriteString("\t\treturn res, fmt.Errorf(\"" + errorMsg + "\", a)\n")
// 		b.WriteString("\t}\n")
// 	}
// 	b.WriteString("\treturn res, nil\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// // writeStringParsingMethod creates a function that maps strings to enum values
// // It now handles both primary aliases and additional aliases for each enum value
// func (g *Writer) writeStringParsingMethod(rep enum.Representation) {
// 	b := strings.NewEnumBuilder(rep)
// 	b.WriteString("// stringTo" + rep.TypeInfo.Camel + " is an internal function that converts a string to a " +
// 		rep.TypeInfo.Camel + " value.\n")
// 	b.WriteString("// It uses a predefined mapping of string representations to enum values.\n")
// 	if rep.CaseInsensitive {
// 		b.WriteString("// This implementation is case-insensitive.\n")
// 	}
// 	b.WriteString("var (\n")
// 	b.WriteString(fmt.Sprintf("\t%sNameMap = map[string]%s{\n", rep.TypeInfo.Lower, rep.TypeInfo.Camel))
// 	seen := make(map[string]bool)

// 	for _, enum := range rep.Enums {
// 		alias := enum.Info.Alias
// 		if alias != "" {
// 			b.WriteString(fmt.Sprintf("\t\t%q: %s.%s, // primary alias\n",
// 				alias, rep.TypeInfo.PluralCamel, enum.Info.Upper))
// 			seen[alias] = true
// 			g.caseInsensitiveName(rep, alias, seen, b, enum)
// 		}
// 		if len(enum.Info.Aliases) > 0 {
// 			for _, alias := range enum.Info.Aliases {
// 				if alias != "" && !seen[alias] {
// 					b.WriteString(fmt.Sprintf("\t\t%q: %s.%s, // additional alias\n",
// 						alias, rep.TypeInfo.PluralCamel, enum.Info.Upper))
// 					seen[alias] = true

// 					g.caseInsensitiveName(rep, alias, seen, b, enum)
// 				}
// 			}
// 		}
// 		if !seen[enum.Info.Name] {
// 			b.WriteString(fmt.Sprintf("\t\t%q: %s.%s, // enum name\n",
// 				enum.Info.Name, rep.TypeInfo.PluralCamel, enum.Info.Upper))
// 			seen[enum.Info.Name] = true
// 			g.caseInsensitiveName(rep, alias, seen, b, enum)
// 		}
// 	}

// 	b.WriteString("\t}\n")
// 	b.WriteString(")\n\n")

// 	// Generate the actual string-to-enum function
// 	b.WriteString("func stringTo" + rep.TypeInfo.Camel + "(s string) " + rep.TypeInfo.Camel + " {\n")
// 	b.WriteString(fmt.Sprintf("\tif v, ok := %sNameMap[s]; ok {\n", rep.TypeInfo.Lower))
// 	b.WriteString("\t\treturn v\n")
// 	b.WriteString("\t}\n")

// 	// Handle case-insensitive lookup if enabled
// 	if rep.CaseInsensitive {
// 		b.WriteString("\tlwr := strings.ToLower(s)\n")
// 		b.WriteString("\tif lwr != s {\n")
// 		b.WriteString(fmt.Sprintf("\t\tif v, ok := %sNameMap[lwr]; ok {\n", rep.TypeInfo.Lower))
// 		b.WriteString("\t\t\treturn v\n")
// 		b.WriteString("\t\t}\n")
// 		b.WriteString("\t}\n")
// 	}

// 	b.WriteString("\treturn invalid" + rep.TypeInfo.Camel + "\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// // Write implements enum.RepresentationWriter.
// func (g *Writer) Write2(ctx context.Context,
// 	enums []enum.Representation) error {
// 	for _, enumRep := range enums {
// 		if ctx.Err() != nil {
// 			return ctx.Err()
// 		}
// 		dirPath := filepath.Dir(enumRep.SourceFilename)
// 		outFilename := fmt.Sprintf("%s_enums.go", enumRep.OutputFilename)
// 		if strings.Contains(outFilename, " ") || strings.Contains(outFilename, "/") {
// 			return fmt.Errorf("%w: '%s' contains invalid characters", ErrWriteGoFile, outFilename)
// 		}
// 		fullPath := filepath.Join(dirPath, outFilename)
// 		err := file.WriteToFileAndFormatFS(ctx, g.fs, fullPath, true,
// 			func(w io.Writer) error {
// 				g.w = w
// 				g.writeEnumFile(enumRep)
// 				return nil
// 			})
// 		if err != nil {
// 			return fmt.Errorf("%w: %s: %w", ErrWriteGoFile, fullPath, err)
// 		}
// 	}
// 	return nil
// }

// func (*Writer) caseInsensitiveName(rep enum.Representation, alias string, seen map[string]bool, b *strings.EnumBuilder, enum enum.Enum) {
// 	if rep.CaseInsensitive {
// 		lowercase := strings.ToLower(alias)
// 		if lowercase != alias && !seen[lowercase] {
// 			b.WriteString(fmt.Sprintf("\t\t%q: %s.%s, // Case-insensitive\n",
// 				lowercase, rep.TypeInfo.PluralCamel, enum.Info.Upper))
// 			seen[lowercase] = true
// 		}
// 	}
// }

// // Change from setupIntToTypeMethod to writeIntParsingMethod
// func (g *Writer) writeIntParsingMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// intTo" + rep.TypeInfo.Camel + " converts an integer to a " + rep.TypeInfo.Camel + " value.\n")
// 	b.WriteString("// The integer is treated as the ordinal position in the enum sequence.\n")
// 	if rep.TypeInfo.Index != 0 {
// 		b.WriteString("// The input is adjusted by -" + strconv.Itoa(rep.TypeInfo.Index) + " to account for the enum starting value.\n")
// 	}
// 	b.WriteString("// If the integer doesn't correspond to a valid enum value, invalid" + rep.TypeInfo.Camel + " is returned.\n")
// 	b.WriteString("func intTo" + rep.TypeInfo.Camel + "(i int) " + rep.TypeInfo.Camel + " {\n")
// 	if rep.TypeInfo.Index != 0 {
// 		b.WriteString("\ti -= " + strconv.Itoa(rep.TypeInfo.Index) + "\n")
// 	}
// 	b.WriteString("\tif i < 0 || i >= len(" + rep.TypeInfo.PluralCamel + ".allSlice()) {\n")
// 	b.WriteString("\t\treturn invalid" + rep.TypeInfo.Camel + "\n")
// 	b.WriteString("\t}\n")
// 	b.WriteString("\treturn " + rep.TypeInfo.PluralCamel + ".allSlice()[i]\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeScanMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// Scan implements the sql.Scanner interface for " + rep.TypeInfo.Camel + ".\n")
// 	b.WriteString("// This allows " + rep.TypeInfo.Camel + " values to be scanned directly from database queries.\n")
// 	b.WriteString("// It supports scanning from strings, []byte, or integers.\n")
// 	b.WriteString("func (p *" + rep.TypeInfo.Camel + ") Scan(value any) error {\n")
// 	b.WriteString("\tnewp, err := Parse" + rep.TypeInfo.Camel + "(value)\n")
// 	b.WriteString("\tif err != nil {\n")
// 	b.WriteString("\t\treturn err\n")
// 	b.WriteString("\t}\n")
// 	b.WriteString("\t*p = newp\n")
// 	b.WriteString("\treturn nil\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeValueMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// Value implements the driver.Valuer interface for " + rep.TypeInfo.Camel + ".\n")
// 	b.WriteString("// This allows " + rep.TypeInfo.Camel + " values to be saved to databases.\n")
// 	b.WriteString("// The value is stored as a string representation of the enum.\n")
// 	b.WriteString("func (p " + rep.TypeInfo.Camel + ") Value() (driver.Value, error) {\n")
// 	b.WriteString("\treturn p.String(), nil\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }
// func (g *Writer) generateIndexAndNameRun(rep enum.Representation) (string, string) {
// 	var namesBuilder strings.EnumBuilder
// 	positions := make([]int, 1, len(rep.Enums)+1)
// 	positions[0] = 0
// 	for _, enum := range rep.Enums {
// 		namesBuilder.WriteString(enum.Info.Alias)
// 		positions = append(positions, namesBuilder.Len())
// 	}
// 	nameStr := namesBuilder.String()
// 	nameConst := fmt.Sprintf("%sName = %q\n", rep.TypeInfo.Lower, nameStr)

// 	var indexBuilder strings.EnumBuilder
// 	indexBuilder.WriteString(fmt.Sprintf("%sIdx = [...]uint16{", rep.TypeInfo.Lower))

// 	for i, pos := range positions {
// 		if i > 0 {
// 			indexBuilder.WriteString(", ")
// 		}
// 		indexBuilder.WriteString(strconv.Itoa(pos))
// 	}
// 	indexBuilder.WriteString("}\n")
// 	return indexBuilder.String(), nameConst
// }
// func (g *Writer) writeStringMethod(rep enum.Representation) {
// 	indexVar, nameConst := g.generateIndexAndNameRun(rep)
// 	var b strings.EnumBuilder
// 	b.WriteString("const " + nameConst)
// 	b.WriteString("var " + indexVar)
// 	b.WriteString("// String returns the string representation of the " + rep.TypeInfo.Camel + " value.\n")
// 	b.WriteString("// For valid values, it returns the name of the constant.\n")
// 	b.WriteString("// For invalid values, it returns a string in the format \"" + rep.TypeInfo.Lower + "(N)\",\n")
// 	b.WriteString("// where N is the numeric value.\n")
// 	b.WriteString("func (i " + rep.TypeInfo.Name + ") String() string {\n")
// 	b.WriteString(fmt.Sprintf("\tif i < %d || i > %s(len(%sIdx)-1)+%d {\n",
// 		rep.MinValue, rep.TypeInfo.Name, rep.TypeInfo.Lower, rep.MinValue-1))
// 	b.WriteString("\t\treturn \"" + rep.TypeInfo.Lower + "(\" + (strconv.FormatInt(int64(i), 10) + \")\")\n")
// 	b.WriteString("\t}\n")

// 	// Calculate zero-based index offset based on MinValue
// 	if rep.MinValue == 0 {
// 		// zero-based enums: index directly
// 		b.WriteString(fmt.Sprintf("\treturn %sName[%sIdx[i]:%sIdx[i+1]]\n",
// 			rep.TypeInfo.Lower, rep.TypeInfo.Lower, rep.TypeInfo.Lower))
// 	} else {
// 		// one-based or higher: subtract MinValue to get zero-based index
// 		b.WriteString(fmt.Sprintf("\tindex := int(i) - %d\n", rep.MinValue))
// 		b.WriteString(fmt.Sprintf("\treturn %sName[%sIdx[index]:%sIdx[index+1]]\n",
// 			rep.TypeInfo.Lower, rep.TypeInfo.Lower, rep.TypeInfo.Lower))
// 	}
// 	b.WriteString("}\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeCompileCheck(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("func _() {\n")
// 	b.WriteString("\t// An \"invalid array index\" compiler error signifies that the constant values have changed.\n")
// 	b.WriteString("\t// Re-run the goenums command to generate them again.\n")
// 	b.WriteString("\t// Does not identify newly added constant values unless order changes\n")
// 	b.WriteString("\tvar x [1]struct{}\n")
// 	for _, v := range rep.Enums {
// 		b.WriteString(fmt.Sprintf("\t_ = x[%s - %d]\n", v.Info.Name, v.TypeInfo.Index))
// 	}
// 	b.WriteString("}\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeJSONMarshalMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// MarshalJSON implements the json.Marshaler interface for " + rep.TypeInfo.Camel + ".\n")
// 	b.WriteString("// The enum value is encoded as its string representation.\n")
// 	b.WriteString("func (p " + rep.TypeInfo.Camel + ") MarshalJSON() ([]byte, error) {\n")
// 	b.WriteString("\treturn []byte(`\"`+p.String() + `\"`), nil\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeTextMarshalMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// MarshalText implements the encoding.TextMarshaler interface for " + rep.TypeInfo.Camel + ".\n")
// 	b.WriteString("// The enum value is encoded as its string representation.\n")
// 	b.WriteString("func (p " + rep.TypeInfo.Camel + ") MarshalText() ([]byte, error) {\n")
// 	b.WriteString("\treturn []byte(p.String()), nil\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeTextUnmarshalMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// UnmarshalText implements the encoding.TextUnmarshaler interface for " + rep.TypeInfo.Camel + ".\n")
// 	b.WriteString("// It supports unmarshaling from a string representation of the enum.\n")
// 	b.WriteString("func (p *" + rep.TypeInfo.Camel + ") UnmarshalText(b []byte) error {\n")
// 	b.WriteString("\tnewp, err := Parse" + rep.TypeInfo.Camel + "(b)\n")
// 	b.WriteString("\tif err != nil {\n")
// 	b.WriteString("\t\treturn err\n")
// 	b.WriteString("\t}\n")
// 	b.WriteString("\t*p = newp\n")
// 	b.WriteString("\treturn nil\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeBinaryMarshalMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// MarshalBinary implements the encoding.BinaryMarshaler interface for " + rep.TypeInfo.Camel + ".\n")
// 	b.WriteString("// It encodes the enum value as a byte slice.\n")
// 	b.WriteString("func (p " + rep.TypeInfo.Camel + ") MarshalBinary() ([]byte, error) {\n")
// 	b.WriteString("\treturn []byte(p.String()), nil\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeBinaryUnmarshalMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for " + rep.TypeInfo.Camel + ".\n")
// 	b.WriteString("// It decodes the enum value from a byte slice.\n")
// 	b.WriteString("func (p *" + rep.TypeInfo.Camel + ") UnmarshalBinary(b []byte) error {\n")
// 	b.WriteString("\tnewp, err := Parse" + rep.TypeInfo.Camel + "(b)\n")
// 	b.WriteString("\tif err != nil {\n")
// 	b.WriteString("\t\treturn err\n")
// 	b.WriteString("\t}\n")
// 	b.WriteString("\t*p = newp\n")
// 	b.WriteString("\treturn nil\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeJSONUnmarshalMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// UnmarshalJSON implements the json.Unmarshaler interface for " + rep.TypeInfo.Camel + ".\n")
// 	b.WriteString("// It supports unmarshaling from a string representation of the enum.\n")
// 	b.WriteString("func (p *" + rep.TypeInfo.Camel + ") UnmarshalJSON(b []byte) error {\n")
// 	b.WriteString("b = bytes.Trim(bytes.Trim(b, `\"`), ` `)\n")
// 	b.WriteString("\tnewp, err := Parse" + rep.TypeInfo.Camel + "(b)\n")
// 	b.WriteString("\tif err != nil {\n")
// 	b.WriteString("\t\treturn err\n")
// 	b.WriteString("\t}\n")
// 	b.WriteString("\t*p = newp\n")
// 	b.WriteString("\treturn nil\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeIsValidMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// valid" + rep.TypeInfo.PluralCamel + " is a map of valid " + rep.TypeInfo.Camel + " values.\n")
// 	b.WriteString("var valid" + rep.TypeInfo.PluralCamel + " = map[" + rep.TypeInfo.Camel + "]bool{\n")
// 	for _, info := range rep.Enums {
// 		if info.Info.Valid {
// 			b.WriteString("\t" + rep.TypeInfo.PluralCamel + "." + info.Info.Upper + ": true,\n")
// 		}
// 	}
// 	b.WriteString("}\n\n")
// 	b.WriteString("// IsValid checks whether the " + rep.TypeInfo.Camel + " value is valid.\n")
// 	b.WriteString("// A valid value is one that is defined in the original enum and not marked as invalid.\n")
// 	b.WriteString("func (p " + rep.TypeInfo.Camel + ") IsValid() bool {\n")
// 	b.WriteString("\treturn valid" + rep.TypeInfo.PluralCamel + "[p]\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeExhaustiveMethod(rep enum.Representation) {
// 	l := len(rep.Enums)
// 	var b strings.EnumBuilder
// 	b.WriteString("// Exhaustive" + rep.TypeInfo.PluralCamel + " calls the provided function once for each valid " + rep.TypeInfo.PluralCamel + " value.\n")
// 	b.WriteString("// This is useful for switch statement exhaustiveness checking and for processing all enum values.\n")
// 	b.WriteString("// Example usage:\n")
// 	b.WriteString("// ```\n")
// 	b.WriteString("// Exhaustive" + rep.TypeInfo.PluralCamel + "(func(x " + rep.TypeInfo.Camel + ") {\n")
// 	b.WriteString("//     switch x {\n")
// 	b.WriteString("//     case " + rep.TypeInfo.PluralCamel + "." + rep.Enums[l-1].Info.Camel + ":\n")
// 	b.WriteString("//         // handle " + rep.Enums[l-1].Info.Camel + "\n")
// 	b.WriteString("//     }\n")
// 	b.WriteString("// })\n")
// 	b.WriteString("// ```\n")
// 	b.WriteString("func Exhaustive" + rep.TypeInfo.PluralCamel + "(f func(" + rep.TypeInfo.Camel + ")) {\n")
// 	b.WriteString("\tfor _, p := range " + rep.TypeInfo.PluralCamel + ".allSlice() {\n")
// 	b.WriteString("\t\tf(p)\n")
// 	b.WriteString("\t}\n")
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeImports(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("import (\n")
// 	b.WriteString("\t\"fmt\"\n")
// 	b.WriteString("\t\"strconv\"\n")
// 	if rep.CaseInsensitive {
// 		b.WriteString("\t\"strings\"\n")
// 	}
// 	b.WriteString("\t\"bytes\"\n")
// 	if !rep.Legacy {
// 		b.WriteString("\t\"iter\"\n")
// 	}
// 	b.WriteString("\t\"database/sql/driver\"\n")
// 	importedPkgs := make(map[string]bool)
// 	for _, pair := range rep.TypeInfo.NameTypePair {
// 		isTypeUsed := false
// 		for _, enum := range rep.Enums {
// 			for _, enumPair := range enum.TypeInfo.NameTypePair {
// 				if enumPair.Name == pair.Name {
// 					isTypeUsed = true
// 					break
// 				}
// 			}
// 			if isTypeUsed {
// 				break
// 			}
// 		}
// 		if isTypeUsed && strings.Contains(pair.Type, ".") {
// 			pkg := strings.Split(pair.Type, ".")[0]
// 			if !importedPkgs[pkg] {
// 				b.WriteString("\t\"" + pkg + "\"\n")
// 				importedPkgs[pkg] = true
// 			}
// 		}
// 	}
// 	b.WriteString(")\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeWrapperType(rep enum.Representation) {
// 	var b strings.EnumBuilder

// 	b.WriteString("type " + rep.TypeInfo.Camel + " struct {\n")
// 	b.WriteString(rep.TypeInfo.Name + "\n")
// 	for _, pair := range rep.TypeInfo.NameTypePair {
// 		b.WriteString("\t" + pair.Name + " " + pair.Type + "\n")
// 	}
// 	b.WriteString("}\n\n")
// 	b.WriteString("type " + rep.TypeInfo.Lower + "Container struct {\n")
// 	for _, info := range rep.Enums {
// 		b.WriteString("\t" + info.Info.Upper + " " + info.TypeInfo.Camel + "\n")
// 	}
// 	b.WriteString("}\n\n")
// 	b.WriteString("var " + rep.TypeInfo.PluralCamel + " = " + rep.TypeInfo.Lower + "Container{\n")
// 	for _, info := range rep.Enums {
// 		if info.Info.Valid {
// 			b.WriteString("\t" + info.Info.Upper + ": " + info.TypeInfo.Camel + "{ \n\t" + info.TypeInfo.Name + ":" + info.Info.Name + ",\n")
// 			for _, ntp := range info.TypeInfo.NameTypePair {
// 				b.WriteString(ntp.Name + ": " + ntp.Value + ",\n")
// 			}
// 			b.WriteString("},\n")
// 		}
// 	}
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }
// func (g *Writer) writeAllMethod(rep enum.Representation) {
// 	var b strings.EnumBuilder
// 	b.WriteString("// allSlice is an internal method that returns all valid " + rep.TypeInfo.Camel + " values as a slice.\n")
// 	b.WriteString("func (c " + rep.TypeInfo.Lower + "Container) allSlice() []" + rep.TypeInfo.Camel + " {\n")
// 	b.WriteString("\treturn []" + rep.TypeInfo.Camel + "{\n")
// 	for _, info := range rep.Enums {
// 		if info.Info.Valid {
// 			b.WriteString("\t\tc." + info.Info.Upper + ",\n")
// 		}
// 	}
// 	b.WriteString("\t}\n")
// 	b.WriteString("}\n\n")
// 	b.WriteString("// AllSlice returns all valid " + rep.TypeInfo.Camel + " values as a slice.\n")
// 	b.WriteString("func (c " + rep.TypeInfo.Lower + "Container) AllSlice() []" + rep.TypeInfo.Camel + " {\n")
// 	b.WriteString("\treturn c.allSlice()\n")
// 	b.WriteString("}\n\n")
// 	b.WriteString("// All returns all valid " + rep.TypeInfo.Camel + " values.\n")
// 	if !rep.Legacy {
// 		b.WriteString("// In Go 1.23+, this can be used with range-over-function iteration:\n")
// 		b.WriteString("// ```\n")
// 		b.WriteString("// for v := range " + rep.TypeInfo.PluralCamel + ".All() {\n")
// 		b.WriteString("//     // process each enum value\n")
// 		b.WriteString("// }\n")
// 		b.WriteString("// ```\n")
// 	} else {
// 		b.WriteString("// Returns a slice of all valid enum values.\n")
// 	}
// 	if !rep.Legacy {
// 		b.WriteString("func (c " + rep.TypeInfo.Lower + "Container) All() iter.Seq[" + rep.TypeInfo.Camel + "] {\n")
// 		b.WriteString("\treturn func(yield func(" + rep.TypeInfo.Camel + ") bool) {\n")
// 		b.WriteString("\t\tfor _, v := range c.allSlice() {\n")
// 		b.WriteString("\t\t\tif !yield(v) {\n")
// 		b.WriteString("\t\t\t\treturn\n")
// 		b.WriteString("\t\t\t}\n")
// 		b.WriteString("\t\t}\n")
// 		b.WriteString("\t}\n")
// 	} else {
// 		b.WriteString("func (c " + rep.TypeInfo.Lower + "Container) All() []" + rep.TypeInfo.Camel + " {\n")
// 		b.WriteString("\treturn c.allSlice()\n")
// 	}
// 	b.WriteString("}\n\n")
// 	g.write(b.String())
// }

// func (g *Writer) writeInvalidTypeDefinition(rep enum.Representation) {
// 	g.write("// invalid" + rep.TypeInfo.Camel + " represents an invalid or undefined " + rep.TypeInfo.Camel + " value.\n")
// 	g.write("// It is used as a default return value for failed parsing or conversion operations.\n")
// 	g.write("var invalid" + rep.TypeInfo.Camel + " = " + rep.TypeInfo.Camel + "{}\n\n")
// }
