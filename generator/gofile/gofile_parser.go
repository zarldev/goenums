// Package gofile provides Go-specific parsing and generation capabilities for enums.
// This parser analyzes Go source files to extract enum-like constant declarations and
// transforms them into language-agnostic enum representations.
package gofile

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"strconv"
	"time"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/internal/version"
	"github.com/zarldev/goenums/source"
	"github.com/zarldev/goenums/strings"
)

// Compile-time check that Parser implements enum.Parser
var _ enum.Parser = (*Parser)(nil)

var (
	// ErrParseGoFile indicates an error occurred while parsing the source file.
	ErrParseGoFile = errors.New("failed to parse Go file")
	// ErrReadSource indicates an error occurred while reading the source file.
	ErrReadGoFile = errors.New("failed to read Go file")
)

// Parser implements the enum.Parser interface for Go source files.
// It analyzes Go constant declarations to identify and extract enum patterns,
// translating them into a standardized representation model.
type Parser struct {
	Configuration config.Configuration
	source        enum.Source
}

// ParserOption is a function that configures a Parser.
type ParserOption func(*Parser)

// WithSource sets the source for the parser.
func WithSource(source enum.Source) ParserOption {
	return func(p *Parser) {
		p.source = source
	}
}

// WithParserConfiguration sets the configuration for the parser.
func WithParserConfiguration(configuration config.Configuration) ParserOption {
	return func(p *Parser) {
		p.Configuration = configuration
	}
}

// NewParser creates a new Go file parser with the specified configuration and source.
// The parser will analyze the source according to the configuration settings.
func NewParser(opts ...ParserOption) *Parser {
	p := Parser{
		Configuration: config.Configuration{},
		source:        source.FromFile(""),
	}
	for _, opt := range opts {
		opt(&p)
	}
	return &p
}

// Parse analyzes Go source code to identify and extract enum-like constant declarations.
// It returns a slice of enum representations or an error if parsing fails.
// The implementation uses Go's standard AST parsing to analyze the source code structure.
func (p *Parser) Parse(ctx context.Context) ([]enum.EnumIota, error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Default().Error("unexpected panic in parser",
				"version", version.CURRENT,
				"build", version.BUILD,
				"commit", version.COMMIT,
				"recovered", true,
				"error", fmt.Sprintf("%v", r),
				"file", p.source.Filename())
		}
	}()
	return p.doParse(ctx)
}

const (
	iotaIdentifier = "iota"
)

func (p *Parser) doParse(ctx context.Context) ([]enum.EnumIota, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	content, err := p.source.Content()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadGoFile, err)
	}
	slog.Default().DebugContext(ctx, "parsing source content")
	filename := p.source.Filename()
	fset := token.NewFileSet()
	slog.Default().DebugContext(ctx, "parsing file", "filename", filename)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	node, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseGoFile, err)
	}
	slog.Default().DebugContext(ctx, "collecting all enum representations")

	enumIotas := p.getEnumIotas(node)
	slog.Default().DebugContext(ctx, "enum iota", "count", len(enumIotas), "enumIota", enumIotas)
	slog.Default().DebugContext(ctx, "collecting all enum representations", "filename", filename)

	for i, enumIota := range enumIotas {
		slog.Default().DebugContext(ctx, "enum iota", "enumIota", enumIota)
		enums := p.getEnums(node, &enumIota)
		slog.Default().DebugContext(ctx, "enums", "count", len(enums), "enums", enums)
		enumIota.Enums = enums
		enumIotas[i] = enumIota
	}
	return enumIotas, nil
}

// func (p *Parser) doParse(ctx context.Context) ([]enum.Representation, error) {
// 	if ctx.Err() != nil {
// 		return nil, ctx.Err()
// 	}
// 	content, err := p.source.Content()
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: %w", ErrReadGoFile, err)
// 	}
// 	slog.Default().DebugContext(ctx, "parsing source content")
// 	filename := p.source.Filename()
// 	fset := token.NewFileSet()
// 	slog.Default().DebugContext(ctx, "parsing file", "filename", filename)
// 	if ctx.Err() != nil {
// 		return nil, ctx.Err()
// 	}
// 	node, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: %w", ErrParseGoFile, err)
// 	}
// 	typeComments := p.getTypeComments(node)
// 	reps := p.collectRepresentations(node, filename, typeComments,
// 		p.Configuration)
// 	slog.Default().DebugContext(ctx, "collected all enum representations", "count", len(reps))
// 	return reps, nil
// }

