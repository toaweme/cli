package help

import (
	"strings"
	"testing"

	"github.com/toaweme/structs"
)

func Test_fieldValue(t *testing.T) {
	type cfg struct {
		Port  int    `arg:"port"`
		Host  string `arg:"host"`
		Dry   bool   `arg:"dry"`
		Empty string `arg:"empty"`
		Path  string `arg:"path"`
		Token string `arg:"token" secret:"true"`
	}

	c := &cfg{Port: 8080, Host: "localhost", Dry: false, Empty: "", Path: "/Users/me/code/toaweme/cli", Token: "sk-live-abcdef"}
	fields, err := structs.GetStructFields(c, nil, structs.DefaultEncodingTags)
	if err != nil {
		t.Fatalf("GetStructFields: %v", err)
	}

	raw := map[string]string{}
	text := map[string]string{}
	for _, f := range fields {
		raw[f.Tags["arg"]] = fieldRawValue(f)
		text[f.Tags["arg"]] = valueText(f)
	}

	// raw form (used by json/jsonschema): full value, secret masked, zero skipped.
	if raw["port"] != "8080" {
		t.Errorf("port raw: want 8080, got %q", raw["port"])
	}
	if raw["path"] != "/Users/me/code/toaweme/cli" {
		t.Errorf("path raw: want full path, got %q", raw["path"])
	}
	if raw["dry"] != "" {
		t.Errorf("dry raw (zero bool): want unannotated, got %q", raw["dry"])
	}
	if raw["token"] == "sk-live-abcdef" || !strings.HasPrefix(raw["token"], "sk-") {
		t.Errorf("token raw (secret): want masked with sk- prefix, got %q", raw["token"])
	}

	// display form: bare value (no brackets/quotes), path shortened, zero skipped.
	if text["port"] != "8080" {
		t.Errorf("port text: want 8080, got %q", text["port"])
	}
	if text["host"] != "localhost" {
		t.Errorf("host text: want localhost, got %q", text["host"])
	}
	if text["path"] != "…/toaweme/cli" {
		t.Errorf("path text: want shortened, got %q", text["path"])
	}
	if text["dry"] != "" || text["empty"] != "" {
		t.Errorf("zero values should be unannotated, got dry=%q empty=%q", text["dry"], text["empty"])
	}
	if text["token"] == "sk-live-abcdef" || !strings.HasPrefix(text["token"], "sk-") {
		t.Errorf("token text (secret): want masked, got %q", text["token"])
	}
}

func Test_printableFields_ValueBeforeDescription(t *testing.T) {
	type cfg struct {
		Steps int `arg:"steps" short:"n" help:"Number of migrations to run"`
	}
	fields, err := structs.GetStructFields(&cfg{Steps: 8}, nil, structs.DefaultEncodingTags)
	if err != nil {
		t.Fatalf("GetStructFields: %v", err)
	}

	lines := printableFieldsWithEnv(fields, false, true, nil)
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d: %v", len(lines), lines)
	}
	line := lines[0]
	vi := strings.Index(line, "int 8")
	hi := strings.Index(line, "Number of migrations to run")
	if vi < 0 || hi < 0 || vi > hi {
		t.Errorf("value should sit in a column before the description, got %q", line)
	}
}

func Test_valueColCell_DimsInPretty(t *testing.T) {
	r := flagRow{Type: "int", Value: "8080"}

	if got := valueColCell(r, false); got != "8080" {
		t.Errorf("plain value cell = %q", got)
	}
	if got := valueColCell(r, true); got != "*8080*" {
		t.Errorf("markdown value cell = %q", got)
	}

	// the markdown emphasis is what the pretty renderer turns into dim ANSI.
	if pretty := prettyInline(valueColCell(r, true)); !strings.Contains(pretty, ansiDim) {
		t.Errorf("pretty render should dim the value, got %q", pretty)
	}

	if got := valueColCell(flagRow{Value: ""}, true); got != "" {
		t.Errorf("a row with no value should have an empty cell, got %q", got)
	}

	// the type column carries only the type (plus a required marker);
	// the default reads as a trailing "(default: x)" hint on the description instead.
	if got := typeCol(flagRow{Type: "int", Default: "5"}); got != "int" {
		t.Errorf("type col should not carry the default, got %q", got)
	}
	if got := typeCol(flagRow{Type: "int", Required: true}); got != "int, required" {
		t.Errorf("type col with required = %q", got)
	}
}

func Test_descCol_DefaultHint(t *testing.T) {
	// a non-zero default trails the description as a "(default: x)" hint.
	if got := descCol(flagRow{Help: "Steps to run", Type: "int", Default: "5"}, false); got != "Steps to run (default: 5)" {
		t.Errorf("plain desc with default = %q", got)
	}
	// markdown emphasizes the hint so the pretty renderer dims it.
	if got := descCol(flagRow{Help: "Steps to run", Type: "int", Default: "5"}, true); got != "Steps to run *(default: 5)*" {
		t.Errorf("markdown desc with default = %q", got)
	}
	// a bool defaulting to false (the zero value) shows no hint - it is implied.
	if got := descCol(flagRow{Help: "Output as JSON", Type: "bool", Default: "false"}, false); got != "Output as JSON" {
		t.Errorf("zero-value bool default should be suppressed, got %q", got)
	}
	// likewise a numeric zero default.
	if got := descCol(flagRow{Help: "Retries", Type: "int", Default: "0"}, false); got != "Retries" {
		t.Errorf("zero-value int default should be suppressed, got %q", got)
	}
	// an env var column is just the name, no "=default".
	if got := envColValue(flagRow{Env: "MEND_JSON", Default: "false"}); got != "`MEND_JSON`" {
		t.Errorf("env col should be the bare name, got %q", got)
	}
}

func Test_maskValue(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"single rune shown as is", "5", "5"},
		{"short value reveals all but last", "ab", "a•"},
		{"reveals first three then masks", "8080", "808•"},
		{"longer value", "0.0.0.0", "0.0••••"},
		{"secret-like value mostly masked", "sk-live-abcdef", "sk-•••••••••••"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskValue(tt.in)
			if got != tt.want {
				t.Fatalf("maskValue(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func Test_maskValue_neverContainsFullSecret(t *testing.T) {
	secret := "super-secret-token-value"
	got := maskValue(secret)
	if got == secret {
		t.Fatalf("maskValue must not return the value unmasked: %q", got)
	}
	// the revealed prefix is short and the rest is masked.
	if len([]rune(got)) != len([]rune(secret)) {
		t.Fatalf("masked value should preserve length, got %d want %d", len([]rune(got)), len([]rune(secret)))
	}
}
