// server demonstrates a CLI app that starts an HTTP server with graceful shutdown, dotenv loading, config persistence,
// and signal handling.
package main

import (
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/commands/completion"
	"github.com/toaweme/cli/commands/help"
	"github.com/toaweme/cli/config"
)

const appName = "server"
const appVersion = "0.1.0"

func main() {
	if err := cli.LoadDotEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load .env: %v\n", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// config store for the last-run server state at ~/.server/last.json
	store := config.NewFileStore(config.HomePath(appName), "last", true)

	app := cli.NewApp(
		cli.Config{Name: appName, Version: appVersion},
		cli.GlobalFlags{Cwd: cwd},
	)

	app.Help(help.NewHelpCommand(app.Config, app.Commands, app.OutputFormats))
	app.Add("completion", completion.NewCompletionCommand(appName))

	startCmd := &StartCommand{
		BaseCommand: cli.NewBaseCommand[StartConfig](),
		store:       store,
	}
	app.Add("start", startCmd)
	// running "server" with no args starts the server
	app.Default(startCmd)

	app.Add("status", &StatusCommand{BaseCommand: cli.NewBaseCommand[StatusConfig]()})

	if err := app.Run(os.Args[1:]); cli.IsRealError(err) {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
