// Package gofile provides Go-specific parsing and generation capabilities for enums.
// This parser analyzes Go source files to extract enum-like constant declarations and
// transforms them into language-agnostic enum representations.
package gofile

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"math"
	"slices"
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
func (p *Parser) Parse(ctx context.Context) ([]enum.Representation, error) {
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

func (p *Parser) doParse(ctx context.Context) ([]enum.Representation, error) {
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
	typeComments := p.getTypeComments(node)
	reps := p.collectRepresentations(node, filename, typeComments,
		p.Configuration)
	slog.Default().DebugContext(ctx, "collected all enum representations", "count", len(reps))
	return reps, nil
}

// tempHolder is a temporary struct used to collect enum representations
// during parsing.
type tempHolder struct {
	enums      []enum.Enum
	iotaType   string
	iotaIdx    int
	nameTPairs []enum.NameTypePair
}

// collectRepresentations analyzes an AST to identify enum-like declarations.
// It extracts type information, constant values, and associated metadata to build
// complete enum representations for code generation.
func (p *Parser) collectRepresentations(node *ast.File,
	filename string, typeComments map[string]string,
	cfg config.Configuration) []enum.Representation {
	packageName := p.getPackageName(node)
	slog.Default().Debug("enum package name", "name", packageName)
	enumsByType := make(map[string]tempHolder)
	slog.Default().Debug("traversing ast")
	ast.Inspect(node, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.CONST {
			return true
		}
		currNTPs, currIotaType, idx := p.parseEnumNameTypePairs(decl, typeComments)
		if currIotaType == "" {
			return true
		}
		entry, exists := enumsByType[currIotaType]
		if !exists {
			entry = tempHolder{
				iotaType:   currIotaType,
				iotaIdx:    idx,
				nameTPairs: currNTPs,
			}
		}
		for _, spec := range decl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name == "_" {
					entry.iotaIdx++
					continue
				}
				enumValue := p.getEnumValue(idx, vs, decl)
				comment := p.getComment(vs)
				valid := !strings.Contains(comment, "invalid")
				if !valid {
					comment = strings.ReplaceAll(comment, "invalid", "")
				}
				comment, alias, addAliases := p.getAliasNames(comment, name)
				values := p.getValues(comment)
				ntps := p.copyNameTPairs(currNTPs, values)
				entry.enums = append(entry.enums, enum.Enum{
					Info: enum.Info{
						Name:    name.Name,
						Camel:   strings.CamelCase(name.Name),
						Lower:   strings.ToLower(name.Name),
						Upper:   strings.ToUpper(name.Name),
						Alias:   alias,
						Aliases: append([]string{alias}, addAliases...),
						Value:   enumValue,
						Valid:   valid,
					},
					TypeInfo: enum.TypeInfo{
						Name:         currIotaType,
						Camel:        strings.CamelCase(currIotaType),
						Lower:        strings.ToLower(currIotaType),
						Upper:        strings.ToUpper(currIotaType),
						NameTypePair: ntps,
						Index:        entry.iotaIdx,
					},
					Raw: enum.Raw{
						Comment:     comment,
						TypeComment: p.getTypeComment(vs, typeComments),
					},
				})
			}
			enumsByType[currIotaType] = entry
		}
		return false
	})
	if len(enumsByType) == 0 {
		return nil
	}
	representations := make([]enum.Representation, 0, len(enumsByType))
	for iotaType, info := range enumsByType {
		slog.Default().Debug("enum representation for type", "type", iotaType)
		for _, v := range info.enums {
			slog.Default().Debug("enum information", "enum", v.Info.Name)
		}
		lowerPlural, camelPlural := strings.PluralAndCamelPlural(iotaType)
		rep := enum.Representation{
			Version:        version.CURRENT,
			GenerationTime: time.Now(),

			PackageName:     packageName,
			Failfast:        cfg.Failfast,
			Legacy:          cfg.Legacy,
			CaseInsensitive: cfg.Insensitive,
			SourceFilename:  filename,
			TypeInfo: enum.TypeInfo{
				Index:        info.iotaIdx,
				Name:         info.iotaType,
				Camel:        strings.CamelCase(info.iotaType),
				Lower:        lowerPlural,
				Upper:        strings.ToUpper(info.iotaType),
				Plural:       lowerPlural,
				PluralCamel:  camelPlural,
				NameTypePair: info.nameTPairs,
			},
			Enums: info.enums,
		}
		representations = append(representations, rep)
	}
	return representations
}

