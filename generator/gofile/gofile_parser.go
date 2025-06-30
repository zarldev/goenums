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

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/internal/version"
	"github.com/zarldev/goenums/source"
	"github.com/zarldev/goenums/strings"
)

// Compile-time check that Parser implements enum.Parser
var _ enum.Parser = (*Parser)(nil)

var (
	// ErrParseGoSource indicates an error occurred while parsing the source file.
	ErrParseGoSource = errors.New("failed to parse Go source")
	// ErrReadSource indicates an error occurred while reading the source file.
	ErrReadGoSource = errors.New("failed to read Go source")
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
func (p *Parser) Parse(ctx context.Context) ([]enum.GenerationRequest, error) {
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

func (p *Parser) doParse(ctx context.Context) ([]enum.GenerationRequest, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	filename, node, err := p.parseSourceContent(ctx)
	if err != nil {
		return nil, err
	}
	packageName, enInfo, err := extractEnumInfo(ctx, p, node)
	if err != nil {
		return nil, err
	}
	slog.Default().DebugContext(ctx, "collected all enum representations from source", "filename", filename)
	return p.buildGenerationRequests(enInfo, packageName, filename)
}

func (p *Parser) buildGenerationRequests(enInfo enumInfo, packageName string, filename string) ([]enum.GenerationRequest, error) {
	genr := make([]enum.GenerationRequest, len(enInfo.Enums))
	enumIotas := enInfo.Enums
	for i, enumIota := range enumIotas {
		lowerPlural := strings.Pluralise(strings.ToLower(enumIota.Type))
		genr[i] = enum.GenerationRequest{
			Package:        packageName,
			EnumIota:       enumIota,
			Version:        version.CURRENT,
			SourceFilename: filename,
			OutputFilename: strings.ToLower(lowerPlural),
			Configuration:  p.Configuration,
			Imports:        enInfo.Imports,
		}
	}
	return genr, nil
}

func extractEnumInfo(ctx context.Context, p *Parser, node *ast.File) (string, enumInfo, error) {
	slog.Default().DebugContext(ctx, "collecting all enum representations")
	packageName := p.getPackageName(node)
	enInfo := p.getEnumInfo(node)
	slog.Default().DebugContext(ctx, "enum iota", "count", len(enInfo.Enums), "enumIota", enInfo.Enums)
	for i, enumIota := range enInfo.Enums {
		slog.Default().DebugContext(ctx, "enum iota", "enumIota", enumIota)
		enums := p.getEnums(node, &enumIota)
		if len(enums) == 0 {
			return "", enumInfo{}, fmt.Errorf("%w: %w",
				ErrParseGoSource,
				enum.ErrNoEnumsFound)
		}
		slog.Default().DebugContext(ctx, "enums", "count", len(enums), "enums", enums)
		enumIota.Enums = enums
		enInfo.Enums[i] = enumIota
	}
	if len(enInfo.Enums) == 0 {
		slog.Default().DebugContext(ctx, "no valid enums found")
		return "", enumInfo{}, fmt.Errorf("%w: %w",
			ErrParseGoSource,
			enum.ErrNoEnumsFound)
	}
	return packageName, enInfo, nil
}

func (p *Parser) parseSourceContent(ctx context.Context) (string, *ast.File, error) {
	content, err := p.source.Content()
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrReadGoSource, err)
	}
	slog.Default().DebugContext(ctx, "parsing source content")
	filename := p.source.Filename()
	fset := token.NewFileSet()
	if err := ctx.Err(); err != nil {
		return "", nil, err
	}
	slog.Default().DebugContext(ctx, "parsing file", "filename", filename)
	node, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrParseGoSource, err)
	}
	return filename, node, nil
}

func (p *Parser) getPackageName(node *ast.File) string {
	var packageName string
	if node.Name != nil {
		packageName = node.Name.Name
	}
	return packageName
}

func (p *Parser) getEnums(node *ast.File, enumIota *enum.EnumIota) []enum.Enum {
	var enums []enum.Enum
	enumsFound := false
	for _, decl := range node.Decls {
		t, ok := decl.(*ast.GenDecl)
		if !ok || t.Tok != token.CONST {
			continue
		}
		// Check if this const block contains iota with the target type
		blockHasIota := false
		blockHasTargetType := false

		// First pass: check if this const block has both iota and the target type
		for _, spec := range t.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Check for iota in values
			if vs.Values != nil {
				for _, v := range vs.Values {
					if ident, ok := v.(*ast.Ident); ok && ident.Name == iotaIdentifier {
						blockHasIota = true
					}
					// Also check for iota in binary expressions
					if binExpr, ok := v.(*ast.BinaryExpr); ok {
						if x, ok := binExpr.X.(*ast.Ident); ok && x.Name == iotaIdentifier {
							blockHasIota = true
						}
					}
				}
			}

			// Check if this spec has the target type
			if vs.Type != nil {
				if typeIdent, ok := vs.Type.(*ast.Ident); ok && typeIdent.Name == enumIota.Type {
					blockHasTargetType = true
				}
			}
		}

		// Only process this const block if it has both iota and the target type
		if !blockHasIota || !blockHasTargetType {
			continue
		}

		// Second pass: collect enums from this const block
		idx := 0
		constBlockIotaFound := false
		for _, spec := range t.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			e := p.getEnum(vs, &idx, enumIota, &constBlockIotaFound)
			if e == nil {
				continue
			}
			enums = append(enums, *e)
			enumsFound = true
			slog.Default().Debug("enum", "enum", e)
		}
	}
	if !enumsFound {
		return nil
	}
	return enums
}

