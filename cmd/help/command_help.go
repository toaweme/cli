package help

import (
	"strings"

	"github.com/toaweme/cli"
)

// HelpConfig holds the inputs for the help command.
type HelpConfig struct{}

// HelpCommand displays usage information for the application or a specific command.
type HelpCommand struct {
	cli.BaseCommand[HelpConfig]

	appName         string
	commandListFunc func() []cli.Command[any]
}

var _ cli.Command[HelpConfig] = (*HelpCommand)(nil)

// NewHelpCommand creates a help command that lists all available commands.
// Pass the app name for display and a function that returns the command list.
func NewHelpCommand(appName string, commandList func() []cli.Command[any]) *HelpCommand {
	return &HelpCommand{appName: appName, commandListFunc: commandList}
}

func (c *HelpCommand) Run(options cli.GlobalOptions, unknowns cli.Unknowns) error {
	commands := c.commandListFunc()

	if options.Agent {
		var filter []string
		if options.Filter != "" {
			filter = strings.Split(options.Filter, ",")
		}
		cli.DisplayHelpAgent(cli.AgentOptions{
			AppName:  c.appName,
			Format:   options.Format,
			Filter:   filter,
			Commands: commands,
		})
		return nil
	}

	if options.JSON {
		cli.DisplayHelpJSON(commands)
		return nil
	}

	if options.JSONSchema {
		cli.DisplayHelpJSONSchema(commands)
		return nil
	}

	cli.DisplayHelp(c.appName, commands, unknowns.Args, cli.HelpDisplayOptions{
		ShowFlags: options.Flags || options.Env,
		ShowEnv:   options.Env,
	})
	return nil
}

func (c *HelpCommand) Validate(_ map[string]any) error {
	return nil
}

func (c *HelpCommand) Help() string {
	return "Display help"
}
