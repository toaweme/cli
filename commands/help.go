package commands

import (
	"fmt"

	"github.com/contentforward/cli"
)

type HelpVars struct {
}
type HelpCommand struct {
	commandListFunc func() map[string]cli.Command
}

func (c *HelpCommand) Structure() any {
	return &HelpVars{}
}

func (c *HelpCommand) Help() any {
	return "Display help"
}

func NewHelpCommand(commandList func() map[string]cli.Command) *HelpCommand {
	return &HelpCommand{commandListFunc: commandList}
}

var _ cli.Command = (*HelpCommand)(nil)

func (c *HelpCommand) Run(vars map[string]any) error {
	commands := c.commandListFunc()
	fmt.Printf("Available commands:\n")
	for name, cmd := range commands {
		fmt.Printf("  %s: %v\n", name, cmd.Help())
	}
	return nil
}

func (c *HelpCommand) Validate(vars map[string]any) error {
	return nil
}
