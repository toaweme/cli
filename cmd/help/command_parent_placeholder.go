package help

import (
	"github.com/awee-ai/cli"
)

type ParentVars struct{}

type ParentCommand struct {
	cli.BaseCommand[ParentVars]
}

var _ cli.Command[ParentVars] = (*ParentCommand)(nil)

func NewParentPlaceholder() *ParentCommand {
	return &ParentCommand{BaseCommand: cli.NewBaseCommand[ParentVars]()}
}

func (c *ParentCommand) Run(options cli.GlobalOptions, unknowns cli.Unknowns) error {
	return cli.ErrDisplaySubCommands
}

func (c *ParentCommand) Help() string {
	return ""
}
