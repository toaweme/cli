// Package help provides a command that displays usage information for the application.
package help

import (
	"fmt"
	"os"

	"github.com/toaweme/cli"
	clihelp "github.com/toaweme/cli/help"
)

// Config holds the inputs for the help command.
type Config struct{}

// Command displays usage information for the application or a specific command.
type Command struct {
	cli.BaseCommand[Config]

	settingsFunc    func() cli.Config
	commandListFunc func() []cli.Command[any]
	formatsFunc     func() []cli.OutputCodec
	defaultFunc     func() cli.Command[any]
}

var _ cli.Command[Config] = (*Command)(nil)

// NewHelpCommand creates a help command that lists all available commands.
// The formats getter (typically App.OutputFormats) supplies the codecs registered via App.HelpOutputs
// so the help renderer can advertise and apply custom --help-format values. The defaultCmd getter
// (typically App.DefaultCommand) lets the renderer flag which command runs on a bare invocation.
func NewHelpCommand(settingsFunc func() cli.Config, commandList func() []cli.Command[any], formats func() []cli.OutputCodec, defaultCmd func() cli.Command[any]) *Command {
	return &Command{settingsFunc: settingsFunc, commandListFunc: commandList, formatsFunc: formats, defaultFunc: defaultCmd}
}

// Run renders help output in the requested format for the app or a filtered command.
func (c *Command) Run(options cli.GlobalFlags, unknowns cli.Unknowns) error {
	cfg := c.settingsFunc()
	commands := c.commandListFunc()
	appName := cfg.Name

	var defaultName string
	if c.defaultFunc != nil {
		if def := c.defaultFunc(); def != nil {
			defaultName = def.Name("")
		}
	}

	format := options.HelpFormat

	codecs := c.formatsFunc()

	// output codecs registered on the app (yaml, toml, ...), keyed by every --help-format name
	// they answer to (e.g. both "yml" and "yaml"). formatNames lists only each codec's primary name,
	// in registration order, for the help hint.
	customCodecs := make(map[string]cli.OutputCodec, len(codecs))
	var formatNames []string
	for _, codec := range codecs {
		aliases := cli.FormatAliases(codec)
		if len(aliases) == 0 {
			continue
		}
		if _, exists := customCodecs[aliases[0]]; !exists {
			formatNames = append(formatNames, aliases[0])
		}
		for _, name := range aliases {
			customCodecs[name] = codec
		}
	}

	filtered := commands
	if len(unknowns.Args) > 0 {
		filtered = clihelp.FilterCommands(commands, unknowns.Args)
	}

	// built-in json/jsonschema keep their dedicated renderers even if a codec also claims that name;
	// every other registered codec renders the command tree.
	if format != "json" && format != "jsonschema" {
		if codec, ok := customCodecs[format]; ok {
			if err := clihelp.DisplayHelpEncoded(os.Stdout, filtered, codec, options.HelpValues); err != nil {
				return fmt.Errorf("failed to display help as %q: %w", format, err)
			}
			return nil
		}
	}

	switch format {
	case "json":
		clihelp.DisplayHelpJSON(os.Stdout, filtered, options.HelpValues)
		return nil
	case "jsonschema":
		clihelp.DisplayHelpJSONSchema(os.Stdout, filtered, options.HelpValues)
		return nil
	case "pretty", "plain", "md":
		clihelp.DisplayHelpAgent(os.Stdout, clihelp.AgentOptions{
			AppName:        appName,
			Format:         format,
			Commands:       filtered,
			Formats:        formatNames,
			ShowValues:     options.HelpValues,
			GlobalValues:   &options,
			DefaultCommand: defaultName,
		})
		return nil
	case "plain-flags":
		clihelp.DisplayHelp(os.Stdout, appName, commands, unknowns.Args, clihelp.DisplayOptions{
			ShowFlags:      true,
			ShowEnv:        true,
			ShowValues:     options.HelpValues,
			GlobalValues:   &options,
			Formats:        formatNames,
			DefaultCommand: defaultName,
		})
		return nil
	}

	clihelp.DisplayHelp(os.Stdout, appName, commands, unknowns.Args, clihelp.DisplayOptions{
		ShowValues:     options.HelpValues,
		GlobalValues:   &options,
		Formats:        formatNames,
		DefaultCommand: defaultName,
	})
	return nil
}

// Validate is a no-op; the command has no required inputs.
func (c *Command) Validate(_ map[string]any) error {
	return nil
}

// Help returns the one-line help summary for the command.
func (c *Command) Help() string {
	return "Display help"
}
