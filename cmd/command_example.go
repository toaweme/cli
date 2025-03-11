package main

import (
	"fmt"
	
	"github.com/toaweme/cli"
)

type ExampleVars struct {
	Verbose bool `arg:"verbose" help:"Verbose output"`
}

type ExampleCommand struct {
	cli.BaseCommand[ExampleVars]
}

var _ cli.Command[ExampleVars] = (*ExampleCommand)(nil)

func NewExampleCommand() *ExampleCommand {
	return &ExampleCommand{BaseCommand: cli.NewBaseCommand[ExampleVars]()}
}

func (c *ExampleCommand) Run(options cli.GlobalOptions, unknowns cli.Unknowns) error {
	fmt.Println("Example command")
	return nil
}

func (c *ExampleCommand) Validate(vars map[string]any) error {
	return nil
}

func (c *ExampleCommand) Help() string {
	return "Example command"
}
