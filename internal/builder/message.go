package builder

import (
	"encoding/json"
	"fmt"
	"strconv"

	"kafkatsgltest/internal/generator"
)

// FieldSpec is the compiled representation of a block's fields tree.
// Must NOT be shared across goroutines — call Compile per worker goroutine.
type FieldSpec map[string]fieldValue

type fieldValue struct {
	gen       generator.Generator
	nested    FieldSpec
	valueType string // "", "int", "float", "bool"
}

// Compile converts the raw map[string]any from YAML into a FieldSpec.
// Recurses into nested maps; calls generator.Parse on string leaves.
// Native YAML int/float/bool values are preserved as their JSON numeric/boolean types.
func Compile(raw map[string]any, words []string) (FieldSpec, error) {
	spec := make(FieldSpec, len(raw))
	for k, v := range raw {
		switch val := v.(type) {
		case string:
			result := generator.Parse(val, words)
			spec[k] = fieldValue{gen: result.Gen, valueType: result.ValueType}
		case map[string]any:
			nested, err := Compile(val, words)
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", k, err)
			}
			spec[k] = fieldValue{nested: nested}
		case int:
			spec[k] = fieldValue{gen: generator.StaticGenerator{Value: strconv.Itoa(val)}, valueType: "int"}
		case int64:
			spec[k] = fieldValue{gen: generator.StaticGenerator{Value: strconv.FormatInt(val, 10)}, valueType: "int"}
		case float64:
			spec[k] = fieldValue{gen: generator.StaticGenerator{Value: strconv.FormatFloat(val, 'f', -1, 64)}, valueType: "float"}
		case float32:
			spec[k] = fieldValue{gen: generator.StaticGenerator{Value: strconv.FormatFloat(float64(val), 'f', -1, 32)}, valueType: "float"}
		case bool:
			spec[k] = fieldValue{gen: generator.StaticGenerator{Value: strconv.FormatBool(val)}, valueType: "bool"}
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
			continue
		}
		s := fv.gen.Resolve()
		switch fv.valueType {
		case "int":
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				out[k] = n
			} else {
				out[k] = s
			}
		case "float":
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				out[k] = f
			} else {
				out[k] = s
			}
		case "bool":
			if b, err := strconv.ParseBool(s); err == nil {
				out[k] = b
			} else {
				out[k] = s
			}
		default:
			out[k] = s
		}
	}
	return out
}

// Build generates one JSON-encoded message from a FieldSpec.
func Build(spec FieldSpec) ([]byte, error) {
	return json.Marshal(resolve(spec))
}

// formatScalar конвертирует нестроковые YAML-значения в строку без потери точности.
// float64 передаётся в десятичной форме (не в экспоненциальной).
func formatScalar(v any) string {
	switch val := v.(type) {
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
