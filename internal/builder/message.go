package builder

import (
	"encoding/json"
	"fmt"

	"kafkatsgltest/internal/generator"
)

// FieldSpec is the compiled representation of a block's fields tree.
// Must NOT be shared across goroutines — call Compile per worker goroutine.
type FieldSpec map[string]fieldValue

type fieldValue struct {
	gen    generator.Generator
	nested FieldSpec
}

// Compile converts the raw map[string]any from YAML into a FieldSpec.
// Recurses into nested maps; calls generator.Parse on string leaves.
func Compile(raw map[string]any, words []string) (FieldSpec, error) {
	spec := make(FieldSpec, len(raw))
	for k, v := range raw {
		switch val := v.(type) {
		case string:
			spec[k] = fieldValue{gen: generator.Parse(val, words)}
		case map[string]any:
			nested, err := Compile(val, words)
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", k, err)
			}
			spec[k] = fieldValue{nested: nested}
		default:
			spec[k] = fieldValue{gen: generator.StaticGenerator{Value: fmt.Sprintf("%v", val)}}
		}
	}
	return spec, nil
}

func resolve(spec FieldSpec) map[string]any {
	out := make(map[string]any, len(spec))
	for k, fv := range spec {
		if fv.nested != nil {
			out[k] = resolve(fv.nested)
		} else {
			out[k] = fv.gen.Resolve()
		}
	}
	return out
}

// Build generates one JSON-encoded message from a FieldSpec.
func Build(spec FieldSpec) ([]byte, error) {
	return json.Marshal(resolve(spec))
}
