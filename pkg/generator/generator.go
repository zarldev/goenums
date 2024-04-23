// generator package is responsible for parsing the file
// and generating the enum go file.
// It uses the ast package to parse the file and generate the
// enum go file from the parsed information.
// The EnumRepresentation struct is the struct to store the
// information that is to be used in writing the enum to a file.
// ParseAndGenerate function to parse the file and generate the
// enum go file for the enum type with failfast mode flag.
// Provides the failfast mode flag to enable failfast mode.
// This mode is 'error on invalid' and will error instead of
// generating an enum class as 'Invalid' when Parsed.
package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

// camelCase is a Caser for turning strings into camelCase.
func camelCase(in string) string {
	first := strings.ToUpper(in[:1])
	rest := in[1:]
	return first + rest
}

// EnumRepresentation is a struct to store the information to be used in writing the enum to a file.
type EnumRepresentation struct {
	PackageName string
	Failfast    bool
	TypeInfo    typeInfo
	Enums       []Enum
}

// Enum is a struct to store the information for each enum to be written.
type Enum struct {
	Info     info
	TypeInfo typeInfo
	Raw      raw
}

type raw struct {
	// raw comment for the enum
	Comment string
	// raw comment for the type
	TypeComment string
}

type info struct {
	// base info for the enum
	Name          string
	AlternateName string
	Camel         string
	Lower         string
	Upper         string
	Value         int
	// valid or invalid
	Valid bool
}

type typeInfo struct {
	Filename string
	Index    int
	// type name for the enum in different cases
	Name        string
	Camel       string
	Lower       string
	Upper       string
	Plural      string
	PluralCamel string
	// name type pairs for the enum not using iota
	NameTypePairs []nameTypePair
}

// nameTypePair is a struct to store the name and type of the extra values for the enum.
type nameTypePair struct {
	// name of the extra value
	Name string
	// type of the extra value
	Type string
	// value of the extra value
	Value string
}

// ErrFailedToParseFile is an error returned when the file cannot be parsed.
var ErrFailedToParseFile = fmt.Errorf("failed to parse file")

// ParseAndGenerate parses the file and generates the enum go file for the enum type with failfast mode flag.
func ParseAndGenerate(filename string, failfast bool) error {
	// Set up the parser
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file while generating enum: %w", err)
	}

	packageName := getPackageName(node)

	// Traverse the AST to find type definitions and collect comments
	// Collect comments associated with the type definition
	typeComments := getTypeComments(node)
	enums, iotaType, iotaIdx, nameTPairs := parseEnums(node, typeComments)
	typeLower, plural := getPlural(iotaType)
	enumRep := EnumRepresentation{
		PackageName: packageName,
		Failfast:    failfast,
		TypeInfo: typeInfo{
			Filename:      filename,
			Index:         iotaIdx,
			Name:          iotaType,
			Camel:         camelCase(iotaType),
			Lower:         typeLower,
			Upper:         strings.ToUpper(iotaType),
			Plural:        plural,
			PluralCamel:   camelCase(plural),
			NameTypePairs: nameTPairs,
		},
		Enums: enums,
	}
	// create new file
	// get the p from the filename

	p := path.Dir(filename)
	// path separator
	linuxPathSeparator := "/"
	fullPath := p + linuxPathSeparator + typeLower + "_enums.go"
	f, err := os.Create(fullPath)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	w := io.StringWriter(f)
	defer f.Close()
	writeAll(w, enumRep)
	// format the file
	err = formatFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to format file: %w", err)
	}
	return nil
}

func getPlural(iotaType string) (string, string) {
	l := len(iotaType)
	if l == 0 {
		return "", ""
	}
	lastChar := iotaType[l-1]
	lower := strings.ToLower(iotaType)
	camel := camelCase(iotaType)
	switch lastChar {
	case 'y':
		return lower[:l-1] + "ies", camel[:l-1] + "ies"
	case 'x', 'z', 'h', 'o', 's':
		return lower + "es", camel + "es"
	default:
		return lower + "s", camel + "s"
	}
}

