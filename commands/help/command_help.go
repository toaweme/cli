package help

import (
	"github.com/toaweme/cli"
	clihelp "github.com/toaweme/cli/help"
)

// HelpConfig holds the inputs for the help command.
type HelpConfig struct {
	// Flags expands help output to show all flags for each command.
	Flags bool `arg:"flags" help:"Show all flags for each command"`
}

// HelpCommand displays usage information for the application or a specific command.
type HelpCommand struct {
	cli.BaseCommand[HelpConfig]

	settingsFunc    func() cli.Settings
	commandListFunc func() []cli.Command[any]
}

var _ cli.Command[HelpConfig] = (*HelpCommand)(nil)

// NewHelpCommand creates a help command that lists all available commands.
func NewHelpCommand(settingsFunc func() cli.Settings, commandList func() []cli.Command[any]) *HelpCommand {
	return &HelpCommand{settingsFunc: settingsFunc, commandListFunc: commandList}
}

func (c *HelpCommand) Run(options cli.GlobalOptions, unknowns cli.Unknowns) error {
	commands := c.commandListFunc()
	appName := c.settingsFunc().Name

	format := options.Format

	filtered := commands
	if len(unknowns.Args) > 0 {
		filtered = clihelp.FilterCommands(commands, unknowns.Args)
	}

	switch format {
	case "json":
		clihelp.DisplayHelpJSON(filtered)
		return nil
	case "jsonschema":
		clihelp.DisplayHelpJSONSchema(filtered)
		return nil
	case "pretty", "plain", "md":
		clihelp.DisplayHelpAgent(clihelp.AgentOptions{
			AppName:  appName,
			Format:   format,
			Commands: filtered,
		})
		return nil
	}

	showFlags := c.Inputs != nil && c.Inputs.Flags
	if !showFlags {
		if _, ok := unknowns.Options["flags"]; ok {
			showFlags = true
		}
	}

	clihelp.DisplayHelp(appName, commands, unknowns.Args, clihelp.HelpDisplayOptions{
		ShowFlags: showFlags,
		ShowEnv:   showFlags,
	})
	return nil
}

func (c *HelpCommand) Validate(_ map[string]any) error {
	return nil
}

func (c *HelpCommand) Help() string {
	return "Display help"
}
