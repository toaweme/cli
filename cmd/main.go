package main

import (
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/cmd/help"
	"github.com/toaweme/cli/cmd/version"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(fmt.Errorf("failed to get current working directory: %w", err))
		return
	}

	options := cli.GlobalOptions{
		Cwd:       cwd,
		Help:      false,
		Verbosity: 0,
	}
	app := cli.NewApp(
		cli.Settings{
			Name:    "cli",
			Version: "0.1.0",
		},
		options,
	)

	commandHelp := help.NewHelpCommand("cli", app.Commands)
	commandVersion := version.NewVersionCommand("cli", "0.1.0")
	commandExample := NewExampleCommand()

	app.Add("help", commandHelp)
	app.Add("version", commandVersion)
	app.Add("example", commandExample)

	err = app.Run(os.Args[1:])
	if err != nil {
		fmt.Println(fmt.Errorf("failed to run app: %w", err))
		return
	}
}