type typeComments = map[string]string

func (p *Parser) parseEnumNameTypePairs(decl *ast.GenDecl, typeComms typeComments) (
	[]enum.NameTypePair, string, int) {
	var (
		currNTPs     = make([]enum.NameTypePair, 0)
		currIotaType = ""
		idx          = 0
	)
	if len(decl.Specs) > 0 {
		if valueSpec, ok := decl.Specs[0].(*ast.ValueSpec); ok && len(valueSpec.Values) == 1 {
			name, iType, iTypeComm, iIdx := p.iotaInfo(valueSpec, typeComms)
			if (name == "" || name == "_") || iType == "" {
				return currNTPs, currIotaType, idx
			}
			currIotaType = iType
			idx = iIdx
			if iTypeComm != "" {
				currNTPs = p.nameTPairsFromComments(iTypeComm, currNTPs)
			}
		}
	}
	return currNTPs, currIotaType, idx
}

// getEnumValue returns the value of the enum at the given index.
// If the value is not specified or calculated, it returns the index.
func (p *Parser) getEnumValue(idx int, vs *ast.ValueSpec, decl *ast.GenDecl) int {
	specIndex := 0
	for i, s := range decl.Specs {
		if s == vs {
			specIndex = i
			break
		}
	}
	if specIndex == 0 && len(vs.Values) > 0 {
		if binExpr, ok := vs.Values[0].(*ast.BinaryExpr); ok {
			return p.specIndex(binExpr, specIndex, idx)
		}
	}
	if specIndex == 0 {
		return idx
	}
	if len(decl.Specs) > 0 {
		var (
			vs *ast.ValueSpec
			ok bool
		)
		if vs, ok = decl.Specs[0].(*ast.ValueSpec); !ok {
			return idx
		}
		if len(vs.Values) > 0 {
			if binExpr, ok := vs.Values[0].(*ast.BinaryExpr); ok {
				return p.specIndex(binExpr, specIndex, idx)
			}
		}
		return idx
	}
	return specIndex
}

// specIndex returns the index of the enum value in the declaration.
// handles cases where the enum value is defined as an expression.
func (*Parser) specIndex(expr *ast.BinaryExpr, specIdx int, idx int) int {
	if x, ok := expr.X.(*ast.Ident); ok && x.Name == iotaIdentifier {
		if lit, ok := expr.Y.(*ast.BasicLit); ok {
			if num, err := strconv.Atoi(lit.Value); err == nil {
				switch expr.Op {
				case token.ADD:
					return specIdx + (num - idx)
				case token.SUB:
					return specIdx - (num - idx)
				case token.MUL:
					return specIdx * (num - idx)
				case token.QUO:
					return specIdx / (num - idx)
				default:
					return idx
				}
			}
		}
	}
	return 0
}

// getTypeComment retrieves the documentation comment associated with a type.
// This is used to extract metadata about enum types from their definitions.
func (p *Parser) getTypeComment(valueSpec *ast.ValueSpec, typeComments typeComments) string {
	if valueSpec.Type != nil {
		constantType := fmt.Sprintf("%s", valueSpec.Type)
		if comment, exists := typeComments[constantType]; exists {
			return comment
		}
	}
	return ""
}

// getPackageName extracts the package name from an AST node.
// This is used to determine the package context for generated code.
func (p *Parser) getPackageName(node *ast.File) string {
	var packageName string
	if node.Name != nil {
		packageName = node.Name.Name
	}
	return packageName
}

// getTypeComments collects all comments associated with type declarations.
// This builds a mapping of type names to their documentation comments.
func (p *Parser) getTypeComments(node *ast.File) typeComments {
	typeComms := make(map[string]string)
	ast.Inspect(node, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			return true
		}
		for _, spec := range decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Comment == nil || len(typeSpec.Comment.List) == 0 {
				continue
			}
			comment := strings.TrimSpace(typeSpec.Comment.List[0].Text[2:])
			typeComms[typeSpec.Name.Name] = comment
		}
		return true
	})
	return typeComms
}

