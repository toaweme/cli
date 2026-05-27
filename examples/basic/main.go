// basic demonstrates the minimal setup for a CLI app:
// creating an app, registering built-in commands, and adding a custom command.
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

	// every app starts with NewApp, passing identity and global options
	app := cli.NewApp(
		cli.Settings{Name: appName, Version: appVersion},
		cli.GlobalOptions{Cwd: cwd},
	)

	// built-in commands: help and version are opt-in, not automatic
	app.Add("help", help.NewHelpCommand(appName, app.Commands))
	app.Add("version", version.NewVersionCommand(appName, appVersion))
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
