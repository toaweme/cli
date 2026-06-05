package help

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/toaweme/structs"

	"github.com/toaweme/cli"
)

// isPositionalArg reports whether an arg tag names a positional argument (a bare
// index like "0"), as opposed to a named flag.
func isPositionalArg(arg string) bool {
	if arg == "" {
		return false
	}
	for _, c := range arg {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// commandDescription returns the command's long-form description with trailing
// newlines trimmed.
func commandDescription(cmd cli.Command[any]) string {
	return strings.TrimRight(cmd.Description(), "\n")
}

// firstLine returns the first line of s, used to keep listing columns aligned
// even if a command's Help summary accidentally spans multiple lines.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// hasRule checks whether a struct field has a specific validation rule (e.g. "required").
func hasRule(field structs.Field, name string) bool {
	for _, r := range field.Rules {
		if r.Name == name {
			return true
		}
	}
	return false
}

// oneOfValues returns the allowed values from a field's `oneof` rule, or nil.
func oneOfValues(field structs.Field) []string {
	for _, r := range field.Rules {
		if r.Name == "oneof" {
			return r.Args
		}
	}
	return nil
}

// withAllowedValues appends a "(one of: ...)" suffix to help text when the field
// carries a oneof rule, so listings and tables show the permitted values. extra
// adds dynamically discovered values (e.g. output codecs registered for --help-format)
// after the static ones, skipping duplicates.
func withAllowedValues(help string, field structs.Field, extra []string) string {
	vals := oneOfValues(field)
	seen := make(map[string]bool, len(vals))
	for _, v := range vals {
		seen[v] = true
	}
	for _, v := range extra {
		if !seen[v] {
			vals = append(vals, v)
			seen[v] = true
		}
	}
	if len(vals) == 0 {
		return help
	}
	suffix := fmt.Sprintf("(one of: %s)", strings.Join(vals, ", "))
	if help == "" {
		return suffix
	}
	return help + " " + suffix
}

// formatHintExtras returns extraFormats only for the global --help-format field, so the
// dynamic format names ride along on that flag's allowed-values hint and nowhere else.
func formatHintExtras(field structs.Field, extraFormats []string) []string {
	if len(extraFormats) > 0 && field.Tags["arg"] == "help-format" {
		return extraFormats
	}
	return nil
}

// displayType renders a field's type for help output, preferring the concrete Go
// type for slices ("[]string") over the bare reflect kind ("slice").
func displayType(field structs.Field) string {
	if field.Value.IsValid() && field.Value.Kind() == reflect.Slice {
		return field.Value.Type().String()
	}
	return field.Type
}

// globalSource returns the global flags struct to render the Global Options block
// from: the populated one when provided (so --help-values shows set values), else a
// zero struct (just the flag definitions).
func globalSource(globalValues *cli.GlobalFlags) any {
	if globalValues == nil {
		return &cli.GlobalFlags{}
	}
	return globalValues
}