func parseEnums(node *ast.File, typeComments map[string]string) ([]Enum, string, int, []nameTypePair) {
	var (
		enums           []Enum
		iotaName        string
		iotaType        string
		iotaTypeComment string
		iotaIdx         int
		foundConstants  = make(map[string]struct{})
		nameTPairs      = make([]nameTypePair, 0)
	)
	ast.Inspect(node, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.CONST {
			return true
		}
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok && len(valueSpec.Values) == 1 {
				iotaName, iotaType, iotaTypeComment, iotaIdx = iotaInfo(valueSpec, typeComments)
			}
			continue
		}
		if iotaTypeComment != "" {
			nameTPairs = nameTPairsFromComments(iotaTypeComment, nameTPairs)
		}
		if iotaName != "" {
			for i, spec := range decl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					if _, found := foundConstants[name.Name]; !found {
						iotaTypeComment = getTypeComment(valueSpec, typeComments)
						comment := getComment(valueSpec)
						valid := !strings.Contains(comment, "invalid")
						comment, alternate := getAlternateName(comment, name, nameTPairs)
						nameTPairsCopy := copyNameTPairs(nameTPairs, getValues(comment))
						enums = append(enums, Enum{
							Info: info{
								Name:          name.Name,
								Camel:         camelCase(name.Name),
								Lower:         strings.ToLower(name.Name),
								Upper:         strings.ToUpper(name.Name),
								AlternateName: alternate,
								Value:         i,
								Valid:         valid,
							},
							TypeInfo: typeInfo{
								Name:          iotaType,
								Camel:         camelCase(iotaType),
								Lower:         strings.ToLower(iotaType),
								Upper:         strings.ToUpper(iotaType),
								NameTypePairs: nameTPairsCopy,
							},
							Raw: raw{
								Comment:     comment,
								TypeComment: iotaTypeComment,
							},
						})
						foundConstants[name.Name] = struct{}{}
					}
				}
			}
		}
		return true
	})
	return enums, iotaType, iotaIdx, nameTPairs
}

func getTypeComment(valueSpec *ast.ValueSpec, typeComments map[string]string) string {
	if valueSpec.Type != nil {
		constantType := fmt.Sprintf("%s", valueSpec.Type)
		if comment, exists := typeComments[constantType]; exists {
			return comment
		}
	}
	return ""
}

func getPackageName(node *ast.File) string {
	var packageName string
	if node.Name != nil {
		packageName = node.Name.Name
	}
	return packageName
}

func getTypeComments(node *ast.File) map[string]string {
	typeComments := make(map[string]string)
	ast.Inspect(node, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			return true
		}
		for _, spec := range decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if typeSpec.Comment != nil && len(typeSpec.Comment.List) > 0 {
				comment := strings.TrimSpace(typeSpec.Comment.List[0].Text[2:])
				typeComments[typeSpec.Name.Name] = comment
			}
		}
		return true
	})
	return typeComments
}

func getValues(comment string) []string {
	values := strings.Split(comment, ",")
	if len(values) > 1 {
		for i, v := range values {
			values[i] = strings.TrimSpace(v)
		}
	}
	return values
}

func copyNameTPairs(nameTPairs []nameTypePair, values []string) []nameTypePair {
	nameTPairsCopy := make([]nameTypePair, len(nameTPairs))
	copy(nameTPairsCopy, nameTPairs)

	if len(values) == len(nameTPairsCopy) {
		for i, v := range nameTPairsCopy {
			v.Value = values[i]
			nameTPairsCopy[i] = v
		}
	}
	return nameTPairsCopy
}