// getValues extracts value information from a comment string.
// This is used to parse comma-separated values in enum comments.
func (p *Parser) getValues(comment string) []string {
	values := strings.Split(comment, ",")
	result := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) {
			v = v[1 : len(v)-1]
		}
		result = append(result, v)
	}
	return result
}

// copyNameTPairs creates a copy of name-type pairs with updated values.
// This ensures that each enum representation has its own isolated metadata.
func (p *Parser) copyNameTPairs(nameTPairs []enum.NameTypePair, values []string) []enum.NameTypePair {
	nameTPairsCopy := slices.Clone(nameTPairs)
	for i, pair := range nameTPairsCopy {
		if i >= len(values) {
			break
		}
		v := strings.TrimSpace(values[i])
		pair.Value = formatValueByType(v, pair.Type)
		nameTPairsCopy[i] = pair
	}
	return nameTPairsCopy
}

func formatValueByType(v, typeName string) string {
	switch typeName {
	case "uint":
		val := parseOrDefault(v, uint(0), func(s string) (uint, error) {
			parsed, err := strconv.ParseUint(s, 10, strconv.IntSize)
			return uint(parsed), err
		})
		return strconv.FormatUint(uint64(val), 10)
	case "uint8":
		val := parseOrDefault(v, uint8(0), func(s string) (uint8, error) {
			parsed, err := strconv.ParseUint(s, 10, 8)
			if err != nil {
				return 0, err
			}
			return uint8(parsed), nil
		})
		return strconv.FormatUint(uint64(val), 10)
	case "uint16":
		val := parseOrDefault(v, uint16(0), func(s string) (uint16, error) {
			parsed, err := strconv.ParseUint(s, 10, 16)
			if err != nil {
				return 0, err
			}
			return uint16(parsed), nil
		})
		return strconv.FormatUint(uint64(val), 10)
	case "uint32":
		val := parseOrDefault(v, uint32(0), func(s string) (uint32, error) {
			parsed, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return 0, err
			}
			return uint32(parsed), nil
		})
		return strconv.FormatUint(uint64(val), 10)
	case "uint64":
		val := parseOrDefault(v, uint64(0), func(s string) (uint64, error) {
			return strconv.ParseUint(s, 10, 64)
		})
		return strconv.FormatUint(val, 10)
	case "int":
		val := parseOrDefault(v, int(0), strconv.Atoi)
		return strconv.Itoa(val)
	case "int8":
		val := parseOrDefault(v, int8(0), func(s string) (int8, error) {
			parsed, err := strconv.ParseInt(s, 10, 8)
			if err != nil {
				return 0, err
			}
			return int8(parsed), nil
		})
		return strconv.Itoa(int(val))
	case "int16":
		val := parseOrDefault(v, int16(0), func(s string) (int16, error) {
			parsed, err := strconv.ParseInt(s, 10, 16)
			if err != nil {
				return 0, err
			}
			return int16(parsed), nil
		})
		return strconv.Itoa(int(val))
	case "int32":
		val := parseOrDefault(v, int32(0), func(s string) (int32, error) {
			parsed, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return 0, err
			}
			return int32(parsed), nil
		})
		return strconv.Itoa(int(val))
	case "int64":
		val := parseOrDefault(v, int64(0), func(s string) (int64, error) {
			return strconv.ParseInt(s, 10, 64)
		})
		return strconv.FormatInt(val, 10)
	case "float32":
		val := parseOrDefault(v, float32(0), func(s string) (float32, error) {
			parsed, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return 0, err
			}
			return float32(parsed), nil
		})
		return fmt.Sprintf("%f", val)
	case "float64":
		val := parseOrDefault(v, float64(0), func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		})
		return fmt.Sprintf("%f", val)
	case "bool":
		val := parseOrDefault(v, false, strconv.ParseBool)
		return strconv.FormatBool(val)
	case "string":
		return fmt.Sprintf("%q", v)
	case "time.Duration":
		val := parseOrDefault(v, 0, time.ParseDuration)
		hours := val.Hours()
		str := fmt.Sprintf("time.Hour * %d", int(hours))
		if hours != math.Floor(hours) {
			if val.Minutes() == math.Floor(val.Minutes()) {
				str = fmt.Sprintf("time.Minute * %d", int(val.Minutes()))
			} else if val.Seconds() == math.Floor(val.Seconds()) {
				str = fmt.Sprintf("time.Second * %d", int(val.Seconds()))
			}
		}
		return str
	case "time.Time":
		val := parseOrDefault(v, time.Time{}, func(s string) (time.Time, error) {
			t, err := time.Parse(time.RFC3339, s)
			if err == nil {
				return t, nil
			}
			t, err = time.Parse(time.DateOnly, s)
			if err == nil {
				return t, nil
			}
			t, err = time.Parse(time.RFC3339Nano, s)
			if err == nil {
				return t, nil
			}
			t, err = time.Parse(time.RFC1123, s)
			if err == nil {
				return t, nil
			}
			t, err = time.Parse(time.RFC1123Z, s)
			if err == nil {
				return t, nil
			}
			t, err = time.Parse(time.RFC822, s)
			if err == nil {
				return t, nil
			}
			t, err = time.Parse(time.RFC822Z, s)
			if err == nil {
				return t, nil
			}
			t, err = time.Parse(time.RFC850, s)
			if err == nil {
				return t, nil
			}
			return time.Time{}, err
		})
		return val.Format(time.RFC3339)
	default:
		return v
	}
}

