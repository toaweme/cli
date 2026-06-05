// greet demonstrates positional args, named flags, short flags, environment variable binding, and validation rules.
package main

import (
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/commands/help"
)

const appName = "greet"
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
	app.Add("greet", &GreetCommand{BaseCommand: cli.NewBaseCommand[GreetConfig]()})

	if err := app.Run(os.Args[1:]); cli.IsRealError(err) {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
