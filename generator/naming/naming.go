// Package naming centralizes the generated Go identifiers used by goenums.
package naming

import (
	"errors"
	"fmt"
	"go/token"
	stdstrings "strings"
	"unicode"
	"unicode/utf8"

	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/strings"
)

var (
	// ErrInvalidAccessorIdentifier is returned when an enum accessor cannot be a valid exported Go identifier.
	ErrInvalidAccessorIdentifier = errors.New("invalid enum accessor identifier")
	// ErrAccessorCollision is returned when two enum values map to the same generated accessor.
	ErrAccessorCollision = errors.New("enum accessor identifier collision")
)

var reservedAccessorIdentifiers = reservedAccessorIdentifierSet()

func reservedAccessorIdentifierSet() map[string]struct{} {
	reserved := make(map[string]struct{}, len(ContainerMethodNames()))
	for _, name := range ContainerMethodNames() {
		reserved[name] = struct{}{}
	}
	return reserved
}

// AccessorCandidate describes one source enum value and its generated accessor.
type AccessorCandidate struct {
	Source   string
	Accessor string
}

// AccessorIdentifier returns the generated container field identifier for source.
func AccessorIdentifier(source string, style config.AccessorStyle) (string, error) {
	style = style.Normalized()
	var identifier string
	switch style {
	case config.AccessorStyleUpper:
		identifier = strings.ToUpper(source)
	case config.AccessorStyleGo:
		identifier = strings.ExportIdentifier(source)
	default:
		return "", fmt.Errorf("%w: unsupported accessor style %q", ErrInvalidAccessorIdentifier, style)
	}
	if err := validateAccessorIdentifier(source, identifier, style); err != nil {
		return "", err
	}
	return identifier, nil
}

// ValidateAccessors rejects duplicate or reserved generated accessors.
func ValidateAccessors(enumType string, candidates []AccessorCandidate, style config.AccessorStyle) error {
	owners := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		if _, reserved := reservedAccessorIdentifiers[candidate.Accessor]; reserved {
			return fmt.Errorf("%w for %s using style %q: source %q produces reserved accessor %q", ErrAccessorCollision, WrapperName(enumType), style.Normalized(), candidate.Source, candidate.Accessor)
		}
		if owner, exists := owners[candidate.Accessor]; exists && owner != candidate.Source {
			return fmt.Errorf("%w for %s using style %q: %q and %q both produce %q", ErrAccessorCollision, WrapperName(enumType), style.Normalized(), owner, candidate.Source, candidate.Accessor)
		}
		owners[candidate.Accessor] = candidate.Source
	}
	return nil
}

func validateAccessorIdentifier(source, identifier string, style config.AccessorStyle) error {
	if identifier == "" || identifier == "_" || !token.IsIdentifier(identifier) || token.Lookup(identifier).IsKeyword() || !token.IsExported(identifier) {
		return fmt.Errorf("%w: source %q with style %q produced %q", ErrInvalidAccessorIdentifier, source, style, identifier)
	}
	return nil
}

// WrapperName returns the exported wrapper type name for an enum type.
func WrapperName(enumType string) string {
	if strings.IsPlural(enumType) {
		enumType = strings.Singularise(enumType)
	}
	return strings.Camel(enumType)
}

// WrapperType returns the exported Go type used by the generated wrapper.
func WrapperType(enumType string) string {
	return strings.Camel(enumType)
}

// ContainerType returns the private generated container type name.
func ContainerType(enumType string) string {
	container := strings.Lower1stCharacter(enumType)
	container = strings.Pluralise(container)
	return container + "Container"
}

// ContainerName returns the exported package-level enum container value name.
func ContainerName(enumType string) string {
	return strings.Pluralise(strings.Camel(enumType))
}

// MatcherName returns the generated matcher interface name.
func MatcherName(enumType string) string {
	return WrapperName(enumType) + "Matcher"
}

// ContainerMethodNames returns exported method names on the generated enum container.
func ContainerMethodNames() []string {
	return []string{"All"}
}

// Receiver returns the generated method receiver name for enumType.
func Receiver(enumType string) string {
	if stdstrings.Contains(enumType, ".") {
		return stdstrings.Split(enumType, ".")[0]
	}
	if len(enumType) == 0 {
		return "r"
	}
	first, _ := utf8.DecodeRuneInString(enumType)
	return string(unicode.ToLower(first))
}

// AccessorMethodName returns the matcher method name for one source enum value.
func AccessorMethodName(source string) string {
	return strings.Camel(source)
}