// getAliasNames extracts primary alias and additional aliases from comments.
// This allows for separate string representations of enum values, including comma-separated aliases.
// It returns the updated comment string, the primary alias, and a slice of additional aliases.
func (p *Parser) getAliasNames(comment string, n *ast.Ident) (string, string, []string) {
	if strings.LastIndex(comment, " ") < 1 {
		comment = strings.TrimLeft(comment, " ")
		if comment == "" {
			return "", n.Name, nil
		}
		if strings.Contains(comment, ",") {
			return comment, n.Name, nil
		}
		if !strings.Contains(comment, " ") {
			return "", comment, nil
		}
		return comment, n.Name, nil
	}
	comment = strings.TrimLeft(comment, " ")

	primaryAlias := n.Name
	var additionalAliases []string

	if comment == "" {
		return "", primaryAlias, nil
	}
	aliasesStr := comment
	if strings.HasPrefix(aliasesStr, `"`) {
		endQI := strings.Index(aliasesStr[1:], `"`)
		if endQI != -1 {
			primaryAlias = aliasesStr[1 : endQI+1]
			restOfComment := aliasesStr[endQI+2:]
			commaIdx := strings.Index(restOfComment, ",")
			if commaIdx != -1 {
				aliasPart := restOfComment[commaIdx+1:]
				idx := strings.Index(aliasPart, " \"")
				comment = aliasPart[idx+1:]
				aliasPart = aliasPart[:idx]
				additionalAliases = p.parseCommaList(aliasPart)
			}
			return comment, primaryAlias, additionalAliases
		}
	}
	if strings.Count(aliasesStr, " ") >= 1 {
		comment = aliasesStr[strings.Index(comment, " ")+1:]
		aliasesStr = aliasesStr[:strings.Index(aliasesStr, " ")]
	}
	if strings.Count(aliasesStr, " ") == 0 {
		// If comment is just a single word and not "invalid", use it as the alias
		if !strings.Contains(aliasesStr, "invalid") {
			primaryAlias = aliasesStr
		}
		return comment, primaryAlias, nil
	}
	return comment, primaryAlias, nil
}

// parseCommaList parses a comma-separated string into a slice of trimmed, non-empty strings.
// It handles quoted values and removes the quotes.
func (p *Parser) parseCommaList(s string) []string {
	var result []string
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			part = part[1 : len(part)-1]
		}
		result = append(result, part)
	}
	return result
}

