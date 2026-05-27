// server demonstrates a CLI app that starts an HTTP server with graceful shutdown,
// dotenv loading, config persistence, and signal handling.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/cmd/completion"
	"github.com/toaweme/cli/cmd/help"
	"github.com/toaweme/cli/cmd/version"
	"github.com/toaweme/cli/config"
)

const appName = "server"
const appVersion = "0.1.0"

func main() {
	if err := cli.DotEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load .env: %v\n", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// config store for server settings at ~/.server/
	store := config.NewFileStore(config.HomePath(appName))

	app := cli.NewApp(
		cli.Settings{Name: appName, Version: appVersion},
		cli.GlobalOptions{Cwd: cwd},
	)

	app.Add("help", help.NewHelpCommand(appName, app.Commands))
	app.Add("version", version.NewVersionCommand(appName, appVersion))
	app.Add("completion", completion.NewCompletionCommand(appName))

	startCmd := &StartCommand{
		BaseCommand: cli.NewBaseCommand[StartConfig](),
		store:       store,
	}
	app.Add("start", startCmd)
	// running "server" with no args starts the server
	app.Default(startCmd)

	app.Add("status", &StatusCommand{BaseCommand: cli.NewBaseCommand[StatusConfig]()})

	err = app.Run(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
