package generator

import (
	"regexp"
	"strconv"
	"strings"
)

var patternRe = regexp.MustCompile(`\$\(([^)]+)\)`)

// Parse returns a ParseResult for the given raw YAML field value string.
// ParseResult implements Generator, so existing callers can still call .Resolve().
// Type annotations $(int:...) and $(float:...) set ValueType for JSON numeric output.
func Parse(raw string, words []string) ParseResult {
	locs := patternRe.FindAllStringIndex(raw, -1)
	if len(locs) == 0 {
		return ParseResult{Gen: StaticGenerator{Value: raw}}
	}
	// Whole-string single pattern — preserve type annotation.
	if len(locs) == 1 && locs[0][0] == 0 && locs[0][1] == len(raw) {
		return parseSingleToken(raw, words)
	}
	return ParseResult{Gen: newTemplateGenerator(raw, words, locs)}
}

// parseSingleToken parses one full $(…) token and returns a ParseResult.
// Supports type prefixes: $(int:…) and $(float:…).
func parseSingleToken(token string, words []string) ParseResult {
	inner := token[2 : len(token)-1] // strip $( and )

	var valueType string
	switch {
	case strings.HasPrefix(inner, "int:"):
		valueType = "int"
		inner = inner[4:]
	case strings.HasPrefix(inner, "float:"):
		valueType = "float"
		inner = inner[6:]
	}

	return ParseResult{Gen: parseGenerator(inner, words), ValueType: valueType}
}

// parseGenerator maps an inner expression (without $( )) to a Generator.
func parseGenerator(inner string, words []string) Generator {
	switch {
	case inner == "CURRENT_TIMESTAMP":
		return &TimestampGenerator{precision: 0}
	case strings.HasPrefix(inner, "CURRENT_TIMESTAMP:"):
		p, _ := strconv.Atoi(inner[len("CURRENT_TIMESTAMP:"):])
		return &TimestampGenerator{precision: p}
	case strings.HasPrefix(inner, "one_of:"):
		return NewOneOfGenerator(inner[7:])
	case strings.HasPrefix(inner, "num:"):
		return NewNumGenerator(inner[4:])
	case strings.HasPrefix(inner, "random_text_"):
		n, _ := strconv.Atoi(inner[len("random_text_"):])
		return NewRandomTextGenerator(n)
	case strings.HasPrefix(inner, "random_word_"):
		n, _ := strconv.Atoi(inner[len("random_word_"):])
		return NewRandomWordGenerator(n, words)
	case inner == "uuid":
		return &UUIDGenerator{}
	case strings.HasPrefix(inner, "random_hex_"):
		n, _ := strconv.Atoi(inner[len("random_hex_"):])
		return NewHexGenerator(n)
	case inner == "random_ip":
		return &IPGenerator{rng: newRand()}
	default:
		return StaticGenerator{Value: "$(" + inner + ")"}
	}
}

type templatePart struct {
	static string
	gen    Generator
}

// TemplateGenerator handles strings with mixed static text and $(…) tokens.
// Each embedded generator is created once at Compile time.
type TemplateGenerator struct {
	parts []templatePart
}

func newTemplateGenerator(raw string, words []string, locs [][]int) *TemplateGenerator {
	parts := make([]templatePart, 0, len(locs)*2+1)
	last := 0
	for _, loc := range locs {
		if loc[0] > last {
			parts = append(parts, templatePart{static: raw[last:loc[0]]})
		}
		parts = append(parts, templatePart{gen: parseSingleToken(raw[loc[0]:loc[1]], words).Gen})
		last = loc[1]
	}
	if last < len(raw) {
		parts = append(parts, templatePart{static: raw[last:]})
	}
	return &TemplateGenerator{parts: parts}
}

func (t *TemplateGenerator) Resolve() string {
	var sb strings.Builder
	for _, p := range t.parts {
		if p.gen != nil {
			sb.WriteString(p.gen.Resolve())
		} else {
			sb.WriteString(p.static)
		}
	}
	return sb.String()
}
