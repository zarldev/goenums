package gofile

import (
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/naming"
)

var (
	// ErrInvalidAccessorIdentifier is returned when an enum accessor cannot be a valid exported Go identifier.
	ErrInvalidAccessorIdentifier = naming.ErrInvalidAccessorIdentifier
	// ErrAccessorCollision is returned when two enum values map to the same generated accessor.
	ErrAccessorCollision = naming.ErrAccessorCollision
)

func enumAccessorIdentifier(source string, style config.AccessorStyle) (string, error) {
	return naming.AccessorIdentifier(source, style)
}

func validateAccessorCollisions(enumType string, defs []enumDefinition, style config.AccessorStyle) error {
	candidates := make([]naming.AccessorCandidate, len(defs))
	for i, def := range defs {
		candidates[i] = naming.AccessorCandidate{
			Source:   def.EnumName,
			Accessor: def.EnumNameIdentifier,
		}
	}
	return naming.ValidateAccessors(enumType, candidates, style)
}
func buildContainerEnums(defs []enumDefinition) []cenum {
	cenums := make([]cenum, len(defs))
	for i, e := range defs {
		cenums[i] = cenum{
			Name:     e.EnumNameIdentifier,
			EnumType: e.EnumType,
		}
	}
	return cenums
}
