package builder

import (
	"encoding/json"
	"testing"
)

var testWords = []string{"account", "select", "message", "error"}

func TestBuildFlat(t *testing.T) {
	raw := map[string]any{
		"projectCode": "TSLG",
		"traceId":     "$(uuid)",
		"level":       "$(one_of:INFO,ERROR)",
		"num":         "$(num:1to9)",
	}
	spec, err := Compile(raw, testWords)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	payload, err := Build(spec)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(payload, &out); err != nil {
		t.Fatalf("unmarshal: %v, json: %s", err, payload)
	}
	if out["projectCode"] != "TSLG" {
		t.Fatalf("expected TSLG, got %v", out["projectCode"])
	}
	if _, ok := out["traceId"].(string); !ok {
		t.Fatalf("traceId is not a string: %v", out["traceId"])
	}
}

func TestBuildNested(t *testing.T) {
	raw := map[string]any{
		"outer": "static",
		"tslgMdc": map[string]any{
			"requestId": "$(uuid)",
			"userId":    "$(num:1000to9999)",
		},
	}
	spec, err := Compile(raw, testWords)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	payload, err := Build(spec)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(payload, &out); err != nil {
		t.Fatalf("unmarshal: %v, json: %s", err, payload)
	}

	nested, ok := out["tslgMdc"].(map[string]any)
	if !ok {
		t.Fatalf("tslgMdc is not an object: %T %v", out["tslgMdc"], out["tslgMdc"])
	}
	if _, ok := nested["requestId"].(string); !ok {
		t.Fatalf("requestId not a string: %v", nested["requestId"])
	}
}

func TestBuildIndependentMessages(t *testing.T) {
	raw := map[string]any{"id": "$(uuid)"}
	spec, _ := Compile(raw, testWords)

	ids := make(map[string]bool)
	for range 20 {
		payload, _ := Build(spec)
		var out map[string]any
		_ = json.Unmarshal(payload, &out)
		ids[out["id"].(string)] = true
	}
	if len(ids) < 18 {
		t.Fatalf("UUIDs not unique enough: only %d distinct out of 20", len(ids))
	}
}
