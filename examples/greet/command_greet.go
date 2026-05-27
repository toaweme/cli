package main

import (
	"fmt"

	"github.com/toaweme/cli"
)

// GreetConfig demonstrates struct tags for CLI arg parsing.
type GreetConfig struct {
	// arg:"0" makes this a positional arg (first non-flag argument)
	Name string `arg:"0" env:"GREET_NAME" help:"Name to greet" rules:"required"`
	// short:"s" enables the -s shorthand alongside --shout
	Shout bool `arg:"shout" short:"s" env:"GREET_SHOUT" help:"Uppercase the greeting"`
	// env:"GREET_REPEAT" allows setting via environment variable
	Repeat int `arg:"repeat" short:"r" env:"GREET_REPEAT" help:"Repeat the greeting N times"`
}

// GreetCommand greets someone by name with optional shouting and repetition.
type GreetCommand struct {
	cli.BaseCommand[GreetConfig]
}

var _ cli.Command[GreetConfig] = (*GreetCommand)(nil)

// Run accesses parsed inputs via c.Inputs, populated from CLI args, env vars, and defaults.
func (c *GreetCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	name := c.Inputs.Name
	if name == "" {
		name = "world"
	}

	msg := fmt.Sprintf("hello, %s!", name)
	if c.Inputs.Shout {
		msg = fmt.Sprintf("HELLO, %s!", name)
	}

	repeat := c.Inputs.Repeat
	if repeat < 1 {
		repeat = 1
	}

	for i := 0; i < repeat; i++ {
		fmt.Println(msg)
	}

	return nil
}

func (c *GreetCommand) Help() string {
	return "Greet someone by name"
}
