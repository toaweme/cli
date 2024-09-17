package help

import (
	"fmt"

	"github.com/contentforward/cli"
)

type HelpVars struct {
	Verbose bool `arg:"-v" help:"Verbose output"`
	Help    bool `arg:"help" short:"h" help:"Show help"`
}

type HelpCommand struct {
	cli.BaseCommand[HelpVars]

	commandListFunc func() []cli.Command[any]
}

var _ cli.Command[HelpVars] = (*HelpCommand)(nil)

func NewHelpCommand(commandList func() []cli.Command[any]) *HelpCommand {
	return &HelpCommand{commandListFunc: commandList}
}

func (c *HelpCommand) Run(options cli.GlobalOptions) error {
	commands := c.commandListFunc()
	fmt.Printf("\nAvailable commands:\n")
	cli.PrintCommands(commands)
	fmt.Printf("\nAvailable options:\n")
	err := cli.PrintOptions(&options)
	if err != nil {
		return fmt.Errorf("failed to print options: %w", err)
	}
	return nil
}

func (c *HelpCommand) Validate(vars map[string]any) error {
	return nil
}

func (c *HelpCommand) Help() string {
	return "Display help"
}
