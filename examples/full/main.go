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

	// config has a global scope (~/.full/config.json) and a project scope
	// (./config.json); reads merge global then project, plus env, then flags.
	cfg := config.New().
		Add(config.Global, config.NewFileStore(config.HomePath(appName)), "config").
		Add(config.Project, config.NewFileStore(cwd), "config").
		WithSecrets(config.FileSecrets(config.HomePath(appName) + "/secrets"))

	// the resolver folds the config files into every command's Options(). The
	// serve command sources its fields from a "server:" section via a mapping rule.
	resolver := config.NewFileResolver(cfg, map[string]map[string]config.Source{
		"serve": {
			"port": "server.port",
			"host": "server.host",
			"tls":  "server.tls",
		},
	})

	app := cli.NewApp(
		cli.Config{Name: appName, Version: appVersion},
		cli.GlobalFlags{Cwd: cwd},
	).Resolve(resolver)

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
