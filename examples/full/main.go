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

	// config persists JSON under ~/.full/ (secrets under ~/.full/secrets with 0600).
	// register yaml/toml codecs via FileConfig.Codecs when an app needs them.
	cfg := cli.NewFileStorage(cli.FileStorage{Name: appName})

	app := cli.NewApp(
		// MergeLayered opts every command into config-file population: each command
		// reads shared top-level config plus its own "<name>:" section, then env,
		// then flags. Commands override per command via ConfigStrategy.
		cli.Config{Name: appName, Version: appVersion, Merge: cli.MergeLayered},
		cli.GlobalFlags{Cwd: cwd},
	).Store(cfg)

	app.Help(help.NewHelpCommand(app.Config, app.Commands, app.OutputFormats))
	app.Add("version", version.NewVersionCommand(app.Config))
	// generates bash/zsh/fish completion scripts: full completion bash
	app.Add("completion", completion.NewCompletionCommand(appName))
	app.Add("dev", dev.NewDevCommand(app.Config))

	buildCmd := &BuildCommand{BaseCommand: cli.NewBaseCommand[BuildConfig]()}
	app.Add("build", buildCmd)
	// default command runs when no args are given: just "full" runs build
	app.Default(buildCmd)

	app.Add("serve", &ServeCommand{BaseCommand: cli.NewBaseCommand[ServeConfig]()})

	// config commands take cfg explicitly through their constructors.
	cfgParent := help.NewParentPlaceholder()
	cfgParent.Add("show", NewConfigShowCommand(cfg))
	cfgParent.Add("set", NewConfigSetCommand(cfg))
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
