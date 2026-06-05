// full_3rd_party is a self-contained module (its own go.mod) that exercises every method on the cli.App interface
// and wires in the third-party yaml/toml output codecs from github.com/toaweme/cli/config/addons.
// The same codec value is used twice: as a config.Codec (so config files can be yaml/toml)
// and a cli.OutputCodec (so `--help-format yml|toml` renders the command tree) - the core never
// imports yaml or toml, the addon modules carry those dependencies.
package main

import (
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/commands/completion"
	"github.com/toaweme/cli/commands/gendocs"
	"github.com/toaweme/cli/commands/help"
	"github.com/toaweme/cli/config"
	tomlcodec "github.com/toaweme/cli/config/addons/toml"
	yamlcodec "github.com/toaweme/cli/config/addons/yaml"
)

const appName = "full3p"
const appVersion = "0.1.0"

func main() {
	// load .env from cwd if present; a missing file is not an error.
	if err := cli.LoadDotEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load .env: %v\n", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// one codec instance, two roles. As config.Codec it lets the store read/write config.yml;
	// as cli.OutputCodec (a structural subset) it backs the --help-format yml|toml help output.
	// The toml codec is used only for help output here.
	yc := yamlcodec.New() // recognizes .yml and .yaml; .yml is primary (output)
	tc := tomlcodec.New()

	// each store is one file with one codec: config.yml and secrets.yml under ~/.full3p/.
	// Secrets are just another store.
	store := config.NewFileStore(config.HomePath(appName), "config", yc)
	secrets := config.FileSecrets(config.HomePath(appName), yc)

	// one resolver per store; the serve command sources its fields from a "server:" section via a mapping rule.
	// The App runs the chain lowest precedence first.
	serveRules := map[string]map[string]config.Source{
		"serve": {
			"port": "server.port",
			"host": "server.host",
			"tls":  "server.tls",
		},
	}

	// NewApp takes the serializable Config DTO and the seed global flags;
	// the chainable Resolve and HelpOutputs setters attach the runtime objects.
	app := cli.NewApp(
		cli.Config{Name: appName, Version: appVersion},
		cli.GlobalFlags{Cwd: cwd},
	).
		Resolve( // App.Resolve: attach the config resolver chain
			config.NewResolver(store, serveRules),
			config.NewResolver(secrets, nil),
		).
		HelpOutputs(yc, tc) // App.HelpOutputs: register yaml/toml as --help-format values

	// App.Help registers the help command under the reserved name.
	// It is handed the App.Config, App.Commands, and App.OutputFormats getters so it can render
	// the identity, the command tree, and the registered custom formats lazily.
	app.Help(help.NewHelpCommand(app.Config, app.Commands, app.OutputFormats))

	// App.Add registers a command under a name and returns it.
	app.Add("completion", completion.NewCompletionCommand(appName))                            // full3p completion bash|zsh|fish
	app.Add("gendocs", gendocs.NewGenDocsCommand(app.Config, app.Commands, app.OutputFormats)) // full3p gendocs -o docs

	build := &BuildCommand{BaseCommand: cli.NewBaseCommand[BuildConfig]()}
	app.Add("build", build)
	// App.Default runs this command when no arguments are given: bare `full3p` builds.
	app.Default(build)

	app.Add("serve", &ServeCommand{BaseCommand: cli.NewBaseCommand[ServeConfig]()})

	// a parent placeholder groups subcommands under "db"; Add chains onto the parent.
	db := help.NewParentPlaceholder()
	app.Add("db", db)
	db.Add("migrate", &DBMigrateCommand{BaseCommand: cli.NewBaseCommand[DBMigrateConfig]()})

	// App.Commands returns the registered top-level commands, in registration order.
	fmt.Fprintf(os.Stderr, "registered %d top-level commands\n", len(app.Commands()))

	// App.Run parses os.Args[1:] and dispatches.
	// Help and version requests come back as sentinel errors, which are clean exits rather than failures.
	if err := app.Run(os.Args[1:]); cli.IsRealError(err) {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
