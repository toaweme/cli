package help

import (
	"github.com/toaweme/cli"
)

// ParentConfig is the (empty) config for a parent placeholder command, which takes no inputs.
type ParentConfig struct{}

// ParentCommand is a placeholder for commands that only serve as a namespace
// for subcommands. Running it directly displays its subcommands.
type ParentCommand struct {
	cli.BaseCommand[ParentConfig]
}

var _ cli.Command[ParentConfig] = (*ParentCommand)(nil)

// NewParentPlaceholder creates a parent command that displays its subcommands when invoked directly.
func NewParentPlaceholder() *ParentCommand {
	return &ParentCommand{BaseCommand: cli.NewBaseCommand[ParentConfig]()}
}

// Run signals that the command's subcommands should be displayed.
func (c *ParentCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	return cli.ErrDisplaySubCommands
}

// Help returns an empty summary; the placeholder has no help of its own.
func (c *ParentCommand) Help() string {
	return ""
}
