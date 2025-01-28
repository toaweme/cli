package main

import (
	"fmt"
	"os"

	"github.com/awee-ai/cli"
	"github.com/awee-ai/cli/cmd/help"
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
		cli.Settings{},
		options,
	)

	commandHelp := help.NewHelpCommand("cli", app.Commands)
	commandExample := NewExampleCommand()

	app.Add("help", commandHelp)
	app.Add("example", commandExample)

	err = app.Run(os.Args[1:])
	if err != nil {
		fmt.Println(fmt.Errorf("failed to run app: %w", err))
		return
	}
}
