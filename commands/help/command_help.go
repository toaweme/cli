package help

import (
	"fmt"
	"os"
	"strings"

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
// App.Formats so the help renderer can advertise and apply custom --format values.
func NewHelpCommand(settingsFunc func() cli.Config, commandList func() []cli.Command[any], formats func() []cli.OutputCodec) *HelpCommand {
	return &HelpCommand{settingsFunc: settingsFunc, commandListFunc: commandList, formatsFunc: formats}
}

func (c *HelpCommand) Run(options cli.GlobalFlags, unknowns cli.Unknowns) error {
	cfg := c.settingsFunc()
	commands := c.commandListFunc()
	appName := cfg.Name

	format := options.Format

	codecs := c.formatsFunc()

	// output codecs registered on the app (yaml, toml, ...), keyed by their
	// --format name. formatNames preserves registration order for the help hint.
	customCodecs := make(map[string]cli.OutputCodec, len(codecs))
	var formatNames []string
	for _, codec := range codecs {
		name := strings.TrimPrefix(codec.Extension(), ".")
		if name == "" {
			continue
		}
		if _, exists := customCodecs[name]; !exists {
			formatNames = append(formatNames, name)
		}
		customCodecs[name] = codec
	}

	filtered := commands
	if len(unknowns.Args) > 0 {
		filtered = clihelp.FilterCommands(commands, unknowns.Args)
	}

	// built-in json/jsonschema keep their dedicated renderers even if a codec also
	// claims that name; every other registered codec renders the command tree.
	if format != "json" && format != "jsonschema" {
		if codec, ok := customCodecs[format]; ok {
			if err := clihelp.DisplayHelpEncoded(os.Stdout, filtered, codec); err != nil {
				return fmt.Errorf("failed to display help as %q: %w", format, err)
			}
			return nil
		}
	}

	switch format {
	case "json":
		clihelp.DisplayHelpJSON(os.Stdout, filtered)
		return nil
	case "jsonschema":
		clihelp.DisplayHelpJSONSchema(os.Stdout, filtered)
		return nil
	case "pretty", "plain", "md":
		clihelp.DisplayHelpAgent(os.Stdout, clihelp.AgentOptions{
			AppName:  appName,
			Format:   format,
			Commands: filtered,
			Formats:  formatNames,
		})
		return nil
	case "plain-flags":
		clihelp.DisplayHelp(os.Stdout, appName, commands, unknowns.Args, clihelp.HelpDisplayOptions{
			ShowFlags: true,
			ShowEnv:   true,
			Formats:   formatNames,
		})
		return nil
	}

	clihelp.DisplayHelp(os.Stdout, appName, commands, unknowns.Args, clihelp.HelpDisplayOptions{Formats: formatNames})
	return nil
}

func (c *HelpCommand) Validate(_ map[string]any) error {
	return nil
}

func (c *HelpCommand) Help() string {
	return "Display help"
}
