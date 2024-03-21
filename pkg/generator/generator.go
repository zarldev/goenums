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

func Write(w io.Writer, rep EnumRepresentation) {
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
	w := io.Writer(f)
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
	f.Close()
	// format the file
	err = formatFile(enumRep.TypeInfo.Lower + "_enum.go")
	if err != nil {
		return fmt.Errorf("failed to format file: %w", err)
	}
	return nil
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

func setupGeneratedComment(w io.Writer) {
	w.Write([]byte("// Code generated by goenums. DO NOT EDIT.\n"))
	w.Write([]byte("// This file was generated by github.com/zarldev/goenums/cmd/goenums \n"))
	w.Write([]byte("// using the command:\n"))
	w.Write([]byte("// goenums filename.go\n"))
	w.Write([]byte("\n"))
}
func setupStringMethod(w io.Writer, rep EnumRepresentation) {
	index, nameConst := generateIndexAndNameRun(rep)
	w.Write([]byte("const " + nameConst + "\n"))
	w.Write([]byte("var " + index + "\n"))
	w.Write([]byte("func (i " + rep.TypeInfo.Lower + ") String() string {\n"))
	w.Write([]byte("\tif i < 0 || i >= " + rep.TypeInfo.Lower + "(len(_" + rep.TypeInfo.Lower + "_index)-1) {\n"))
	w.Write([]byte("\t\treturn \"" + rep.TypeInfo.Lower + "(\" + (strconv.FormatInt(int64(i), 10) + \")\")\n"))
	w.Write([]byte("\t}\n"))
	w.Write([]byte("\treturn _" + rep.TypeInfo.Lower + "_name[_" + rep.TypeInfo.Lower + "_index[i]:_" + rep.TypeInfo.Lower + "_index[i+1]]\n"))
	w.Write([]byte("}\n"))
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

func setupCompileCheck(w io.Writer, rep EnumRepresentation) {
	// Generate code that will fail if the constants change value.
	w.Write([]byte("func _() {\n"))
	w.Write([]byte("\t// An \"invalid array index\" compiler error signifies that the constant values have changed.\n"))
	w.Write([]byte("\t// Re-run the goenums command to generate them again.\n"))
	w.Write([]byte("\t// Does not identify newly added constant values unless order changes\n"))
	w.Write([]byte("\tvar x [1]struct{}\n"))
	for _, v := range rep.Enums {
		w.Write([]byte(fmt.Sprintf("\t_ = x[%s - %d]\n", v.Info.Name, v.Info.Value)))
	}
	w.Write([]byte("}\n"))
}

func setupJSONMarshalMethod(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("func (p " + rep.TypeInfo.Camel + ") MarshalJSON() ([]byte, error) {\n"))
	w.Write([]byte("\treturn []byte(`\"`+p.String() + `\"`), nil\n"))
	w.Write([]byte("}\n\n"))
}

func setupJSONUnmarshalMethod(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("func (p *" + rep.TypeInfo.Camel + ") UnmarshalJSON(b []byte) error {\n"))
	w.Write([]byte("b = bytes.Trim(bytes.Trim(b, `\"`), ` `)\n"))
	w.Write([]byte("\t*p = Parse" + rep.TypeInfo.Camel + "(string(b))\n"))
	w.Write([]byte("\treturn nil\n"))
	w.Write([]byte("}\n\n"))
}

func setupIsValidMethod(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("var valid" + rep.TypeInfo.PluralCamel + " = map[" + rep.TypeInfo.Camel + "]bool{\n"))
	for _, info := range rep.Enums {
		if info.Info.Valid {
			w.Write([]byte("\t" + rep.TypeInfo.PluralCamel + "." + info.Info.Upper + ": true,\n"))
		}
	}
	w.Write([]byte("}\n\n"))
	w.Write([]byte("func (p " + rep.TypeInfo.Camel + ") IsValid() bool {\n"))
	w.Write([]byte("\treturn valid" + rep.TypeInfo.PluralCamel + "[p]\n"))
	w.Write([]byte("}\n\n"))
}

func setupExhaustiveMethod(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("func Exhaustive" + rep.TypeInfo.Camel + "s(f func(" + rep.TypeInfo.Camel + ")) {\n"))
	w.Write([]byte("\tfor _, p := range " + rep.TypeInfo.PluralCamel + ".All() {\n"))
	w.Write([]byte("\t\tf(p)\n"))
	w.Write([]byte("\t}\n"))
	w.Write([]byte("}\n\n"))
}

func setupPackage(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("package " + rep.PackageName + "\n\n"))
}

func setupImports(w io.Writer) {
	w.Write([]byte("import (\n"))
	w.Write([]byte("\t\"fmt\"\n"))
	w.Write([]byte("\t\"strings\"\n"))
	w.Write([]byte("\t\"strconv\"\n"))
	w.Write([]byte("\t\"bytes\"\n"))
	w.Write([]byte(")\n\n"))
}