func (p *Parser) getEnums(node *ast.File, enumIota *enum.EnumIota) []enum.Enum {
	var enums []enum.Enum

	for _, decl := range node.Decls {
		switch t := decl.(type) {
		case *ast.GenDecl:
			idx := 0
			for _, spec := range t.Specs {
				switch vs := spec.(type) {
				case *ast.ValueSpec:
					e := p.getEnum(vs, &idx, enumIota)
					if e == nil {
						continue
					}
					enums = append(enums, *e)
					slog.Default().Debug("enum", "enum", e)
				}
			}
		}
	}
	return enums
}

func (p *Parser) getEnum(vs *ast.ValueSpec, idx *int, enumIota *enum.EnumIota) *enum.Enum {
	if len(vs.Names) == 0 {
		return nil
	}
	if vs.Type != nil {
		switch t := vs.Type.(type) {
		case *ast.Ident:
			if t.Name != enumIota.Type {
				return nil
			}
		}
	}
	name := vs.Names[0].Name
	if name == "_" {
		*idx++
		return nil
	}
	enum := enum.Enum{
		Name: vs.Names[0].Name,
	}
	for _, v := range vs.Values {
		switch t := v.(type) {
		case *ast.BinaryExpr:
			x, ok := t.X.(*ast.Ident)
			if !ok {
				return nil
			}
			if x.Name != iotaIdentifier {
				return nil
			}
			y, ok := t.Y.(*ast.BasicLit)
			if !ok {
				return nil
			}
			if y.Kind != token.INT {
				return nil
			}
			val, err := strconv.Atoi(y.Value)
			if err != nil {
				return nil
			}
			*idx = val
			enumIota.StartIndex = *idx
		}
	}
	enum.Index = *idx
	*idx++
	// get comment if exists and set descriptio
	if vs.Comment != nil && len(vs.Comment.List) > 0 {
		comment := vs.Comment.List[0].Text[2:]
		valid := !strings.Contains(comment, "invalid")
		if !valid {
			comment = strings.ReplaceAll(comment, "invalid", "")
		}
		enum.Valid = valid
		s1, s2 := strings.SplitBySpace(strings.TrimLeft(comment, " "))
		expectedFields := len(enumIota.Fields)
		if s1 == "" && s2 == "" {
			return &enum
		}
		if s1 != "" && s2 == "" {
			if expectedFields > 0 {
				enum.Fields = parseEnumFields(s1, *enumIota)
				return &enum
			}
			enum.Aliases = parseEnumAliases(s1)
			return &enum
		}
		if s1 != "" && s2 != "" {
			enum.Aliases = parseEnumAliases(s1)
			enum.Fields = parseEnumFields(s2, *enumIota)
			return &enum
		}
	}
	return &enum
}

func parseEnumAliases(s string) []string {
	if strings.Contains(s, ",") {
		aliases := strings.Split(s, ",")
		clnAli := make([]string, len(aliases))
		for i, alias := range aliases {
			if alias[0] == '"' && alias[len(alias)-1] == '"' {
				alias = alias[1 : len(alias)-1]
			}
			clnAli[i] = alias
		}
		return clnAli

	}
	return []string{s}
}

func parseEnumFields(s string, enumIota enum.EnumIota) []enum.Field {
	fieldValues := strings.Split(s, ",")
	enumFields := make([]enum.Field, len(enumIota.Fields))
	for i, f := range enumIota.Fields {
		valRaw := fieldValues[i]
		val := parseValue(valRaw, f.Value)
		enumFields[i] = enum.Field{
			Name:  f.Name,
			Value: val,
		}
	}
	return enumFields
}

