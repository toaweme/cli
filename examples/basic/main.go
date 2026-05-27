package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/cmd/help"
	"github.com/toaweme/cli/cmd/version"
)

const appName = "basic"
const appVersion = "0.1.0"

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
	app.Add("info", &InfoCommand{BaseCommand: cli.NewBaseCommand[InfoConfig]()})

	err = app.Run(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// InfoConfig holds the inputs for the info command.
type InfoConfig struct{}

// InfoCommand prints the current global options for debugging.
type InfoCommand struct {
	cli.BaseCommand[InfoConfig]
}

var _ cli.Command[InfoConfig] = (*InfoCommand)(nil)

func (c *InfoCommand) Run(options cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf("cwd=%s verbosity=%d\n", options.Cwd, options.Verbosity)
	return nil
}

func (c *InfoCommand) Validate(_ map[string]any) error {
	return nil
}

func (c *InfoCommand) Help() string {
	return "Print current global options"
}
