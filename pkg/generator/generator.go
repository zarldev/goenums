package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	// camelCase is a Caser for turning strings into camelCase
	camelCase = cases.Title(language.English)
)

type Generator struct {
}

func NewGenerator() *Generator {
	return &Generator{}
}

func Write(w io.StringWriter, rep EnumRepresentation) {
	setupPackage(w, rep)
	setupImports(w)
	setupWrapperType(w, rep)
	setupAllMethod(w, rep)
	setupParseMethod(w, rep)
	setupExhaustiveMethod(w, rep)
	setupIsValidMethod(w, rep)
	setupJSONMarshalMethod(w, rep)
	setupJSONUnmarshalMethod(w, rep)
	setupCompileCheck(w, rep)
	setupStringMethod(w, rep)
}

// EnumRepresentation is a struct to store the information to be used in writing the enum to a file
type EnumRepresentation struct {
	PackageName string
	TypeInfo    typeInfo
	Enums       []Enum
}

// Enum is a struct to store the information for each enum to be written
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
	// type name for the enum in different cases
	Name        string
	Camel       string
	Lower       string
	Upper       string
	Plural      string
	PluralCamel string
	// name type pairs for the enum not using iota
	NameTypePairs []NameTypePair
}

// NameTypePair is a struct to store the name and type of the extra values for the enum
type NameTypePair struct {
	// name of the extra value
	Name string
	// type of the extra value
	Type string
	// value of the extra value
	Value string
}

var ErrFailedToParseFile = fmt.Errorf("failed to parse file")