func parseValue(valRaw string, t any) any {
	switch t.(type) {
	case bool:
		val, err := strconv.ParseBool(valRaw)
		if err != nil {
			return nil
		}
		return val
	case float64:
		val, err := strconv.ParseFloat(valRaw, 64)
		if err != nil {
			return nil
		}
		return val
	case float32:
		val, err := strconv.ParseFloat(valRaw, 32)
		if err != nil {
			return nil
		}
		return float32(val)
	case int:
		val, err := strconv.Atoi(valRaw)
		if err != nil {
			return nil
		}
		return val
	case int64:
		val, err := strconv.ParseInt(valRaw, 10, 64)
		if err != nil {
			return nil
		}
		return val
	case int32:
		val, err := strconv.ParseInt(valRaw, 10, 32)
		if err != nil {
			return nil
		}
		return int32(val)
	case int16:
		val, err := strconv.ParseInt(valRaw, 10, 16)
		if err != nil {
			return nil
		}
		return int16(val)
	case int8:
		val, err := strconv.ParseInt(valRaw, 10, 8)
		if err != nil {
			return nil
		}
		return int8(val)
	case uint:
		val, err := strconv.ParseUint(valRaw, 10, 64)
		if err != nil {
			return nil
		}
		return uint(val)
	case uint64:
		val, err := strconv.ParseUint(valRaw, 10, 64)
		if err != nil {
			return nil
		}
		return val
	case uint32:
		val, err := strconv.ParseUint(valRaw, 10, 32)
		if err != nil {
			return nil
		}
		return uint32(val)
	case uint16:
		val, err := strconv.ParseUint(valRaw, 10, 16)
		if err != nil {
			return nil
		}
		return uint16(val)
	case uint8:
		val, err := strconv.ParseUint(valRaw, 10, 8)
		if err != nil {
			return nil
		}
		return uint8(val)
	case string:
		if valRaw[0] == '"' && valRaw[len(valRaw)-1] == '"' {
			return valRaw[1 : len(valRaw)-1]
		}
		return valRaw
	case time.Time:
		val, err := time.Parse(time.RFC3339, valRaw)
		if err != nil {
			return nil
		}
		return val
	case time.Duration:
		val, err := time.ParseDuration(valRaw)
		if err != nil {
			return nil
		}
		return val
	default:
		return nil
	}
}

func (p *Parser) getEnumFields(valueStr string, enumIota enum.EnumIota) []enum.Field {
	if valueStr == "" {
		return nil
	}
	values := strings.Split(valueStr, ",")
	if len(values) != len(enumIota.Fields) {
		return nil
	}
	fields := make([]enum.Field, 0)
	for i, value := range values {
		fields = append(fields, enum.Field{
			Name:  enumIota.Fields[i].Name,
			Value: value,
		})
	}
	return fields
}

func (p *Parser) getEnumIotas(node *ast.File) []enum.EnumIota {
	var enumIotas []enum.EnumIota
	for _, decl := range node.Decls {
		switch t := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range t.Specs {
				switch ts := spec.(type) {
				case *ast.TypeSpec:
					if ts.Type != nil {
						enumIota := enum.EnumIota{
							Type: ts.Name.Name,
						}
						if ts.Comment != nil &&
							len(ts.Comment.List) > 0 {
							comment := ts.Comment.List[0].Text
							if strings.HasPrefix(comment, "//") {
								comment = comment[2:]
							}
							opener, closer, fields := extractFields(comment)
							enumIota.Comment = comment
							enumIota.Fields = fields
							enumIota.Opener = opener
							enumIota.Closer = closer
						}
						enumIotas = append(enumIotas, enumIota)
					}
				}
			}
		}
	}
	return enumIotas
}

func extractFields(comment string) (string, string, []enum.Field) {
	fields := make([]enum.Field, 0)
	comment = strings.TrimSpace(comment)
	open, closer := " ", " "
	if comment == "" {
		return open, closer, fields
	}
	fieldVals := strings.Split(comment, ",")
	for _, val := range fieldVals {
		field := strings.TrimSpace(val)
		open, closer = openCloser(field)

		nO, nC, tO, tC := 0, 0, 0, 0
		n, f := "", ""

		if open == " " {
			extra := strings.Split(field, " ")
			if len(extra) > 1 {
				n = extra[0]
				f = extra[1]
			} else {
				f = extra[0]
			}
			fields = append(fields, enum.Field{
				Name:  n,
				Value: fieldToType(f),
			})
			continue
		}

		nO = strings.Index(field, open)
		if nO == -1 {
			continue
		}
		nC = strings.Index(field[nO:], closer) + nO
		if nC == -1 {
			continue
		}
		tO = nO + len(open)
		tC = nC
		n = field[:nO]
		f = field[tO:tC]
		fields = append(fields, enum.Field{
			Name:  n,
			Value: fieldToType(f),
		})
	}
	return open, closer, fields
}

func openCloser(field string) (string, string) {
	open := " "
	closer := " "
	if strings.Contains(field, "[") {
		open = "["
		closer = "]"
	} else if strings.Contains(field, "(") {
		open = "("
		closer = ")"
	}
	return open, closer
}

func fieldToType(field string) any {
	f := strings.TrimSpace(field)
	switch f {
	case "bool":
		return false
	case "int":
		return 0
	case "string":
		return ""
	case "time.Duration":
		return time.Duration(0)
	case "time.Time":
		return time.Time{}
	case "float64":
		return 0.0
	case "float32":
		return float32(0.0)
	case "int64":
		return int64(0)
	case "int32":
		return int32(0)
	case "int16":
		return int16(0)
	case "int8":
		return int8(0)
	case "uint64":
		return uint64(0)
	case "uint32":
		return uint32(0)
	case "uint16":
		return uint16(0)
	case "uint8":
		return uint8(0)
	case "uint":
		return uint(0)
	case "byte":
		return byte(0)
	case "rune":
		return rune(0)
	case "complex64":
		return complex64(0)
	case "complex128":
		return complex128(0)
	case "uintptr":
		return uintptr(0)
	default:
		return nil
	}
}

