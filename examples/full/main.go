// full demonstrates every framework feature: dotenv loading, default commands,
// shell completion, config store, subcommands, ExampleProvider, and all help formats.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/commands/completion"
	"github.com/toaweme/cli/commands/dev"
	"github.com/toaweme/cli/commands/help"
	"github.com/toaweme/cli/commands/version"
	"github.com/toaweme/cli/config"
)

const appName = "full"
const appVersion = "0.1.0"

func main() {
	// load .env from cwd if present; missing file is not an error
	if err := cli.DotEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load .env: %v\n", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	app := cli.NewApp(
		cli.Settings{Name: appName, Version: appVersion},
		cli.GlobalOptions{Cwd: cwd},
	)

	app.Add("help", help.NewHelpCommand(app.Settings, app.Commands))
	app.Add("version", version.NewVersionCommand(app.Settings))
	// generates bash/zsh/fish completion scripts: full completion bash
	app.Add("completion", completion.NewCompletionCommand(appName))
	app.Add("dev", dev.NewDevCommand(app.Settings))

	buildCmd := &BuildCommand{BaseCommand: cli.NewBaseCommand[BuildConfig]()}
	app.Add("build", buildCmd)
	// default command runs when no args are given: just "full" runs build
	app.Default(buildCmd)

	app.Add("serve", &ServeCommand{BaseCommand: cli.NewBaseCommand[ServeConfig]()})

	// config store persists JSON to ~/.full/ with 0644 perms
	configStore := config.NewFileStore(config.HomePath(appName))
	cfgParent := help.NewParentPlaceholder()
	cfgParent.Add("show", &ConfigShowCommand{
		BaseCommand: cli.NewBaseCommand[ConfigShowConfig](),
		store:       configStore,
	})
	cfgParent.Add("set", &ConfigSetCommand{
		BaseCommand: cli.NewBaseCommand[ConfigSetConfig](),
		store:       configStore,
	})
	app.Add("config", cfgParent)

	// parent placeholder groups subcommands under "db"
	db := help.NewParentPlaceholder()
	app.Add("db", db)
	db.Add("migrate", &DBMigrateCommand{BaseCommand: cli.NewBaseCommand[DBMigrateConfig]()})
	db.Add("seed", &DBSeedCommand{BaseCommand: cli.NewBaseCommand[DBSeedConfig]()})
	db.Add("reset", &DBResetCommand{BaseCommand: cli.NewBaseCommand[DBResetConfig]()})

	err = app.Run(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
