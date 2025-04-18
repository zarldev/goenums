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
	// ErrReadSource indicates an error occurred while reading the source file.
	ErrReadSource = errors.New("failed to read source content")
	// ErrParseGoFile indicates an error occurred while parsing the source file.
	ErrParseGoFile = errors.New("failed to parse Go file")
)

// Parser implements the enum.Parser interface for Go source files.
// It analyzes Go constant declarations to identify and extract enum patterns,
// translating them into a standardized representation model.
type Parser struct {
	Configuration config.Configuration
	source        enum.Source
}

type ParserOption func(*Parser)

func WithSource(source enum.Source) ParserOption {
	return func(p *Parser) {
		p.source = source
	}
}

// NewParser creates a new Go file parser with the specified configuration and source.
// The parser will analyze the source according to the configuration settings.
func NewParser(configuration config.Configuration, opts ...ParserOption) *Parser {
	p := Parser{
		Configuration: configuration,
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
			slog.Error("unexpected panic in parser",
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
		return nil, fmt.Errorf("%w: %w", ErrReadSource, err)
	}
	slog.Debug("parsing source content")
	filename := p.source.Filename()
	fset := token.NewFileSet()
	slog.Debug("parsing file", "filename", filename)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	node, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseGoFile, err)
	}
	slog.Debug("collecting all enum representations")
	typeComments := p.getTypeComments(node)
	reps := p.collectRepresentations(node, filename, typeComments,
		p.Configuration)
	slog.Debug("collected all enum representations", "count", len(reps))
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
	slog.Debug("enum package name", "name", packageName)
	enumsByType := make(map[string]tempHolder)
	slog.Debug("traversing ast")
	ast.Inspect(node, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.CONST {
			return true
		}
		var (
			currIotaType string
			currIotaIdx  int
		)
		currNTPs := make([]enum.NameTypePair, 0)
		if len(decl.Specs) > 0 {
			if valueSpec, ok := decl.Specs[0].(*ast.ValueSpec); ok && len(valueSpec.Values) == 1 {
				iotaName, iotaType, iotaTypeComment, iotaIdx := p.iotaInfo(valueSpec, typeComments)
				if iotaName != "" && iotaType != "" {
					currIotaType = iotaType
					currIotaIdx = iotaIdx
					if iotaTypeComment != "" {
						currNTPs = p.nameTPairsFromComments(iotaTypeComment, currNTPs)
					}
				}
			}
		}
		if currIotaType != "" {
			entry, exists := enumsByType[currIotaType]
			if !exists {
				entry = tempHolder{
					iotaType:   currIotaType,
					iotaIdx:    currIotaIdx,
					nameTPairs: currNTPs,
				}
			}
			for i, spec := range decl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					comment := p.getComment(vs)
					valid := !strings.Contains(comment, "invalid")
					if !valid {
						comment = strings.ReplaceAll(comment, "invalid", "")
					}
					comment, primaryAlias, additionalAliases := p.getAliasNames(comment, name)
					values := p.getValues(comment)
					nameTPairsCopy := p.copyNameTPairs(currNTPs, values)
					entry.enums = append(entry.enums, enum.Enum{
						Info: enum.Info{
							Name:    name.Name,
							Camel:   strings.CamelCase(name.Name),
							Lower:   strings.ToLower(name.Name),
							Upper:   strings.ToUpper(name.Name),
							Alias:   primaryAlias,
							Aliases: append([]string{primaryAlias}, additionalAliases...),
							Value:   i,
							Valid:   valid,
						},
						TypeInfo: enum.TypeInfo{
							Name:         currIotaType,
							Camel:        strings.CamelCase(currIotaType),
							Lower:        strings.ToLower(currIotaType),
							Upper:        strings.ToUpper(currIotaType),
							NameTypePair: nameTPairsCopy,
						},
						Raw: enum.Raw{
							Comment:     comment,
							TypeComment: p.getTypeComment(vs, typeComments),
						},
					})
				}
			}
			enumsByType[currIotaType] = entry
		}
		return true
	})
	if len(enumsByType) == 0 {
		return nil
	}
	representations := make([]enum.Representation, 0, len(enumsByType))
	for iotaType, info := range enumsByType {
		slog.Debug("enum representation for type", "type", iotaType)
		for _, v := range info.enums {
			slog.Debug("enum information", "enum", v.Info.Name)
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
				Filename:     packageName,
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

// getTypeComment retrieves the documentation comment associated with a type.
// This is used to extract metadata about enum types from their definitions.
func (p *Parser) getTypeComment(valueSpec *ast.ValueSpec, typeComments map[string]string) string {
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
func (p *Parser) getTypeComments(node *ast.File) map[string]string {
	typeComments := make(map[string]string)
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
			typeComments[typeSpec.Name.Name] = comment
		}
		return true
	})
	return typeComments
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

	case "int":
		val := parseOrDefault(v, 0, func(s string) (int, error) {
			return strconv.Atoi(s)
		})
		return fmt.Sprintf("%d", val)
	case "bool":
		val := parseOrDefault(v, false, strconv.ParseBool)
		return fmt.Sprintf("%t", val)
	case "string":
		return fmt.Sprintf("%q", v)
	case "uint", "uint32", "uint16", "uint8", "uint64":
		val := parseOrDefault(v, 0, func(s string) (uint64, error) {
			return strconv.ParseUint(s, 10, 64)
		})
		return fmt.Sprintf("%d", val)
	case "int64", "int32", "int16", "int8":
		val := parseOrDefault(v, 0, func(s string) (int64, error) {
			return strconv.ParseInt(s, 10, 64)
		})
		return fmt.Sprintf("%d", val)
	case "float32":
		val := parseOrDefault(v, 0.0, func(s string) (float64, error) {
			return strconv.ParseFloat(s, 32)
		})
		return fmt.Sprintf("%g", val)
	case "float64":
		val := parseOrDefault(v, 0.0, func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		})
		return fmt.Sprintf("%g", val)
	case "time.Duration":
		val := parseOrDefault(v, 0, func(s string) (time.Duration, error) {
			return time.ParseDuration(s)
		})
		return fmt.Sprintf("%d", val)
	default:
		return formatValue(v)
	}
}