// // getTypeComments collects all comments associated with type declarations.
// // This builds a mapping of type names to their documentation comments.
// func (p *Parser) getTypeComments(node *ast.File) typeComments {
// 	typeComms := make(map[string]string)
// 	ast.Inspect(node, func(n ast.Node) bool {
// 		decl, ok := n.(*ast.GenDecl)
// 		if !ok || decl.Tok != token.TYPE {
// 			return true
// 		}
// 		for _, spec := range decl.Specs {
// 			typeSpec, ok := spec.(*ast.TypeSpec)
// 			if !ok || typeSpec.Comment == nil || len(typeSpec.Comment.List) == 0 {
// 				continue
// 			}
// 			comment := strings.TrimSpace(typeSpec.Comment.List[0].Text[2:])
// 			typeComms[typeSpec.Name.Name] = comment
// 		}
// 		return true
// 	})
// 	return typeComms
// }

// // tempHolder is a temporary struct used to collect enum representations
// // during parsing.
// type tempHolder struct {
// 	enums      []enum.Enum
// 	iotaType   string
// 	iotaIdx    int
// 	nameTPairs []enum.NameTypePair
// }

// // collectRepresentations analyzes an AST to identify enum-like declarations.
// // It extracts type information, constant values, and associated metadata to build
// // complete enum representations for code generation.
// func (p *Parser) collectRepresentations(node *ast.File,
// 	filename string, typeComments map[string]string,
// 	cfg config.Configuration) []enum.Representation {
// 	packageName := p.getPackageName(node)
// 	slog.Default().Debug("enum package name", "name", packageName)
// 	enumsByType := make(map[string]tempHolder)
// 	slog.Default().Debug("traversing ast")
// 	ast.Inspect(node, func(n ast.Node) bool {
// 		decl, ok := n.(*ast.GenDecl)
// 		if !ok || decl.Tok != token.CONST {
// 			return true
// 		}
// 		currNTPs, currIotaType, idx := p.parseEnumNameTypePairs(decl, typeComments)
// 		if currIotaType == "" {
// 			return true
// 		}
// 		entry, exists := enumsByType[currIotaType]
// 		if !exists {
// 			entry = tempHolder{
// 				iotaType:   currIotaType,
// 				iotaIdx:    idx,
// 				nameTPairs: currNTPs,
// 			}
// 		}
// 		for _, spec := range decl.Specs {
// 			vs, ok := spec.(*ast.ValueSpec)
// 			if !ok {
// 				continue
// 			}
// 			for _, name := range vs.Names {
// 				if name.Name == "_" {
// 					entry.iotaIdx++
// 					continue
// 				}
// 				enumValue := p.getEnumValue(idx, vs, decl)
// 				comment := p.getComment(vs)
// 				valid := !strings.Contains(comment, "invalid")
// 				if !valid {
// 					comment = strings.ReplaceAll(comment, "invalid", "")
// 				}
// 				var (
// 					aliases   = make([]string, 0)
// 					valueStrs = make([]string, 0)
// 				)
// 				s1, s2 := strings.SplitBySpace(strings.TrimLeft(comment, " "))
// 				hasAliases := false
// 				if s1 == "" && s2 == "" {
// 					hasAliases = false
// 				}
// 				if s2 != "" && len(currNTPs) >= 1 {
// 					hasAliases = true
// 					aliases = strings.Split(strings.TrimSpace(s1), ",")
// 					vsr := strings.Split(strings.TrimSpace(s2), ",")
// 					for _, v := range vsr {
// 						if v != "" {
// 							valueStrs = append(valueStrs, v)
// 						}
// 					}
// 				}
// 				if s1 != "" && len(currNTPs) == 0 {
// 					hasAliases = true
// 					aliases = strings.Split(strings.TrimSpace(s1), ",")
// 					vsr := strings.Split(strings.TrimSpace(s2), ",")
// 					for _, v := range vsr {
// 						if v != "" {
// 							valueStrs = append(valueStrs, v)
// 						}
// 					}
// 				}
// 				if !hasAliases {
// 					vsr := strings.Split(strings.TrimSpace(s1), ",")
// 					for _, v := range vsr {
// 						if v != "" {
// 							valueStrs = append(valueStrs, v)
// 						}
// 					}
// 					aliases = []string{}
// 				}