func ParseAndGenerate(filename string) error {
	// Set up the parser
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file while generating enum: %w", err)
	}
	// Get package name
	var packageName string
	if node.Name != nil {
		packageName = node.Name.Name
	}
	// Map to store type comments
	typeComments := make(map[string]string)
	// Traverse the AST to find type definitions and collect comments
	ast.Inspect(node, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			// Not a type declaration, continue traversal
			return true
		}
		// Check each type specification in the declaration
		for _, spec := range decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			// Collect comments associated with the type definition
			if typeSpec.Comment != nil && len(typeSpec.Comment.List) > 0 {
				comment := strings.TrimSpace(typeSpec.Comment.List[0].Text[2:])
				typeComments[typeSpec.Name.Name] = comment
			}
		}
		return true
	})
	// Slice to store constant info
	var enums []Enum
	// iota info
	// iotaName is the name of the constant using iota
	var iotaName string
	// iotaType is the type of the constant using iota
	var iotaType string
	// iotaTypeComment is the comment associated with the type of the constant using iota which is used to associate other type value pairs
	var iotaTypeComment string
	// Temporary map to store found constant names
	foundConstants := make(map[string]struct{})
	nameTPairs := make([]NameTypePair, 0)
	// Traverse the AST to find constant declarations
	ast.Inspect(node, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.CONST {
			// Not a constant declaration, continue traversal
			return true
		}
		// Check each constant in the declaration
		for _, spec := range decl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			// Check if this constant is using iota
			if len(valueSpec.Values) == 1 {
				ident, ok := valueSpec.Values[0].(*ast.Ident)
				if ok && ident.Name == "iota" {
					// Constant using iota found
					iotaName = valueSpec.Names[0].Name
					if valueSpec.Type != nil {
						iotaType = fmt.Sprintf("%s", valueSpec.Type)
						// Associate comments from the type definition
						if comment, exists := typeComments[iotaType]; exists {
							iotaTypeComment = comment
						}
					}
				}
			}
		}
		if iotaTypeComment != "" {
			typeValues := strings.Split(iotaTypeComment, ",")
			for i, v := range typeValues {
				typeValues[i] = strings.TrimSpace(v)
				idx := strings.Index(v, "[")
				if idx == -1 {
					continue
				}
				name := v[:idx]
				name = strings.TrimSpace(name)
				typeName := v[strings.Index(v, "[")+1 : strings.Index(v, "]")]
				nameTypePair := NameTypePair{Name: name, Type: typeName, Value: fmt.Sprintf("%d", i)}
				nameTPairs = append(nameTPairs, nameTypePair)
			}
		}

		// For constants not using iota in the same block as iota, set their values and types accordingly
		if iotaName != "" {
			for i, spec := range decl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					if _, found := foundConstants[name.Name]; !found {
						var constantType string
						if valueSpec.Type != nil {
							constantType = fmt.Sprintf("%s", valueSpec.Type)
							// Associate comments from the type definition
							if comment, exists := typeComments[constantType]; exists {
								iotaTypeComment = comment
							}
						}
						var comment string
						if valueSpec.Comment != nil && len(valueSpec.Comment.List) > 0 {
							comment = strings.TrimSpace(valueSpec.Comment.List[0].Text)
							// remove the '//' from the comment
							comment = comment[2:]
							// trim the comment
							comment = strings.TrimSpace(comment)
						}
						valid := !strings.Contains(comment, "invalid")
						// count the number of spaces before the 1st value
						// if there are no spaces, then it is the first value
						count := strings.Count(comment, " ")
						alternate := name.Name
						if count > 0 {
							nameComm := strings.Split(comment, " ")
							alternate = nameComm[0]
							comment = nameComm[1]
						}
						comment = strings.TrimSpace(comment)
						values := strings.Split(comment, ",")
						if len(values) > 1 {
							for i, v := range values {
								values[i] = strings.TrimSpace(v)
							}
						}
						nameTPairsCopy := make([]NameTypePair, len(nameTPairs))
						copy(nameTPairsCopy, nameTPairs)

						// set name type pair values for the type not using iota
						if len(values) == len(nameTPairsCopy) {
							for i, v := range nameTPairsCopy {
								v.Value = values[i]
								nameTPairsCopy[i] = v
							}
						}

						enums = append(enums, Enum{
							Info: info{
								Name:          name.Name,
								Camel:         camelCase.String(name.Name),
								Lower:         strings.ToLower(name.Name),
								Upper:         strings.ToUpper(name.Name),
								AlternateName: alternate,
								Value:         i,
								Valid:         valid,
							},
							TypeInfo: typeInfo{
								Name:          iotaType,
								Camel:         camelCase.String(iotaType),
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
	typeLower := strings.ToLower(iotaType)
	pluralSuffix := "s"
	if typeLower[len(typeLower)-1] == 's' {
		pluralSuffix = "es"
	}
	plural := typeLower + pluralSuffix
	enumRep := EnumRepresentation{
		PackageName: packageName,
		TypeInfo: typeInfo{
			Name:          iotaType,
			Camel:         camelCase.String(iotaType),
			Lower:         typeLower,
			Upper:         strings.ToUpper(iotaType),
			Plural:        plural,
			PluralCamel:   camelCase.String(plural),
			NameTypePairs: nameTPairs,
		},
		Enums: enums,
	}

	// create new file
	f, err := os.Create(enumRep.TypeInfo.Lower + "_enum.go")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer f.Close()
	w := io.StringWriter(f)
	setupGeneratedComment(w)
	setupPackage(w, enumRep)
	setupImports(w)
	setupWrapperType(w, enumRep)
	setupAllMethod(w, enumRep)
	setupParseMethod(w, enumRep)
	setupExhaustiveMethod(w, enumRep)
	setupIsValidMethod(w, enumRep)
	setupJSONMarshalMethod(w, enumRep)
	setupJSONUnmarshalMethod(w, enumRep)
	setupCompileCheck(w, enumRep)
	setupStringMethod(w, enumRep)
	formatGeneratedFile(f)

	return nil
}

func setupGeneratedComment(w io.StringWriter) {
	w.WriteString("// Code generated by goenums. DO NOT EDIT.\n")
	w.WriteString("// This file was generated by github.com/zarldev/goenums/cmd/goenums \n")
	w.WriteString("// using the command:\n")
	w.WriteString("// goenums filename.go\n")
	w.WriteString("\n")
}
func setupStringMethod(w io.StringWriter, rep EnumRepresentation) {
	index, nameConst := generateIndexAndNameRun(rep)
	w.WriteString("const " + nameConst + "\n")
	w.WriteString("var " + index + "\n")
	w.WriteString("func (i " + rep.TypeInfo.Lower + ") String() string {\n")
	w.WriteString("\tif i < 0 || i >= " + rep.TypeInfo.Lower + "(len(_" + rep.TypeInfo.Lower + "_index)-1) {\n")
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
	for _, i := range indexes {
		if i > 0 {
			fmt.Fprintf(b, ", ")
		}
		fmt.Fprintf(b, "%d", i)
	}
	fmt.Fprintf(b, "}\n")
	return b.String(), nameConst
}

func setupCompileCheck(w io.StringWriter, rep EnumRepresentation) {
	// Generate code that will fail if the constants change value.
	w.WriteString("func _() {\n")
	w.WriteString("\t// An \"invalid array index\" compiler error signifies that the constant values have changed.\n")
	w.WriteString("\t// Re-run the goenums command to generate them again.\n")
	w.WriteString("\t// Does not identify newly added constant values unless order changes\n")
	w.WriteString("\tvar x [1]struct{}\n")
	for _, v := range rep.Enums {
		w.WriteString(fmt.Sprintf("\t_ = x[%s - %d]\n", v.Info.Name, v.Info.Value))
	}
	w.WriteString("}\n")
}

func setupJSONMarshalMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func (p " + rep.TypeInfo.Camel + ") MarshalJSON() ([]byte, error) {\n")
	w.WriteString("\treturn []byte(`\"`+p.String() + `\"`), nil\n")
	w.WriteString("}\n\n")
}

func setupJSONUnmarshalMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func (p *" + rep.TypeInfo.Camel + ") UnmarshalJSON(b []byte) error {\n")
	w.WriteString("b = bytes.Trim(bytes.Trim(b, `\"`), ` `)\n")
	w.WriteString("\t*p = Parse" + rep.TypeInfo.Camel + "(string(b))\n")
	w.WriteString("\treturn nil\n")
	w.WriteString("}\n\n")
}

func setupIsValidMethod(w io.StringWriter, rep EnumRepresentation) {
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

func setupExhaustiveMethod(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("func Exhaustive" + rep.TypeInfo.Camel + "s(f func(" + rep.TypeInfo.Camel + ")) {\n")
	w.WriteString("\tfor _, p := range " + rep.TypeInfo.PluralCamel + ".All() {\n")
	w.WriteString("\t\tf(p)\n")
	w.WriteString("\t}\n")
	w.WriteString("}\n\n")
}

func setupPackage(w io.StringWriter, rep EnumRepresentation) {
	w.WriteString("package " + rep.PackageName + "\n\n")
}

func setupImports(w io.StringWriter) {
	w.WriteString("import (\n")
	w.WriteString("\t\"fmt\"\n")
	w.WriteString("\t\"strings\"\n")
	w.WriteString("\t\"strconv\"\n")
	w.WriteString("\t\"bytes\"\n")
	w.WriteString(")\n\n")
}

func setupWrapperType(w io.StringWriter, rep EnumRepresentation) {
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
		if info.Info.Valid && len(info.TypeInfo.NameTypePairs) > 0 {
			w.WriteString("\t" + info.Info.Upper + ": " + info.TypeInfo.Camel + "{ \n\t" + info.TypeInfo.Name + ":" + info.Info.Name + ",\n")
			for i := range info.TypeInfo.NameTypePairs {
				w.WriteString(info.TypeInfo.NameTypePairs[i].Name + ": " + info.TypeInfo.NameTypePairs[i].Value + ",\n")
			}
			w.WriteString("},\n")
		}
	}
	w.WriteString("}\n\n")
}

func setupAllMethod(w io.StringWriter, rep EnumRepresentation) {
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
func setupParseMethod(w io.StringWriter, rep EnumRepresentation) {
	setupInvalidTypeMethod(w, rep)
	w.WriteString("func Parse" + rep.TypeInfo.Camel + "(a any) " + rep.TypeInfo.Camel + " {\n")
	w.WriteString("\tswitch v := a.(type) {\n")
	w.WriteString("\tcase " + rep.TypeInfo.Camel + ":\n")
	w.WriteString("\t\treturn v\n")
	w.WriteString("\tcase string:\n")
	w.WriteString("\t\treturn stringTo" + rep.TypeInfo.Camel + "(v)\n")
	w.WriteString("\tcase fmt.Stringer:\n")
	w.WriteString("\t\treturn stringTo" + rep.TypeInfo.Camel + "(v.String())\n")
	w.WriteString("\tcase int:\n")
	w.WriteString("\t\treturn intTo" + rep.TypeInfo.Camel + "(v)\n")
	w.WriteString("\tcase int64:\n")
	w.WriteString("\t\treturn intTo" + rep.TypeInfo.Camel + "(int(v))\n")
	w.WriteString("\tcase int32:\n")
	w.WriteString("\t\treturn intTo" + rep.TypeInfo.Camel + "(int(v))\n")
	w.WriteString("\t}\n")
	w.WriteString("\treturn invalid" + rep.TypeInfo.Camel + "\n")
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
	w.WriteString("\tlwr := strings.ToLower(s)\n")
	w.WriteString("\tswitch lwr {\n")
	for _, info := range rep.Enums {
		w.WriteString("\tcase \"" + info.Info.Lower + "\":\n")
		w.WriteString("\t\treturn " + rep.TypeInfo.PluralCamel + "." + info.Info.Upper + "\n")
	}
	w.WriteString("\t}\n")
	w.WriteString("\treturn invalid" + rep.TypeInfo.Camel + "\n")
	w.WriteString("}\n\n")
}

func formatGeneratedFile(f *os.File) {
	cmd := exec.Command("gofmt", "-w", f.Name())
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to format generated file using gofmt: %v", err)
	}

	if string(output) != "" {
		fmt.Println(output)
	}
}
