package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/cmd/help"
	"github.com/toaweme/cli/cmd/version"
)

const appName = "greet"
const appVersion = "0.1.0"

// GreetConfig holds the inputs for the greet command.
type GreetConfig struct {
	Name   string `arg:"0" env:"GREET_NAME" help:"Name to greet" validate:"required"`
	Shout  bool   `arg:"shout" short:"s" env:"GREET_SHOUT" help:"Uppercase the greeting"`
	Repeat int    `arg:"repeat" short:"r" env:"GREET_REPEAT" help:"Repeat the greeting N times"`
}

// GreetCommand greets someone by name with optional shouting and repetition.
type GreetCommand struct {
	cli.BaseCommand[GreetConfig]
}

var _ cli.Command[GreetConfig] = (*GreetCommand)(nil)

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

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	app := cli.NewApp(
		cli.Settings{Name: appName, Version: appVersion},
		cli.GlobalOptions{Cwd: cwd},
	)

	app.Add("help", help.NewHelpCommand(appName, app.Commands))
	app.Add("version", version.NewVersionCommand(appName, appVersion))
	app.Add("greet", &GreetCommand{BaseCommand: cli.NewBaseCommand[GreetConfig]()})

	err = app.Run(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