func setupWrapperType(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("type " + rep.TypeInfo.Camel + " struct {\n"))
	w.Write([]byte(rep.TypeInfo.Name + "\n"))
	for _, pair := range rep.TypeInfo.NameTypePairs {
		w.Write([]byte("\t" + pair.Name + " " + pair.Type + "\n"))
	}
	w.Write([]byte("}\n\n"))
	w.Write([]byte("type " + rep.TypeInfo.Lower + "Container struct {\n"))
	for _, info := range rep.Enums {
		w.Write([]byte("\t" + info.Info.Upper + " " + info.TypeInfo.Camel + "\n"))
	}
	w.Write([]byte("}\n\n"))
	w.Write([]byte("var " + rep.TypeInfo.PluralCamel + " = " + rep.TypeInfo.Lower + "Container{\n"))
	for _, info := range rep.Enums {
		if info.Info.Valid {
			w.Write([]byte("\t" + info.Info.Upper + ": " + info.TypeInfo.Camel + "{ \n\t" + info.TypeInfo.Name + ":" + info.Info.Name + ",\n"))
			for i := range info.TypeInfo.NameTypePairs {
				w.Write([]byte(info.TypeInfo.NameTypePairs[i].Name + ": " + info.TypeInfo.NameTypePairs[i].Value + ",\n"))
			}
			w.Write([]byte("},\n"))
		}
	}
	w.Write([]byte("}\n\n"))
}

func setupAllMethod(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("func (c " + rep.TypeInfo.Lower + "Container) All() []" + rep.TypeInfo.Camel + " {\n"))
	w.Write([]byte("\treturn []" + rep.TypeInfo.Camel + "{\n"))
	for _, info := range rep.Enums {
		if info.Info.Valid {
			w.Write([]byte("\t\tc." + info.Info.Upper + ",\n"))
		}
	}
	w.Write([]byte("\t}\n"))
	w.Write([]byte("}\n\n"))
}

func setupInvalidTypeMethod(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("var invalid" + rep.TypeInfo.Camel + " = " + rep.TypeInfo.Camel + "{}\n\n"))
}
func setupParseMethod(w io.Writer, rep EnumRepresentation) {
	setupInvalidTypeMethod(w, rep)
	w.Write([]byte("func Parse" + rep.TypeInfo.Camel + "(a any) " + rep.TypeInfo.Camel + " {\n"))
	w.Write([]byte("\tswitch v := a.(type) {\n"))
	w.Write([]byte("\tcase " + rep.TypeInfo.Camel + ":\n"))
	w.Write([]byte("\t\treturn v\n"))
	w.Write([]byte("\tcase string:\n"))
	w.Write([]byte("\t\treturn stringTo" + rep.TypeInfo.Camel + "(v)\n"))
	w.Write([]byte("\tcase fmt.Stringer:\n"))
	w.Write([]byte("\t\treturn stringTo" + rep.TypeInfo.Camel + "(v.String())\n"))
	w.Write([]byte("\tcase int:\n"))
	w.Write([]byte("\t\treturn intTo" + rep.TypeInfo.Camel + "(v)\n"))
	w.Write([]byte("\tcase int64:\n"))
	w.Write([]byte("\t\treturn intTo" + rep.TypeInfo.Camel + "(int(v))\n"))
	w.Write([]byte("\tcase int32:\n"))
	w.Write([]byte("\t\treturn intTo" + rep.TypeInfo.Camel + "(int(v))\n"))
	w.Write([]byte("\t}\n"))
	w.Write([]byte("\treturn invalid" + rep.TypeInfo.Camel + "\n"))
	w.Write([]byte("}\n\n"))
	setupStringToTypeMethod(w, rep)
	setupIntToTypeMethod(w, rep)

}

func setupIntToTypeMethod(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("func intTo" + rep.TypeInfo.Camel + "(i int) " + rep.TypeInfo.Camel + " {\n"))
	w.Write([]byte("\tif i < 0 || i >= len(" + rep.TypeInfo.PluralCamel + " .All()) {\n"))
	w.Write([]byte("\t\treturn invalid" + rep.TypeInfo.Camel + "\n"))
	w.Write([]byte("\t}\n"))
	w.Write([]byte("\treturn " + rep.TypeInfo.PluralCamel + " .All()[i]\n"))
	w.Write([]byte("}\n\n"))
}

func setupStringToTypeMethod(w io.Writer, rep EnumRepresentation) {
	w.Write([]byte("func stringTo" + rep.TypeInfo.Camel + "(s string) " + rep.TypeInfo.Camel + " {\n"))
	w.Write([]byte("\tlwr := strings.ToLower(s)\n"))
	w.Write([]byte("\tswitch lwr {\n"))
	for _, info := range rep.Enums {
		w.Write([]byte("\tcase \"" + info.Info.Lower + "\":\n"))
		w.Write([]byte("\t\treturn " + rep.TypeInfo.PluralCamel + "." + info.Info.Upper + "\n"))
	}
	w.Write([]byte("\t}\n"))
	w.Write([]byte("\treturn invalid" + rep.TypeInfo.Camel + "\n"))
	w.Write([]byte("}\n\n"))
}