// 				ntps := p.copyNameTPairs(currNTPs, valueStrs)
// 				camel := strings.CamelCase(currIotaType)
// 				if strings.IsRegularPlural(camel) {
// 					camel = strings.Singular(camel)
// 				}
// 				alias := name.Name
// 				if len(aliases) > 0 {
// 					alias = aliases[0]
// 				}
// 				entry.enums = append(entry.enums, enum.Enum{
// 					Info: enum.Info{
// 						Name:    name.Name,
// 						Camel:   strings.CamelCase(name.Name),
// 						Lower:   strings.ToLower(name.Name),
// 						Upper:   strings.ToUpper(name.Name),
// 						Alias:   alias,
// 						Aliases: append([]string{alias}, aliases...),
// 						Value:   enumValue,
// 						Valid:   valid,
// 					},
// 					TypeInfo: enum.TypeInfo{
// 						Name:         currIotaType,
// 						Camel:        camel,
// 						Lower:        strings.ToLower(currIotaType),
// 						Upper:        strings.ToUpper(currIotaType),
// 						NameTypePair: ntps,
// 						Index:        entry.iotaIdx,
// 					},
// 					Raw: enum.Raw{
// 						Comment:     comment,
// 						TypeComment: p.getTypeComment(vs, typeComments),
// 					},
// 				})
// 				entry.iotaIdx++
// 			}
// 			enumsByType[currIotaType] = entry
// 		}
// 		return false
// 	})
// 	if len(enumsByType) == 0 {
// 		return nil
// 	}
// 	representations := make([]enum.Representation, 0, len(enumsByType))
// 	for iotaType, info := range enumsByType {
// 		slog.Default().Debug("enum representation for type", "type", iotaType)
// 		for _, v := range info.enums {
// 			slog.Default().Debug("enum information", "enum", v.Info.Name)
// 		}
// 		plural := strings.Plural(iotaType)
// 		lowerPlural := string(unicode.ToLower(rune(plural[0]))) + plural[1:]
// 		camelPlural := strings.CamelCase(lowerPlural)
// 		camel := strings.CamelCase(iotaType)
// 		if strings.IsRegularPlural(camel) {
// 			camel = strings.Singular(camel)
// 		}
// 		minValue := math.MaxInt32
// 		for _, e := range info.enums {
// 			if e.TypeInfo.Index < minValue {
// 				minValue = e.TypeInfo.Index
// 			}
// 		}
// 		rep := enum.Representation{
// 			Version:         version.CURRENT,
// 			GenerationTime:  time.Now(),
// 			MinValue:        minValue,
// 			Enums:           info.enums,
// 			PackageName:     packageName,
// 			Failfast:        cfg.Failfast,
// 			Legacy:          cfg.Legacy,
// 			CaseInsensitive: cfg.Insensitive,
// 			SourceFilename:  filename,
// 			OutputFilename:  strings.ToLower(lowerPlural),
// 			TypeInfo: enum.TypeInfo{
// 				Index:        info.iotaIdx,
// 				Name:         info.iotaType,
// 				Camel:        camel,
// 				Lower:        lowerPlural,
// 				Upper:        strings.ToUpper(info.iotaType),
// 				Plural:       lowerPlural,
// 				PluralCamel:  camelPlural,
// 				NameTypePair: info.nameTPairs,
// 			},
// 		}
// 		representations = append(representations, rep)
// 	}
// 	return representations
// }

// type typeComments = map[string]string

// func (p *Parser) parseEnumNameTypePairs(decl *ast.GenDecl, typeComms typeComments) (
// 	[]enum.NameTypePair, string, int) {
// 	var (
// 		currNTPs     = make([]enum.NameTypePair, 0)
// 		currIotaType = ""
// 		idx          = 0
// 	)
// 	if len(decl.Specs) > 0 {
// 		if valueSpec, ok := decl.Specs[0].(*ast.ValueSpec); ok && len(valueSpec.Values) == 1 {
// 			name, iType, iTypeComm, iIdx := p.iotaInfo(valueSpec, typeComms)
// 			if (name == "" || name == "_") || iType == "" {
// 				return currNTPs, currIotaType, idx
// 			}
// 			currIotaType = iType
// 			idx = iIdx
// 			if iTypeComm != "" {
// 				currNTPs = p.nameTPairsFromComments(iTypeComm, currNTPs)
// 			}
// 		}
// 	}
// 	return currNTPs, currIotaType, idx
// }