// getComment retrieves the comment associated with a value specification.
// This extracts documentation from the source code for use in generation.
func (p *Parser) getComment(valueSpec *ast.ValueSpec) string {
	comment := ""
	if valueSpec.Comment != nil && len(valueSpec.Comment.List) > 0 {
		comment = valueSpec.Comment.List[0].Text
		comment = comment[2:]
	}
	return comment
}

// nameTPairsFromComments parses type comments to extract name-type pairs.
// This allows for metadata extraction from type documentation.
func (p *Parser) nameTPairsFromComments(iotaTypeComment string, nameTPairs []enum.NameTypePair) []enum.NameTypePair {
	typeValues := strings.Split(iotaTypeComment, ",")
	for _, v := range typeValues {
		if len(v) == 0 {
			continue
		}
		v = strings.TrimSpace(v)
		var (
			formatType         = "space"
			openR, closeR      = " ", " "
			nEnd, tStart, tEnd int
		)
		if strings.Contains(v, "[") {
			formatType = "bracket"
			openR, closeR = "[", "]"
		} else if strings.Contains(v, "(") {
			formatType = "parenthesis"
			openR, closeR = "(", ")"
		}
		nEnd = strings.Index(v, openR)
		if nEnd == -1 {
			continue
		}
		tStart = nEnd + len(openR)
		tEnd = len(v)

		if formatType != "space" {
			tEnd = strings.Index(v[tStart:], closeR)
			if tEnd == -1 {
				continue
			}
			tEnd += tStart
		}
		name := strings.TrimSpace(v[:nEnd])
		typeName := strings.TrimSpace(v[tStart:tEnd])
		if name == "" || typeName == "" {
			continue
		}
		nameTypePair := enum.NameTypePair{
			Name:  name,
			Type:  typeName,
			Value: fmt.Sprintf("%s%s%s", openR, typeName, closeR),
		}
		nameTPairs = append(nameTPairs, nameTypePair)
	}
	return nameTPairs
}

// iotaIdentifier is the token that identifies iota-based enum declarations
const (
	iotaIdentifier = "iota"
)

// iotaInfo extracts information about iota-based enum declarations.
// It identifies the enum type, name, and starting index value.
func (p *Parser) iotaInfo(valueSpec *ast.ValueSpec, typeComments typeComments) (
	string, string, string, int) {
	if len(valueSpec.Values) == 0 ||
		len(valueSpec.Names) == 0 {
		return "", "", "", 0
	}
	var (
		iotaName, iotaType, iotaTypeComment string
		iotaIdx                             int
		vsVal                               = valueSpec.Values[0]
		vsName                              = valueSpec.Names[0]
	)
	ident, ok := vsVal.(*ast.Ident)
	if ok && ident.Name == iotaIdentifier {
		iotaName = vsName.Name
		if valueSpec.Type != nil {
			iotaType = fmt.Sprintf("%s", valueSpec.Type)
			if comment, exists := typeComments[iotaType]; exists {
				iotaTypeComment = comment
			}
		}
	}
	if !ok {
		if be, ok := vsVal.(*ast.BinaryExpr); ok {
			if x, ok := be.X.(*ast.Ident); ok {
				if x.Name == iotaIdentifier {
					iotaName = vsName.Name
					if valueSpec.Type != nil {
						iotaType = fmt.Sprintf("%s", valueSpec.Type)
						if comment, exists := typeComments[iotaType]; exists {
							iotaTypeComment = comment
						}
					}
				}
			}
			if y, ok := be.Y.(*ast.BasicLit); ok {
				if idx, err := strconv.Atoi(y.Value); err == nil {
					iotaIdx = idx
				}
			}
		}
	}
	return iotaName, iotaType, iotaTypeComment, iotaIdx
}

// Parsable is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
type Parsable interface {
	cmp.Ordered | bool | time.Time
}

// parseOrDefault is a generic function that attempts to parse a string as type T,
// returning the parsed value if successful or the default value if not.
func parseOrDefault[T Parsable](s string, defaultVal T, parser func(string) (T, error)) T {
	if val, err := parser(s); err == nil {
		return val
	}
	return defaultVal
}
