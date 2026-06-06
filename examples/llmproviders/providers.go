// Package llmproviders shows how goenums handles a camelCase enum whose
// leading word is a known initialism. The source type llmProvider exports
// as LLMProvider (not LlmProvider) so generated code passes revive and
// staticcheck var-naming. The same rule applies to apiKey -> APIKey,
// httpStatus -> HTTPStatus, jsonField -> JSONField, etc.
package llmproviders

//go:generate goenums -f providers.go

type llmProvider int // Vendor string, ContextWindow int, SupportsTools bool

const (
	openai    llmProvider = iota + 1 // "OpenAI",128000,true
	anthropic                        // "Anthropic",200000,true
	google                           // "Google",1000000,true
	meta                             // "Meta",128000,false
	mistral                          // "Mistral",32000,true
)
