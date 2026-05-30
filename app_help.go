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

func (c *app) getGlobalOptions(osArgs []string) (map[string]any, map[string]any) {
	// c.globalOptions is always a non-nil *GlobalOptions (set once in NewApp),
	// so GetStructFields cannot return an error here.
	globalFields, _ := structs.GetStructFields(c.globalOptions, nil, structs.DefaultEncodingTags)

	_, _, globalOptions, unknownOptions := getCommandArgs(osArgs, globalFields)

	return globalOptions, unknownOptions
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
// GlobalOptions.Format) followed by the names of any codecs registered in
// Config.Formats, derived from each Extension (".yaml" -> "yaml"), without duplicates.
func (c *app) allowedFormats() []string {
	allowed := builtinFormatValues()
	for _, codec := range c.config.Formats {
		name := strings.TrimPrefix(codec.Extension(), ".")
		if name != "" && !slices.Contains(allowed, name) {
			allowed = append(allowed, name)
		}
	}
	return allowed
}

// builtinFormatValues reads the built-in --format values from the oneof rule on
// GlobalOptions.Format, keeping that struct tag the single source of truth.
func builtinFormatValues() []string {
	fields, err := structs.GetStructFields(&GlobalOptions{}, nil, defaultTags)
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
			err := cmd.Run(*c.globalOptions, Unknowns{
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
