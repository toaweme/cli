// basic demonstrates the minimal setup for a CLI app:
// creating an app, registering built-in commands, and adding a custom command.
package main

import (
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/commands/help"
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

	// built-in commands: help is opt-in, not automatic. version is a built-in flag (--version / -V), not a command.
	app.Help(help.NewHelpCommand(app.Config, app.Commands, app.OutputFormats))
	app.Add("info", &InfoCommand{BaseCommand: cli.NewBaseCommand[InfoConfig]()})

	// IsRealError filters out the clean-exit sentinels (help/version already handled)
	if err := app.Run(os.Args[1:]); cli.IsRealError(err) {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