func getAlternateName(comment string, name *ast.Ident, nameTPairs []nameTypePair) (string, string) {
	// get value between the first space and the first comma
	comment = strings.TrimLeft(comment, " ")
	count := strings.Count(comment, " ")
	switch count {
	case 0:
		if comment == "" {
			return "", name.Name
		}
		if strings.Contains(comment, ",") {
			return comment, name.Name
		}
		if len(nameTPairs) == 1 {
			return comment, name.Name
		}
		if strings.Contains(comment, "invalid") {
			return comment, name.Name
		}
		return comment, comment
	case 1:
		split := strings.Split(comment, " ")
		if len(split) == 2 {
			if strings.Contains(split[0], "invalid") {
				return split[1], split[1]
			}
			return split[1], split[0]
		}
		return comment, name.Name
	}
	return comment, name.Name
}

func getComment(valueSpec *ast.ValueSpec) string {
	var comment string
	if valueSpec.Comment != nil && len(valueSpec.Comment.List) > 0 {
		comment = valueSpec.Comment.List[0].Text
		comment = comment[2:]
	}
	return comment
}

func nameTPairsFromComments(iotaTypeComment string, nameTPairs []nameTypePair) []nameTypePair {
	typeValues := strings.Split(iotaTypeComment, ",")
	for i, v := range typeValues {
		if len(v) == 0 {
			continue
		}
		if v[0] == ' ' {
			v = v[1:]
		}
		o := "["
		c := "]"
		if strings.Contains(v, "(") {
			o = "("
			c = ")"
		}
		if strings.Contains(v, " ") {
			o = " "
			c = " "
		}
		idx := strings.Index(v, o)
		if idx == -1 {
			continue
		}
		name := v[:idx]
		name = strings.TrimSpace(name)

		endIndex := strings.Index(v, c)
		if o == " " {
			endIndex = len(v)
		}
		typeName := v[strings.Index(v, o)+1 : endIndex]
		nameTypePair := nameTypePair{Name: name, Type: typeName, Value: fmt.Sprintf("%d", i)}
		nameTPairs = append(nameTPairs, nameTypePair)
	}

	return nameTPairs
}

func iotaInfo(valueSpec *ast.ValueSpec, typeComments map[string]string) (string, string, string, int) {
	var (
		iotaName, iotaType, iotaTypeComment string
	)
	ident, ok := valueSpec.Values[0].(*ast.Ident)

	iotaIdx := 0
	if ok && ident.Name == "iota" {
		iotaName = valueSpec.Names[0].Name
		if valueSpec.Type != nil {
			iotaType = fmt.Sprintf("%s", valueSpec.Type)
			if comment, exists := typeComments[iotaType]; exists {
				iotaTypeComment = comment
			}
		}
	}
	if !ok {
		if be, ok := valueSpec.Values[0].(*ast.BinaryExpr); ok {
			if x, ok := be.X.(*ast.Ident); ok {
				if x.Name == "iota" {
					iotaName = valueSpec.Names[0].Name
					if valueSpec.Type != nil {
						iotaType = fmt.Sprintf("%s", valueSpec.Type)
						if comment, exists := typeComments[iotaType]; exists {
							iotaTypeComment = comment
						}
					}
				}
			}
			if y, ok := be.Y.(*ast.BasicLit); ok {
				var err error
				iotaIdx, err = strconv.Atoi(y.Value)
				if err != nil {
					iotaIdx = 0
				}
			}
		}
	}
	return iotaName, iotaType, iotaTypeComment, iotaIdx
}

