package main

import (
	"fmt"

	"github.com/toaweme/cli"
)

// InfoConfig is an empty config struct, meaning the command takes no flags or args.
type InfoConfig struct{}

// InfoCommand prints the current global options for debugging.
// Embed BaseCommand[T] to get default Name, Add, Options, Validate for free.
type InfoCommand struct {
	cli.BaseCommand[InfoConfig]
}

// compile-time assertion: InfoCommand implements Command[InfoConfig]
var _ cli.Command[InfoConfig] = (*InfoCommand)(nil)

// Run receives GlobalOptions which are available to every command.
func (c *InfoCommand) Run(options cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf("cwd=%s verbosity=%d\n", options.Cwd, options.Verbosity)
	return nil
}

func (c *InfoCommand) Validate(_ map[string]any) error {
	return nil
}

// Help returns the one-line description shown in command listings.
func (c *InfoCommand) Help() string {
	return "Print current global options"
}
