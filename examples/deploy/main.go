// deploy demonstrates parent commands with subcommands.
// The "deploy" command has no logic of its own; it serves as a namespace
// for "deploy staging" and "deploy production".
package main

import (
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/commands/help"
)

const appName = "deploy"
const appVersion = "0.1.0"

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	app := cli.NewApp(
		cli.Config{Name: appName, Version: appVersion},
		cli.GlobalFlags{Cwd: cwd},
	)

	app.Help(help.NewHelpCommand(app.Config, app.Commands, app.OutputFormats))

	// NewParentPlaceholder creates a command that only holds subcommands.
	// Running "deploy" alone shows its subcommands via help.
	parent := help.NewParentPlaceholder()
	app.Add("deploy", parent)
	// subcommands share the same config struct but target different environments
	parent.Add("staging", &DeployCommand{BaseCommand: cli.NewBaseCommand[DeployConfig]()})
	parent.Add("production", &DeployCommand{BaseCommand: cli.NewBaseCommand[DeployConfig]()})

	if err := app.Run(os.Args[1:]); cli.IsRealError(err) {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