func formatFile(filename string) error {
	f, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	b, err := format.Source(f)
	if err != nil {
		return fmt.Errorf("failed to format file: %w", err)
	}
	err = os.WriteFile(filename, b, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func writeAll(w io.StringWriter, enum EnumRepresentation) {
	writeGeneratedComment(w, enum)
	writePackage(w, enum)
	writeImports(w, enum)
	writeWrapperType(w, enum)
	writeAllMethod(w, enum)
	writeParseMethod(w, enum)
	writeExhaustiveMethod(w, enum)
	writeIsValidMethod(w, enum)
	writeJSONMarshalMethod(w, enum)
	writeJSONUnmarshalMethod(w, enum)
	writeScanMethod(w, enum)
	writeValueMethod(w, enum)
	writeCompileCheck(w, enum)
	writeStringMethod(w, enum)
}

func writeScanMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func (p *" + rep.TypeInfo.Camel + ") Scan(value any) error {\n")
	w.WriteString("\tnewp, err := Parse" + rep.TypeInfo.Camel + "(value)\n")
	w.WriteString("\tif err != nil {\n")
	w.WriteString("\t\treturn err\n")
	w.WriteString("\t}\n")
	w.WriteString("\t*p = newp\n")
	w.WriteString("\treturn nil\n")
	w.WriteString("}\n\n")
}

func writeValueMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func (p " + rep.TypeInfo.Camel + ") Value() (driver.Value, error) {\n")
	w.WriteString("\treturn p.String(), nil\n")
	w.WriteString("}\n\n")
}

func writeGeneratedComment(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("// Code generated by goenums. DO NOT EDIT.\n")
	w.WriteString("// This file was generated by github.com/zarldev/goenums \n")
	w.WriteString("// using the command:\n")
	w.WriteString("// goenums ")
	if rep.Failfast {
		w.WriteString("-f ")
	}
	w.WriteString(rep.TypeInfo.Filename + "\n")
	w.WriteString("\n")
}

func writeStringMethod(w io.StringWriter, rep EnumRepresentation) {
	index, nameConst := generateIndexAndNameRun(rep)
	w.WriteString("const " + nameConst + "\n")
	w.WriteString("var " + index + "\n")
	w.WriteString("func (i " + rep.TypeInfo.Name + ") String() string {\n")
	w.WriteString("\tif i < 0 || i >= " + rep.TypeInfo.Name + "(len(_" + rep.TypeInfo.Lower + "_index)-1) {\n")
	w.WriteString("\t\treturn \"" + rep.TypeInfo.Lower + "(\" + (strconv.FormatInt(int64(i), 10) + \")\")\n")
	w.WriteString("\t}\n")
	w.WriteString("\treturn _" + rep.TypeInfo.Lower + "_name[_" + rep.TypeInfo.Lower + "_index[i]:_" + rep.TypeInfo.Lower + "_index[i+1]]\n")
	w.WriteString("}\n")
}

func generateIndexAndNameRun(rep EnumRepresentation) (string, string) {
	b := new(bytes.Buffer)
	indexes := make([]int, len(rep.Enums))
	for i := range rep.Enums {
		b.WriteString(rep.Enums[i].Info.AlternateName)
		indexes[i] = b.Len()
	}
	nameConst := fmt.Sprintf("_%s_name = %q\n", rep.TypeInfo.Lower, b.String())
	b.Reset()
	fmt.Fprintf(b, " _%s_index = [...]uint16{0", rep.TypeInfo.Lower)
	for range rep.TypeInfo.Index {
		fmt.Fprintf(b, ", %d", 0)
	}
	for _, i := range indexes {
		if i > 0 {
			fmt.Fprintf(b, ", ")
		}
		fmt.Fprintf(b, "%d", i)
	}
	fmt.Fprintf(b, "}\n")
	return b.String(), nameConst
}

func writeCompileCheck(w io.StringWriter, rep EnumRepresentation) {
	// Generate code that will fail if the constants change value.
	w.WriteString("func _() {\n")
	w.WriteString("\t// An \"invalid array index\" compiler error signifies that the constant values have changed.\n")
	w.WriteString("\t// Re-run the goenums command to generate them again.\n")
	w.WriteString("\t// Does not identify newly added constant values unless order changes\n")
	w.WriteString("\tvar x [1]struct{}\n")
	for _, v := range rep.Enums {
		w.WriteString(fmt.Sprintf("\t_ = x[%s - %d]\n", v.Info.Name, v.Info.Value+rep.TypeInfo.Index))
	}
	w.WriteString("}\n")
}

func writeJSONMarshalMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func (p " + rep.TypeInfo.Camel + ") MarshalJSON() ([]byte, error) {\n")
	w.WriteString("\treturn []byte(`\"`+p.String() + `\"`), nil\n")
	w.WriteString("}\n\n")
}

func writeJSONUnmarshalMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func (p *" + rep.TypeInfo.Camel + ") UnmarshalJSON(b []byte) error {\n")
	w.WriteString("b = bytes.Trim(bytes.Trim(b, `\"`), ` `)\n")
	w.WriteString("\tnewp, err := Parse" + rep.TypeInfo.Camel + "(b)\n")
	w.WriteString("\tif err != nil {\n")
	w.WriteString("\t\treturn err\n")
	w.WriteString("\t}\n")
	w.WriteString("\t*p = newp\n")
	w.WriteString("\treturn nil\n")
	w.WriteString("}\n\n")
}

func writeIsValidMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("var valid" + rep.TypeInfo.PluralCamel + " = map[" + rep.TypeInfo.Camel + "]bool{\n")
	for _, info := range rep.Enums {
		if info.Info.Valid {
			w.WriteString("\t" + rep.TypeInfo.PluralCamel + "." + info.Info.Upper + ": true,\n")
		}
	}
	w.WriteString("}\n\n")
	w.WriteString("func (p " + rep.TypeInfo.Camel + ") IsValid() bool {\n")
	w.WriteString("\treturn valid" + rep.TypeInfo.PluralCamel + "[p]\n")
	w.WriteString("}\n\n")
}

func writeExhaustiveMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func Exhaustive" + rep.TypeInfo.Camel + "s(f func(" + rep.TypeInfo.Camel + ")) {\n")
	w.WriteString("\tfor _, p := range " + rep.TypeInfo.PluralCamel + ".All() {\n")
	w.WriteString("\t\tf(p)\n")
	w.WriteString("\t}\n")
	w.WriteString("}\n\n")
}

func writePackage(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("package " + rep.PackageName + "\n\n")
}

func writeImports(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("import (\n")
	w.WriteString("\t\"fmt\"\n")
	w.WriteString("\t\"strconv\"\n")
	w.WriteString("\t\"bytes\"\n")
	w.WriteString("\t\"database/sql/driver\"\n")
	for _, pair := range rep.TypeInfo.NameTypePairs {
		if strings.Contains(pair.Type, ".") {
			pkg := strings.Split(pair.Type, ".")[0]
			w.WriteString("\t\"" + pkg + "\"\n")
		}
	}
	w.WriteString(")\n\n")
}

func writeWrapperType(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("type " + rep.TypeInfo.Camel + " struct {\n")
	w.WriteString(rep.TypeInfo.Name + "\n")
	for _, pair := range rep.TypeInfo.NameTypePairs {
		w.WriteString("\t" + pair.Name + " " + pair.Type + "\n")
	}
	w.WriteString("}\n\n")
	w.WriteString("type " + rep.TypeInfo.Lower + "Container struct {\n")
	for _, info := range rep.Enums {
		w.WriteString("\t" + info.Info.Upper + " " + info.TypeInfo.Camel + "\n")
	}
	w.WriteString("}\n\n")
	w.WriteString("var " + rep.TypeInfo.PluralCamel + " = " + rep.TypeInfo.Lower + "Container{\n")
	for _, info := range rep.Enums {
		if info.Info.Valid {
			w.WriteString("\t" + info.Info.Upper + ": " + info.TypeInfo.Camel + "{ \n\t" + info.TypeInfo.Name + ":" + info.Info.Name + ",\n")
			for i := range info.TypeInfo.NameTypePairs {
				w.WriteString(info.TypeInfo.NameTypePairs[i].Name + ": " + info.TypeInfo.NameTypePairs[i].Value + ",\n")
			}
			w.WriteString("},\n")
		}
	}
	w.WriteString("}\n\n")
}

func writeAllMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func (c " + rep.TypeInfo.Lower + "Container) All() []" + rep.TypeInfo.Camel + " {\n")
	w.WriteString("\treturn []" + rep.TypeInfo.Camel + "{\n")
	for _, info := range rep.Enums {
		if info.Info.Valid {
			w.WriteString("\t\tc." + info.Info.Upper + ",\n")
		}
	}
	w.WriteString("\t}\n")
	w.WriteString("}\n\n")
}

func setupInvalidTypeMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("var invalid" + rep.TypeInfo.Camel + " = " + rep.TypeInfo.Camel + "{}\n\n")
}
func writeParseMethod(w io.StringWriter, rep EnumRepresentation) {
	setupInvalidTypeMethod(w, rep)
	w.WriteString("func Parse" + rep.TypeInfo.Camel + "(a any) (" + rep.TypeInfo.Camel + ", error) {\n")
	w.WriteString("\tres := invalid" + rep.TypeInfo.Camel + "\n")
	w.WriteString("\tswitch v := a.(type) {\n")
	w.WriteString("\tcase " + rep.TypeInfo.Camel + ":\n")
	w.WriteString("\t\treturn v, nil\n")
	w.WriteString("\tcase []byte:\n")
	w.WriteString("\t\tres = stringTo" + rep.TypeInfo.Camel + "(string(v))\n")
	w.WriteString("\tcase string:\n")
	w.WriteString("\t\tres = stringTo" + rep.TypeInfo.Camel + "(v)\n")
	w.WriteString("\tcase fmt.Stringer:\n")
	w.WriteString("\t\tres = stringTo" + rep.TypeInfo.Camel + "(v.String())\n")
	w.WriteString("\tcase int:\n")
	w.WriteString("\t\tres = intTo" + rep.TypeInfo.Camel + "(v)\n")
	w.WriteString("\tcase int64:\n")
	w.WriteString("\t\tres = intTo" + rep.TypeInfo.Camel + "(int(v))\n")
	w.WriteString("\tcase int32:\n")
	w.WriteString("\t\tres = intTo" + rep.TypeInfo.Camel + "(int(v))\n")
	w.WriteString("\t}\n")
	if rep.Failfast {
		w.WriteString("\tif res == invalid" + rep.TypeInfo.Camel + " {\n")
		w.WriteString("\t\treturn res, fmt.Errorf(\"failed to parse %v\", a)\n")
		w.WriteString("\t}\n")
	}
	w.WriteString("\treturn res, nil\n")
	w.WriteString("}\n\n")
	setupStringToTypeMethod(w, rep)
	setupIntToTypeMethod(w, rep)
}

func setupIntToTypeMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func intTo" + rep.TypeInfo.Camel + "(i int) " + rep.TypeInfo.Camel + " {\n")
	w.WriteString("\tif i < 0 || i >= len(" + rep.TypeInfo.PluralCamel + " .All()) {\n")
	w.WriteString("\t\treturn invalid" + rep.TypeInfo.Camel + "\n")
	w.WriteString("\t}\n")
	w.WriteString("\treturn " + rep.TypeInfo.PluralCamel + " .All()[i]\n")
	w.WriteString("}\n\n")
}

func setupStringToTypeMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func stringTo" + rep.TypeInfo.Camel + "(s string) " + rep.TypeInfo.Camel + " {\n")
	// w.WriteString("\tlwr := strings.ToLower(s)\n")
	w.WriteString("\tswitch s {\n")
	for _, info := range rep.Enums {
		w.WriteString("\tcase \"" + info.Info.AlternateName + "\":\n")
		w.WriteString("\t\treturn " + rep.TypeInfo.PluralCamel + "." + info.Info.Upper + "\n")
	}
	w.WriteString("\t}\n")
	w.WriteString("\treturn invalid" + rep.TypeInfo.Camel + "\n")
	w.WriteString("}\n\n")
}