func (p *Parser) getEnum(vs *ast.ValueSpec, idx *int, enumIota *enum.EnumIota, iotaFound *bool) *enum.Enum {
	if len(vs.Names) == 0 {
		slog.Default().Debug("valuespec has no names")
		return nil
	}
	if vs.Values != nil {
		for _, v := range vs.Values {
			t, ok := v.(*ast.Ident)
			if !ok {
				continue
			}
			if t.Name == "iota" {
				*iotaFound = true
			}
		}
	}
	if vs.Type != nil {
		t, ok := vs.Type.(*ast.Ident)
		if !ok {
			return nil
		}
		if t.Name != enumIota.Type {
			return nil
		}
	}
	name := vs.Names[0].Name
	if name == "_" {
		*idx++
		return nil
	}
	en := enum.Enum{
		Name:  vs.Names[0].Name,
		Valid: true, // enums are valid by default unless marked as invalid
	}
	for _, v := range vs.Values {
		t, ok := v.(*ast.BinaryExpr)
		if !ok {
			continue
		}
		x, ok := t.X.(*ast.Ident)
		if !ok {
			return nil
		}
		if x.Name != iotaIdentifier {
			return nil
		} else {
			*iotaFound = true
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
		// Calculate the actual starting value based on the operation
		switch t.Op {
		case token.ADD:
			enumIota.StartIndex = 0 + val // iota + val, where iota starts at 0
		case token.SUB:
			enumIota.StartIndex = 0 - val // iota - val, where iota starts at 0
		case token.MUL:
			enumIota.StartIndex = 0 * val // iota * val, where iota starts at 0
		case token.QUO:
			if val != 0 {
				enumIota.StartIndex = 0 / val // iota / val, where iota starts at 0
			} else {
				enumIota.StartIndex = 0
			}
		default:
			enumIota.StartIndex = val // fallback to original behavior
		}
	}
	en.Index = *idx                       // 0-based position in enum sequence
	en.Value = enumIota.StartIndex + *idx // actual constant value
	*idx++
	// get comment if exists and set descriptio
	if vs.Comment != nil && len(vs.Comment.List) > 0 {
		commentText := vs.Comment.List[0].Text
		const commentPrefix = "//"
		if len(commentText) < len(commentPrefix) || !strings.HasPrefix(commentText, commentPrefix) {
			return &en
		}
		comment := commentText[len(commentPrefix):]
		valid := !strings.Contains(comment, "invalid")
		if !valid {
			comment = strings.ReplaceAll(comment, "invalid", "")
		}
		en.Valid = valid
		s1, s2 := strings.SplitBySpace(strings.TrimLeft(comment, " "))
		expectedFields := len(enumIota.Fields)

		if s1 == "" && s2 == "" {
			return &en
		}

		if s1 == "" {
			return &en
		}

		if s2 == "" {
			if expectedFields > 0 {
				f, err := enum.ParseEnumFields(s1, *enumIota)
				if err != nil {
					slog.Default().Warn("failed to parse enum fields",
						"enum", vs.Names[0].Name,
						"error", err)
					return &en
				}
				en.Fields = f
				return &en
			}
			en.Aliases = enum.ParseEnumAliases(s1)
			return &en
		}

		// Both s1 and s2 are not empty
		en.Aliases = enum.ParseEnumAliases(s1)
		f, err := enum.ParseEnumFields(s2, *enumIota)
		if err != nil {
			return nil
		}
		en.Fields = f
		return &en
	}
	return &en
}

type enumInfo struct {
	Imports []string
	Enums   []enum.EnumIota
}

func (p *Parser) getEnumInfo(node *ast.File) enumInfo {
	var enumIotas []enum.EnumIota
	for _, decl := range node.Decls {
		t, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range t.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
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
					opener, closer, fields := enum.ExtractFields(comment)

					enumIota.Comment = comment
					enumIota.Fields = fields
					enumIota.Opener = opener
					enumIota.Closer = closer
				}
				enumIotas = append(enumIotas, enumIota)
			}
		}
	}
	imports := enum.ExtractImports(enumIotas)
	return enumInfo{
		Imports: imports,
		Enums:   enumIotas,
	}
}