// // getEnumValue returns the value of the enum at the given index.
// // If the value is not specified or calculated, it returns the index.
// func (p *Parser) getEnumValue(idx int, vs *ast.ValueSpec, decl *ast.GenDecl) int {
// 	// Find the position of this ValueSpec in the declaration
// 	specIndex := 0
// 	for i, s := range decl.Specs {
// 		if s == vs {
// 			specIndex = i
// 			break
// 		}
// 	}

// 	// For iota-based enums, return the index directly
// 	if len(vs.Values) == 0 {
// 		return specIndex
// 	}

// 	// Handle explicit values
// 	if len(vs.Values) > 0 {
// 		// Check for binary expressions (like x + 1)
// 		if binExpr, ok := vs.Values[0].(*ast.BinaryExpr); ok {
// 			return p.specIndex(binExpr, specIndex, idx)
// 		}

// 		// Handle literal values
// 		if lit, ok := vs.Values[0].(*ast.BasicLit); ok {
// 			val, err := strconv.Atoi(lit.Value)
// 			if err == nil {
// 				return val
// 			}
// 		}
// 	}

// 	// Default to using the spec index
// 	return specIndex
// }

// // specIndex returns the index of the enum value in the declaration.
// // handles cases where the enum value is defined as an expression.
// func (*Parser) specIndex(expr *ast.BinaryExpr, specIdx int, idx int) int {
// 	if x, ok := expr.X.(*ast.Ident); ok && x.Name == iotaIdentifier {
// 		if lit, ok := expr.Y.(*ast.BasicLit); ok {
// 			if num, err := strconv.Atoi(lit.Value); err == nil {
// 				switch expr.Op {
// 				case token.ADD:
// 					return specIdx + (num - idx)
// 				case token.SUB:
// 					return specIdx - (num - idx)
// 				case token.MUL:
// 					return specIdx * (num - idx)
// 				case token.QUO:
// 					return specIdx / (num - idx)
// 				}
// 			}
// 		}
// 	}
// 	return idx
// }

// // getTypeComment retrieves the documentation comment associated with a type.
// // This is used to extract metadata about enum types from their definitions.
// func (p *Parser) getTypeComment(valueSpec *ast.ValueSpec, typeComments typeComments) string {
// 	if valueSpec.Type != nil {
// 		constantType := fmt.Sprintf("%s", valueSpec.Type)
// 		if comment, exists := typeComments[constantType]; exists {
// 			return comment
// 		}
// 	}
// 	return ""
// }

// // getPackageName extracts the package name from an AST node.
// // This is used to determine the package context for generated code.
// func (p *Parser) getPackageName(node *ast.File) string {
// 	var packageName string
// 	if node.Name != nil {
// 		packageName = node.Name.Name
// 	}
// 	return packageName
// }

// // copyNameTPairs creates a copy of name-type pairs with updated values.
// // This ensures that each enum representation has its own isolated metadata.
// func (p *Parser) copyNameTPairs(nameTPairs []enum.NameTypePair, values []string) []enum.NameTypePair {
// 	nameTPairsCopy := slices.Clone(nameTPairs)
// 	for i, pair := range nameTPairsCopy {
// 		if i >= len(values) {
// 			break
// 		}
// 		v := strings.TrimSpace(values[i])
// 		pair.Value = formatValueByType(v, pair.Type)
// 		nameTPairsCopy[i] = pair
// 	}
// 	return nameTPairsCopy
// }

