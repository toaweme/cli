// basic demonstrates the minimal setup for a CLI app:
// creating an app, registering built-in commands, and adding a custom command.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/commands/help"
	"github.com/toaweme/cli/commands/version"
)

const appName = "basic"
const appVersion = "0.1.0"

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// every app starts with NewApp, passing identity and global options
	app := cli.NewApp(
		cli.Config{Name: appName, Version: appVersion},
		cli.GlobalFlags{Cwd: cwd},
	)

	// built-in commands: help and version are opt-in, not automatic
	app.Help(help.NewHelpCommand(app.Config, app.Commands, app.OutputFormats))
	app.Add("version", version.NewVersionCommand(app.Config))
	app.Add("info", &InfoCommand{BaseCommand: cli.NewBaseCommand[InfoConfig]()})

	// ErrShowingHelp and ErrShowingVersion are sentinel errors, not failures
	err = app.Run(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
