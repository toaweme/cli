package help

import (
	"fmt"
	"strings"

	"github.com/toaweme/structs"
)

// fieldRawValue is field's resolved value as a bare string for --help-values. A zero
// or empty value yields "" so unset flags are left unannotated (no noisy "0"/"false").
// Only fields tagged secret:"true" are prefix-masked; everything else shows its real
// value, since masking a port or a boolean helps no one. This is the structured form
// used by the json/jsonschema renderers; text renderers wrap it via valueText.
func fieldRawValue(field structs.Field) string {
	v := field.Value
	if !v.IsValid() || !v.CanInterface() || v.IsZero() {
		return ""
	}
	s := fmt.Sprintf("%v", v.Interface())
	if s == "" {
		return ""
	}
	if isSecretField(field) {
		return maskValue(s)
	}
	return s
}

// valueText is the resolved value as shown in help: secrets masked and path-like
// values shortened to their last segments, with no surrounding brackets or quotes.
// Returns "" for an unset value. The raw, unshortened form is fieldRawValue, used by
// the json/jsonschema renderers.
func valueText(field structs.Field) string {
	raw := fieldRawValue(field)
	if raw == "" {
		return ""
	}
	return shortenPath(raw)
}

// valueColumn is the `<type> <value>` cell shown in the compact help listing (which
// has no separate type column), e.g. `int 8`. Empty for an unset value.
func valueColumn(field structs.Field) string {
	v := valueText(field)
	if v == "" {
		return ""
	}
	return displayType(field) + " " + v
}

// shortenPath collapses a path-like value to its last two segments (e.g. a long
// working directory to `…/toaweme/cli`), so resolved paths do not blow out the help
// column. Non-path values and already-short paths are returned unchanged.
func shortenPath(s string) string {
	if !strings.Contains(s, "/") {
		return s
	}
	var parts []string
	for _, p := range strings.Split(s, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) <= 2 {
		return s
	}
	return "…/" + strings.Join(parts[len(parts)-2:], "/")
}

// isSecretField reports whether a field is marked sensitive via secret:"true", so
// its resolved value is masked in help output.
func isSecretField(field structs.Field) bool {
	switch strings.ToLower(strings.TrimSpace(field.Tags["secret"])) {
	case "true", "1", "yes", "y", "on":
		return true
	}
	return false
}

// maskValue reveals a short prefix of v and masks the rest, so secret values shown
// in help - which may be pulled from env or a .env file - never leak in full to
// logs, screenshots, or pasted issues. A single-rune value is shown as is (nothing
// meaningful to hide).
func maskValue(v string) string {
	runes := []rune(v)
	n := len(runes)
	if n <= 1 {
		return v
	}
	reveal := 3
	if reveal > n-1 {
		reveal = n - 1
	}
	return string(runes[:reveal]) + strings.Repeat("•", n-reveal)
}