// func formatValueByType(v, typeName string) string {
// 	switch typeName {
// 	case "uint":
// 		val := parseOrDefault(v, uint(0), func(s string) (uint, error) {
// 			parsed, err := strconv.ParseUint(s, 10, strconv.IntSize)
// 			return uint(parsed), err
// 		})
// 		return strconv.FormatUint(uint64(val), 10)
// 	case "uint8":
// 		val := parseOrDefault(v, uint8(0), func(s string) (uint8, error) {
// 			parsed, err := strconv.ParseUint(s, 10, 8)
// 			if err != nil {
// 				return 0, err
// 			}
// 			return uint8(parsed), nil
// 		})
// 		return strconv.FormatUint(uint64(val), 10)
// 	case "uint16":
// 		val := parseOrDefault(v, uint16(0), func(s string) (uint16, error) {
// 			parsed, err := strconv.ParseUint(s, 10, 16)
// 			if err != nil {
// 				return 0, err
// 			}
// 			return uint16(parsed), nil
// 		})
// 		return strconv.FormatUint(uint64(val), 10)
// 	case "uint32":
// 		val := parseOrDefault(v, uint32(0), func(s string) (uint32, error) {
// 			parsed, err := strconv.ParseUint(s, 10, 32)
// 			if err != nil {
// 				return 0, err
// 			}
// 			return uint32(parsed), nil
// 		})
// 		return strconv.FormatUint(uint64(val), 10)
// 	case "uint64":
// 		val := parseOrDefault(v, uint64(0), func(s string) (uint64, error) {
// 			return strconv.ParseUint(s, 10, 64)
// 		})
// 		return strconv.FormatUint(val, 10)
// 	case "int":
// 		val := parseOrDefault(v, int(0), strconv.Atoi)
// 		return strconv.Itoa(val)
// 	case "int8":
// 		val := parseOrDefault(v, int8(0), func(s string) (int8, error) {
// 			parsed, err := strconv.ParseInt(s, 10, 8)
// 			if err != nil {
// 				return 0, err
// 			}
// 			return int8(parsed), nil
// 		})
// 		return strconv.Itoa(int(val))
// 	case "int16":
// 		val := parseOrDefault(v, int16(0), func(s string) (int16, error) {
// 			parsed, err := strconv.ParseInt(s, 10, 16)
// 			if err != nil {
// 				return 0, err
// 			}
// 			return int16(parsed), nil
// 		})
// 		return strconv.Itoa(int(val))
// 	case "int32":
// 		val := parseOrDefault(v, int32(0), func(s string) (int32, error) {
// 			parsed, err := strconv.ParseInt(s, 10, 32)
// 			if err != nil {
// 				return 0, err
// 			}
// 			return int32(parsed), nil
// 		})
// 		return strconv.Itoa(int(val))
// 	case "int64":
// 		val := parseOrDefault(v, int64(0), func(s string) (int64, error) {
// 			return strconv.ParseInt(s, 10, 64)
// 		})
// 		return strconv.FormatInt(val, 10)
// 	case "float32":
// 		val := parseOrDefault(v, float32(0), func(s string) (float32, error) {
// 			parsed, err := strconv.ParseFloat(s, 32)
// 			if err != nil {
// 				return 0, err
// 			}
// 			return float32(parsed), nil
// 		})
// 		return fmt.Sprintf("%f", val)
// 	case "float64":
// 		val := parseOrDefault(v, float64(0), func(s string) (float64, error) {
// 			return strconv.ParseFloat(s, 64)
// 		})
// 		return fmt.Sprintf("%f", val)
// 	case "bool":
// 		val := parseOrDefault(v, false, strconv.ParseBool)
// 		return strconv.FormatBool(val)
// 	case "string":
// 		return fmt.Sprintf("%q", v)
// 	case "time.Duration":
// 		val := parseOrDefault(v, 0, time.ParseDuration)
// 		hours := val.Hours()
// 		str := fmt.Sprintf("time.Hour * %d", int(hours))
// 		if hours != math.Floor(hours) {
// 			if val.Minutes() == math.Floor(val.Minutes()) {
// 				str = fmt.Sprintf("time.Minute * %d", int(val.Minutes()))
// 			} else if val.Seconds() == math.Floor(val.Seconds()) {
// 				str = fmt.Sprintf("time.Second * %d", int(val.Seconds()))
// 			}
// 		}
// 		return str
// 	case "time.Time":
// 		val := parseOrDefault(v, time.Time{}, func(s string) (time.Time, error) {
// 			t, err := time.Parse(time.RFC3339, s)
// 			if err == nil {
// 				return t, nil
// 			}
// 			t, err = time.Parse(time.DateOnly, s)
// 			if err == nil {
// 				return t, nil
// 			}
// 			t, err = time.Parse(time.RFC3339Nano, s)
// 			if err == nil {
// 				return t, nil
// 			}
// 			t, err = time.Parse(time.RFC1123, s)
// 			if err == nil {
// 				return t, nil
// 			}
// 			t, err = time.Parse(time.RFC1123Z, s)
// 			if err == nil {
// 				return t, nil
// 			}
// 			t, err = time.Parse(time.RFC822, s)
// 			if err == nil {
// 				return t, nil
// 			}
// 			t, err = time.Parse(time.RFC822Z, s)
// 			if err == nil {
// 				return t, nil
// 			}
// 			t, err = time.Parse(time.RFC850, s)
// 			if err == nil {
// 				return t, nil
// 			}
// 			return time.Time{}, err
// 		})
// 		return val.Format(time.RFC3339)
// 	default:
// 		return v
// 	}
// }

// // getComment retrieves the comment associated with a value specification.
// // This extracts documentation from the source code for use in generation.
// func (p *Parser) getComment(valueSpec *ast.ValueSpec) string {
// 	comment := ""
// 	if valueSpec.Comment != nil && len(valueSpec.Comment.List) > 0 {
// 		comment = valueSpec.Comment.List[0].Text
// 		comment = comment[2:]
// 	}
// 	return comment
// }

