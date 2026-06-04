package help

import (
	"fmt"
	"os"

	"github.com/toaweme/cli"
	clihelp "github.com/toaweme/cli/help"
)

// HelpConfig holds the inputs for the help command.
type HelpConfig struct{}

// HelpCommand displays usage information for the application or a specific command.
type HelpCommand struct {
	cli.BaseCommand[HelpConfig]

	settingsFunc    func() cli.Config
	commandListFunc func() []cli.Command[any]
	formatsFunc     func() []cli.OutputCodec
}

var _ cli.Command[HelpConfig] = (*HelpCommand)(nil)

// NewHelpCommand creates a help command that lists all available commands. The
// formats getter (typically App.OutputFormats) supplies the codecs registered via
// App.Formats so the help renderer can advertise and apply custom --help-format values.
func NewHelpCommand(settingsFunc func() cli.Config, commandList func() []cli.Command[any], formats func() []cli.OutputCodec) *HelpCommand {
	return &HelpCommand{settingsFunc: settingsFunc, commandListFunc: commandList, formatsFunc: formats}
}

func (c *HelpCommand) Run(options cli.GlobalFlags, unknowns cli.Unknowns) error {
	cfg := c.settingsFunc()
	commands := c.commandListFunc()
	appName := cfg.Name

	format := options.HelpFormat

	codecs := c.formatsFunc()

	// output codecs registered on the app (yaml, toml, ...), keyed by every --help-format
	// name they answer to (e.g. both "yml" and "yaml"). formatNames lists only each
	// codec's primary name, in registration order, for the help hint.
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

	// built-in json/jsonschema keep their dedicated renderers even if a codec also
	// claims that name; every other registered codec renders the command tree.
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
			AppName:      appName,
			Format:       format,
			Commands:     filtered,
			Formats:      formatNames,
			ShowValues:   options.HelpValues,
			GlobalValues: &options,
		})
		return nil
	case "plain-flags":
		clihelp.DisplayHelp(os.Stdout, appName, commands, unknowns.Args, clihelp.DisplayOptions{
			ShowFlags:    true,
			ShowEnv:      true,
			ShowValues:   options.HelpValues,
			GlobalValues: &options,
			Formats:      formatNames,
		})
		return nil
	}

	clihelp.DisplayHelp(os.Stdout, appName, commands, unknowns.Args, clihelp.DisplayOptions{
		ShowValues:   options.HelpValues,
		GlobalValues: &options,
		Formats:      formatNames,
	})
	return nil
}

func (c *HelpCommand) Validate(_ map[string]any) error {
	return nil
}

func (c *HelpCommand) Help() string {
	return "Display help"
}
