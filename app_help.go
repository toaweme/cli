package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/toaweme/structs"
)

// Help registers cmd as the help command under the reserved "help" name, so callers
// never type that name themselves. Use it instead of Add. It returns cmd.
func (c *app) Help(cmd Command[any]) Command[any] {
	return c.Add(helpCommand, cmd)
}

func (c *app) getGlobalFlags(osArgs []string) (map[string]any, map[string]any) {
	// c.globalFlags is always a non-nil *GlobalFlags (set once in NewApp),
	// so GetStructFields cannot return an error here.
	globalFields, _ := structs.GetStructFields(c.globalFlags, nil, structs.DefaultEncodingTags)

	_, _, globalFlags, unknownOptions := getCommandArgs(osArgs, globalFields)

	return globalFlags, unknownOptions
}

// boolFlagRequested reports whether any of names (long or short, without dashes)
// appears as a flag token anywhere in osArgs, independent of the value-consuming
// parser. The parser lets a value-taking flag (a command's own "--target", or any
// flag unknown to the global scan) swallow a following "--help" as its value, so a
// direct scan is what makes built-in bool flags like -h/--help and -v/--version
// trigger no matter where they sit. Scanning stops at a "--" terminator, and the
// explicit "--flag=false" form does not count as set.
func boolFlagRequested(osArgs []string, names []string) bool {
	for _, raw := range osArgs {
		if raw == "--" {
			break
		}
		if !strings.HasPrefix(raw, optionPrefix) {
			continue
		}
		name, value := splitKeyValue(strings.TrimLeft(raw, optionPrefix))
		if !slices.Contains(names, name) {
			continue
		}
		if strings.Contains(raw, "=") && !truthy(value) {
			continue
		}
		return true
	}
	return false
}

// globalBoolFlagNames returns the long and short spellings of the GlobalFlags field
// tagged arg:"<arg>" (e.g. "help" -> ["help", "h"]), keeping the struct tags the
// single source of truth for the built-in flag names.
func globalBoolFlagNames(arg string) []string {
	names := []string{arg}
	fields, err := structs.GetStructFields(&GlobalFlags{}, nil, defaultTags)
	if err != nil {
		return names
	}
	for _, field := range fields {
		if field.Tags[tagArg] != arg {
			continue
		}
		if short := field.Tags[tagShort]; short != "" {
			names = append(names, short)
		}
		break
	}
	return names
}

// truthy reports whether v (the right side of a "--flag=value" token) reads as
// true. An empty value and any non-false word count as true, so only the explicit
// false spellings suppress a bool flag.
func truthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "false", "0", "no", "n", "f", "off":
		return false
	}
	return true
}

// validateFormat checks a --format value against the full allowed set (built-ins
// plus registered output codecs). An absent or empty value is fine (default help).
// The error lists every accepted format, including the registered ones the static
// oneof rule cannot know about.
func (c *app) validateFormat(value any) error {
	name, ok := value.(string)
	if !ok || name == "" {
		return nil
	}
	allowed := c.allowedFormats()
	if slices.Contains(allowed, name) {
		return nil
	}
	return fmt.Errorf("invalid --format %q: must be one of %s", name, strings.Join(allowed, ", "))
}

// allowedFormats is the built-in --format values (from the oneof rule on
// GlobalFlags.Format) followed by every name the registered output codecs answer to
// (each of their FormatAliases), without duplicates.
func (c *app) allowedFormats() []string {
	allowed := builtinFormatValues()
	for _, codec := range c.formats {
		for _, name := range FormatAliases(codec) {
			if !slices.Contains(allowed, name) {
				allowed = append(allowed, name)
			}
		}
	}
	return allowed
}

// outputExtensions is the optional interface an OutputCodec implements to answer to
// more than one --format name (e.g. a YAML codec: "yml" and "yaml").
type outputExtensions interface {
	Extensions() []string
}

// FormatAliases returns every --format name a codec answers to: each extension it
// reports (Extensions() when implemented, otherwise its Extension()), with the
// leading dot trimmed and empties dropped. The first is the primary, used for the
// help hint and for writing; the rest are accepted aliases.
func FormatAliases(codec OutputCodec) []string {
	exts := []string{codec.Extension()}
	if oe, ok := codec.(outputExtensions); ok {
		if reported := oe.Extensions(); len(reported) > 0 {
			exts = reported
		}
	}
	names := make([]string, 0, len(exts))
	for _, ext := range exts {
		if name := strings.TrimPrefix(ext, "."); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// builtinFormatValues reads the built-in --format values from the oneof rule on
// GlobalFlags.Format, keeping that struct tag the single source of truth.
func builtinFormatValues() []string {
	fields, err := structs.GetStructFields(&GlobalFlags{}, nil, defaultTags)
	if err != nil {
		return nil
	}
	for _, field := range fields {
		if field.Tags["arg"] != "format" {
			continue
		}
		for _, rule := range field.Rules {
			if rule.Name == "oneof" {
				return rule.Args
			}
		}
	}
	return nil
}

func (c *app) printVersion() {
	fmt.Printf("%s %s\n", c.config.Name, c.config.Version)
}

func (c *app) runHelp(args []string, opts ...map[string]any) error {
	options := map[string]any{}
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	for _, cmd := range c.commands {
		if cmd.Name("") == helpCommand {
			err := cmd.Run(*c.globalFlags, Unknowns{
				Args:    args,
				Options: options,
			})
			if err != nil {
				return fmt.Errorf("failed to run help command: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("help command not found")
}
