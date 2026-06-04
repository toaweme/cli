// full demonstrates every framework feature: dotenv loading, default commands,
// shell completion, config store, subcommands, ExampleProvider, and all help formats.
package main

import (
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/commands/completion"
	"github.com/toaweme/cli/commands/dev"
	"github.com/toaweme/cli/commands/help"
	"github.com/toaweme/cli/config"
)

const appName = "full"
const appVersion = "0.1.0"

func main() {
	// load .env from cwd if present; missing file is not an error
	if err := cli.LoadDotEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load .env: %v\n", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// each store is one config file: a global one (~/.full/config.json), a project
	// one (./config.json), and a secrets file (~/.full/secrets.json, 0600). Secrets
	// are just another store, so they layer in like any other.
	global := config.NewFileStore(config.HomePath(appName), "config")
	project := config.NewFileStore(cwd, "config")
	secrets := config.FileSecrets(config.HomePath(appName))

	// one resolver per store; the App runs them in order, lowest precedence first, so
	// project overrides global and secrets overlay both, then env, then flags. The
	// serve command sources its fields from a "server:" section via a mapping rule.
	serveRules := map[string]map[string]config.Source{
		"serve": {
			"port": "server.port",
			"host": "server.host",
			"tls":  "server.tls",
		},
	}

	app := cli.NewApp(
		cli.Config{Name: appName, Version: appVersion},
		cli.GlobalFlags{Cwd: cwd},
	).Resolve(
		config.NewResolver(global, nil),
		config.NewResolver(project, serveRules),
		config.NewResolver(secrets, nil),
	)

	app.Help(help.NewHelpCommand(app.Config, app.Commands, app.OutputFormats))
	// generates bash/zsh/fish completion scripts: full completion bash
	app.Add("completion", completion.NewCompletionCommand(appName))
	app.Add("dev", dev.NewDevCommand(app.Config))

	buildCmd := &BuildCommand{BaseCommand: cli.NewBaseCommand[BuildConfig]()}
	app.Add("build", buildCmd)
	// default command runs when no args are given: just "full" runs build
	app.Default(buildCmd)

	app.Add("serve", &ServeCommand{BaseCommand: cli.NewBaseCommand[ServeConfig]()})

	// config commands take the global store explicitly through their constructors.
	cfgParent := help.NewParentPlaceholder()
	cfgParent.Add("show", NewConfigShowCommand(global))
	cfgParent.Add("set", NewConfigSetCommand(global))
	app.Add("config", cfgParent)

	// parent placeholder groups subcommands under "db"
	db := help.NewParentPlaceholder()
	app.Add("db", db)
	db.Add("migrate", &DBMigrateCommand{BaseCommand: cli.NewBaseCommand[DBMigrateConfig]()})
	db.Add("seed", &DBSeedCommand{BaseCommand: cli.NewBaseCommand[DBSeedConfig]()})
	db.Add("reset", &DBResetCommand{BaseCommand: cli.NewBaseCommand[DBResetConfig]()})

	if err := app.Run(os.Args[1:]); cli.IsRealError(err) {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
