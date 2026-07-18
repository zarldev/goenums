package gofile

import "github.com/zarldev/goenums/enum"

type renderRequest struct {
	enum.GenerationRequest
	AllEnums   []enumDefinition
	ValidEnums []enumDefinition
	ParseEnums []enumDefinition
}

func newRenderRequest(req enum.GenerationRequest) renderRequest {
	allEnums := buildEnumDefinitions(req)
	validEnums := filterValidEnumDefinitions(allEnums)
	return renderRequest{
		GenerationRequest: req,
		AllEnums:          allEnums,
		ValidEnums:        validEnums,
		ParseEnums:        buildParseEnumDefinitions(allEnums, req.Configuration.Insensitive),
	}
}

func filterValidEnumDefinitions(defs []enumDefinition) []enumDefinition {
	valid := make([]enumDefinition, 0, len(defs))
	for _, def := range defs {
		if def.Valid {
			valid = append(valid, def)
		}
	}
	return valid
}

func buildParseEnumDefinitions(defs []enumDefinition, insensitive bool) []enumDefinition {
	parseDefs := make([]enumDefinition, len(defs))
	copy(parseDefs, defs)
	if !insensitive {
		return parseDefs
	}
	for i := range parseDefs {
		parseDefs[i].Aliases = lowercaseAliases(parseDefs[i].Aliases)
		parseDefs[i].QuotedAliases = quotedAliases(parseDefs[i].Aliases)
	}
	return parseDefs
}

func estimateGeneratedSize(req renderRequest) int {
	enumCount := len(req.AllEnums)
	aliasCount := 0
	fieldCount := 0
	for _, enumDef := range req.AllEnums {
		aliasCount += len(enumDef.Aliases)
		fieldCount += len(enumDef.Fields)
	}

	size := 8 * 1024
	size += enumCount * 512
	size += aliasCount * 64
	size += fieldCount * 96

	if req.Configuration.Handlers.JSON {
		size += 2 * 1024
	}
	if req.Configuration.Handlers.Text {
		size += 2 * 1024
	}
	if req.Configuration.Handlers.SQL {
		size += 2 * 1024
	}
	if req.Configuration.Handlers.YAML {
		size += 2 * 1024
	}
	if req.Configuration.Handlers.Binary {
		size += 2 * 1024
	}

	return size
}
