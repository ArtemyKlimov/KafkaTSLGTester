package generator

import (
	"regexp"
	"testing"
)

var words = []string{"account", "select", "message", "error", "kafka"}

func TestStaticValue(t *testing.T) {
	g := Parse("TSLG", words)
	if g.Resolve() != "TSLG" {
		t.Fatalf("expected TSLG, got %s", g.Resolve())
	}
}

func TestTimestampPrecision(t *testing.T) {
	cases := map[string]*regexp.Regexp{
		"$(CURRENT_TIMESTAMP)":   regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`),
		"$(CURRENT_TIMESTAMP:3)": regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`),
		"$(CURRENT_TIMESTAMP:6)": regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z$`),
		"$(CURRENT_TIMESTAMP:9)": regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z$`),
	}
	for pattern, re := range cases {
		val := Parse(pattern, words).Resolve()
		if !re.MatchString(val) {
			t.Errorf("pattern %q: %q did not match %s", pattern, val, re)
		}
	}
}

func TestOneOf(t *testing.T) {
	allowed := map[string]bool{"INFO": true, "ERROR": true, "WARNING": true, "DEBUG": true}
	g := Parse("$(one_of:INFO,ERROR,WARNING,DEBUG)", words)
	for i := 0; i < 50; i++ {
		v := g.Resolve()
		if !allowed[v] {
			t.Fatalf("unexpected value %q", v)
		}
	}
}

func TestNum(t *testing.T) {
	g := Parse("$(num:0to500)", words)
	re := regexp.MustCompile(`^\d+$`)
	for i := 0; i < 50; i++ {
		v := g.Resolve()
		if !re.MatchString(v) {
			t.Fatalf("expected digits, got %q", v)
		}
	}
}

func TestUUID(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	g := Parse("$(uuid)", words)
	for i := 0; i < 10; i++ {
		v := g.Resolve()
		if !re.MatchString(v) {
			t.Fatalf("invalid UUID: %q", v)
		}
	}
}

func TestRandomHex(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{16}$`)
	g := Parse("$(random_hex_16)", words)
	for i := 0; i < 10; i++ {
		v := g.Resolve()
		if !re.MatchString(v) {
			t.Fatalf("invalid hex: %q", v)
		}
	}
}

func TestRandomIP(t *testing.T) {
	re := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	g := Parse("$(random_ip)", words)
	for i := 0; i < 10; i++ {
		if !re.MatchString(g.Resolve()) {
			t.Fatal("invalid IP")
		}
	}
}

func TestRandomText(t *testing.T) {
	g := Parse("$(random_text_20)", words)
	for i := 0; i < 10; i++ {
		v := g.Resolve()
		if len(v) != 20 {
			t.Fatalf("expected 20 chars, got %d: %q", len(v), v)
		}
	}
}

func TestRandomWord(t *testing.T) {
	g := Parse("$(random_word_15)", words)
	for i := 0; i < 10; i++ {
		v := g.Resolve()
		if len(v) != 15 {
			t.Fatalf("expected 15 chars, got %d: %q", len(v), v)
		}
	}
}

func TestTemplate(t *testing.T) {
	g := Parse("user-$(num:1to99)-$(uuid)", words)
	re := regexp.MustCompile(`^user-\d+-[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	for i := 0; i < 10; i++ {
		v := g.Resolve()
		if !re.MatchString(v) {
			t.Fatalf("template mismatch: %q", v)
		}
	}
}
