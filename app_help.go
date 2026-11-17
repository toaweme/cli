package cli

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/toaweme/structs"
)

// Help registers cmd as the help command under the reserved "help" name,
// so callers never type that name themselves. Use it instead of Add. It returns cmd.
func (c *app) Help(cmd Command[any]) Command[any] {
	return c.Add(helpCommand, cmd)
}

// validateFormat checks a --help-format value against the full allowed set
// (built-ins plus registered output codecs). An absent or empty value is fine (default help).
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
	return fmt.Errorf("invalid --help-format %q: must be one of %s", name, strings.Join(allowed, ", "))
}

// allowedFormats is the built-in --help-format values (from the oneof rule on
// GlobalFlags.HelpFormat) followed by every name the registered output codecs answer to
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
// more than one --help-format name (e.g. a YAML codec: "yml" and "yaml").
type outputExtensions interface {
	Extensions() []string
}

// FormatAliases returns every --help-format name a codec answers to: each extension it
// reports (Extensions() when implemented, otherwise its Extension()), with the leading
// dot trimmed and empties dropped. The first is the primary, used for the help hint and
// for writing; the rest are accepted aliases.
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

// builtinFormatValues reads the built-in --help-format values from the oneof rule on
// GlobalFlags.HelpFormat, keeping that struct tag the single source of truth.
func builtinFormatValues() []string {
	fields, err := structs.GetStructFields(&GlobalFlags{}, nil, defaultTags)
	if err != nil {
		return nil
	}
	for _, field := range fields {
		if field.Tags[tagArg] != argHelpFormat {
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

func (c *app) runHelp(args []string) error {
	for _, cmd := range c.commands {
		if cmd.Name("") == helpCommand {
			if help, ok := cmd.(HelpCommand); ok {
				help.SetPath(args)
			}
			err := cmd.Run(*c.globalFlags)
			if err != nil {
				return fmt.Errorf("failed to run help command: %w", err)
			}
			return nil
		}
	}

	return errors.New("help command not found")
}
