package main

import (
	"github.com/contentforward/cli"
)

type IndexVars struct {
	Verbose bool `arg:"-v" help:"Verbose output"`
}

// IndexCommand is used to display the subcommands
type IndexCommand struct {
	cli.BaseCommand[IndexVars]
}

var _ cli.Command[IndexVars] = (*IndexCommand)(nil)

func NewIndexCommand() *IndexCommand {
	return &IndexCommand{}
}

func (c *IndexCommand) Run(options cli.GlobalOptions) error {
	return nil
}

func (c *IndexCommand) Validate(vars map[string]any) error {
	return nil
}

func (c *IndexCommand) Help() string {
	return "Index command"
}