// // nameTPairsFromComments parses type comments to extract name-type pairs.
// // This allows for metadata extraction from type documentation.
// func (p *Parser) nameTPairsFromComments(iotaTypeComment string, nameTPairs []enum.NameTypePair) []enum.NameTypePair {
// 	typeValues := strings.Split(iotaTypeComment, ",")
// 	for _, v := range typeValues {
// 		if len(v) == 0 {
// 			continue
// 		}
// 		v = strings.TrimSpace(v)
// 		var (
// 			formatType         = "space"
// 			openR, closeR      = " ", " "
// 			nEnd, tStart, tEnd int
// 		)
// 		if strings.Contains(v, "[") {
// 			formatType = "bracket"
// 			openR, closeR = "[", "]"
// 		} else if strings.Contains(v, "(") {
// 			formatType = "parenthesis"
// 			openR, closeR = "(", ")"
// 		}
// 		nEnd = strings.Index(v, openR)
// 		if nEnd == -1 {
// 			continue
// 		}
// 		tStart = nEnd + len(openR)
// 		tEnd = len(v)

// 		if formatType != "space" {
// 			tEnd = strings.Index(v[tStart:], closeR)
// 			if tEnd == -1 {
// 				continue
// 			}
// 			tEnd += tStart
// 		}
// 		name := strings.TrimSpace(v[:nEnd])
// 		typeName := strings.TrimSpace(v[tStart:tEnd])
// 		if name == "" || typeName == "" {
// 			continue
// 		}
// 		nameTypePair := enum.NameTypePair{
// 			Name:  name,
// 			Type:  typeName,
// 			Value: fmt.Sprintf("%s%s%s", openR, typeName, closeR),
// 		}
// 		nameTPairs = append(nameTPairs, nameTypePair)
// 	}
// 	return nameTPairs
// }

// // iotaIdentifier is the token that identifies iota-based enum declarations
// const (
// 	iotaIdentifier = "iota"
// )

// // iotaInfo extracts information about iota-based enum declarations.
// // It identifies the enum type, name, and starting index value.
// func (p *Parser) iotaInfo(valueSpec *ast.ValueSpec, typeComments typeComments) (
// 	string, string, string, int) {
// 	if len(valueSpec.Values) == 0 ||
// 		len(valueSpec.Names) == 0 {
// 		return "", "", "", 0
// 	}
// 	var (
// 		iotaName, iotaType, iotaTypeComment string
// 		iotaIdx                             int
// 		vsVal                               = valueSpec.Values[0]
// 		vsName                              = valueSpec.Names[0]
// 	)
// 	ident, ok := vsVal.(*ast.Ident)
// 	if ok && ident.Name == iotaIdentifier {
// 		iotaName = vsName.Name
// 		if valueSpec.Type != nil {
// 			iotaType = fmt.Sprintf("%s", valueSpec.Type)
// 			if comment, exists := typeComments[iotaType]; exists {
// 				iotaTypeComment = comment
// 			}
// 		}
// 	}
// 	if !ok {
// 		if be, ok := vsVal.(*ast.BinaryExpr); ok {
// 			if x, ok := be.X.(*ast.Ident); ok {
// 				if x.Name == iotaIdentifier {
// 					iotaName = vsName.Name
// 					if valueSpec.Type != nil {
// 						iotaType = fmt.Sprintf("%s", valueSpec.Type)
// 						if comment, exists := typeComments[iotaType]; exists {
// 							iotaTypeComment = comment
// 						}
// 					}
// 				}
// 			}
// 			if be.Op == token.ADD {
// 				if x, ok := be.X.(*ast.Ident); ok && x.Name == iotaIdentifier {
// 					if y, ok := be.Y.(*ast.BasicLit); ok {
// 						if idx, err := strconv.Atoi(y.Value); err == nil {
// 							iotaIdx = idx
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return iotaName, iotaType, iotaTypeComment, iotaIdx
// }

// // Parsable is a constraint that permits any ordered type: any type
// // that supports the operators < <= >= >.
// type Parsable interface {
// 	cmp.Ordered | bool | time.Time
// }

// // parseOrDefault is a generic function that attempts to parse a string as type T,
// // returning the parsed value if successful or the default value if not.
// func parseOrDefault[T Parsable](s string, defaultVal T, parser func(string) (T, error)) T {
// 	if val, err := parser(s); err == nil {
// 		return val
// 	}
// 	return defaultVal
// }
