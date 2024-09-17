package main

import (
	"fmt"
	"os"

	"github.com/contentforward/cli"
	"github.com/contentforward/cli/cmd/help"
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

	commandHelp := help.NewHelpCommand(app.Commands)
	commandExample := NewExampleCommand()
	commandExample2 := NewExampleCommand()
	commandInit := NewInitCommand()

	app.Add("help", commandHelp)
	app.Add("example", commandExample)
	app.Add("init", commandInit)
	app.Add("init somesing", commandExample2)

	err = app.Run(os.Args[1:])
	if err != nil {
		fmt.Println(fmt.Errorf("failed to run app: %w", err))
		return
	}
}
