package help

import (
	"github.com/toaweme/cli"
)

// ParentConfig holds the inputs for a parent placeholder command.
type ParentConfig struct{}

// ParentCommand is a placeholder for commands that only serve as a
// namespace for subcommands. Running it directly displays its subcommands.
type ParentCommand struct {
	cli.BaseCommand[ParentConfig]
}

var _ cli.Command[ParentConfig] = (*ParentCommand)(nil)

// NewParentPlaceholder creates a parent command that displays its subcommands
// when invoked directly.
func NewParentPlaceholder() *ParentCommand {
	return &ParentCommand{BaseCommand: cli.NewBaseCommand[ParentConfig]()}
}

func (c *ParentCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	return cli.ErrDisplaySubCommands
}

func (c *ParentCommand) Help() string {
	return ""
}
