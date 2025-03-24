package help

import (
	"github.com/toaweme/cli"
)

type HelpVars struct {
}

type HelpCommand struct {
	cli.BaseCommand[HelpVars]

	appName         string
	commandListFunc func() []cli.Command[any]
}

var _ cli.Command[HelpVars] = (*HelpCommand)(nil)

func NewHelpCommand(appName string, commandList func() []cli.Command[any]) *HelpCommand {
	return &HelpCommand{appName: appName, commandListFunc: commandList}
}

func (c *HelpCommand) Run(options cli.GlobalOptions, unknowns cli.Unknowns) error {
	commands := c.commandListFunc()
	cli.DisplayHelp(c.appName, commands, unknowns.Args)
	return nil
}

func (c *HelpCommand) Validate(vars map[string]any) error {
	return nil
}

func (c *HelpCommand) Help() string {
	return "Display help"
}
