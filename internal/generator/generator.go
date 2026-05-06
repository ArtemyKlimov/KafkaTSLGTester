package generator

// Generator produces a single string value on each call.
// Implementations are NOT safe for concurrent use — each goroutine
// must use its own compiled set of generators.
type Generator interface {
	Resolve() string
}

// StaticGenerator returns the same string every time.
type StaticGenerator struct{ Value string }

func (s StaticGenerator) Resolve() string { return s.Value }

// ParseResult holds the compiled generator and optional JSON type hint.
// It implements Generator so existing call sites (Parse(...).Resolve()) continue to work.
type ParseResult struct {
	Gen       Generator
	ValueType string // "" = string, "int" = integer, "float" = floating-point, "bool" = boolean
}

func (r ParseResult) Resolve() string { return r.Gen.Resolve() }