// getAliasNames extracts primary alias and additional aliases from comments.
// This allows for separate string representations of enum values, including comma-separated aliases.
// It returns the updated comment string, the primary alias, and a slice of additional aliases.
func (p *Parser) getAliasNames(comment string, n *ast.Ident) (string, string, []string) {
	if strings.LastIndex(comment, " ") < 1 {
		comment = strings.TrimLeft(comment, " ")
		return comment, n.Name, nil
	}
	comment = strings.TrimLeft(comment, " ")
	// Initialize the primary alias to the enum name by default
	primaryAlias := n.Name
	var additionalAliases []string

	// Handle empty comments
	if comment == "" {
		return "", primaryAlias, nil
	}
	aliasesStr := comment
	// Check for quoted values first
	if strings.HasPrefix(aliasesStr, `"`) {
		endQI := strings.Index(aliasesStr[1:], `"`)
		if endQI != -1 {
			// Extract the main quoted value as the primary alias
			primaryAlias = aliasesStr[1 : endQI+1]

			// Look for comma-separated aliases after the quoted part
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
	// Check for comma-separated values in the comment
	if strings.Contains(aliasesStr, ",") {
		parts := strings.Split(aliasesStr, ",")
		firstPart := strings.TrimSpace(parts[0])

		// The first part may be a single alias or a comment with spaces
		if strings.Count(firstPart, " ") == 0 {
			// If first part is a single word, use it as the primary alias
			primaryAlias = firstPart
		} else if strings.Count(firstPart, " ") == 1 {
			// If it's two words, the first might be an alias
			words := strings.Split(firstPart, " ")
			if !strings.Contains(words[0], "invalid") {
				primaryAlias = words[0]
			}
		}

		// Process the remaining parts as additional aliases
		for i := 1; i < len(parts); i++ {
			alias := strings.TrimSpace(parts[i])
			if alias != "" {
				// Remove any quotes from aliases
				if strings.HasPrefix(alias, `"`) && strings.HasSuffix(alias, `"`) {
					alias = alias[1 : len(alias)-1]
				}
				additionalAliases = append(additionalAliases, alias)
			}
		}

		return comment, primaryAlias, additionalAliases
	}

	// Handle single-word comments (no commas, no spaces)
	if strings.Count(aliasesStr, " ") == 0 {
		// If comment is just a single word and not "invalid", use it as the alias
		if !strings.Contains(aliasesStr, "invalid") {
			primaryAlias = aliasesStr
		}
		return comment, primaryAlias, nil
	}

	// Handle two-word comments (like "ALIAS description")
	if strings.Count(comment, " ") == 1 {
		parts := strings.Split(comment, " ")
		if len(parts) == 2 {
			if strings.Contains(parts[0], "invalid") {
				// If first word contains "invalid", use the second word
				primaryAlias = parts[1]
			} else {
				// Otherwise, first word is likely the alias
				primaryAlias = parts[0]
			}
		}
		return comment, primaryAlias, nil
	}

	// For more complex comments, just return the original
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

		// Remove quotes if present
		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			part = part[1 : len(part)-1]
		}

		result = append(result, part)
	}

	return result
}

// Keep the original getAliasName for backward compatibility
// but make it use the new implementation
func (p *Parser) getAliasName(comment string, n *ast.Ident, nameTPairs []enum.NameTypePair) (string, string) {
	updatedComment, primaryAlias, _ := p.getAliasNames(comment, n)
	return updatedComment, primaryAlias
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
		var (
			formatType         = "unknown"
			openR, closeR      string
			nEnd, tStart, tEnd int
		)
		v = strings.TrimSpace(v)
		if strings.Contains(v, "[") {
			formatType = "bracket"
			openR, closeR = "[", "]"
		} else if strings.Contains(v, "(") {
			formatType = "parenthesis"
			openR, closeR = "(", ")"
		} else if strings.Contains(v, " ") {
			formatType = "space"
			openR, closeR = " ", " "
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
func (p *Parser) iotaInfo(valueSpec *ast.ValueSpec, typeComments map[string]string) (string, string, string, int) {
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

// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
type Ordered interface {
	cmp.Ordered | bool
}

// parseOrDefault is a generic function that attempts to parse a string as type T,
// returning the parsed value if successful or the default value if not.
func parseOrDefault[T Ordered](s string, defaultVal T, parser func(string) (T, error)) T {
	if val, err := parser(s); err == nil {
		return val
	}
	return defaultVal
}

// formatValue formats values for code generation according to their type.
// It ensures that values are properly quoted and formatted in generated code.
func formatValue[T any](val T) string {
	switch v := any(val).(type) {
	case string:
		if v == "" {
			return `""`
		}
		return fmt.Sprintf("%q", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
