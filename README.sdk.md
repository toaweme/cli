# SDK Reference

A small, struct-tag-driven CLI framework. You define a config struct per command,
embed `BaseCommand[T]`, implement `Run`, and register the command on an app.

## Minimal app

```go
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/commands/help"
	"github.com/toaweme/cli/commands/version"
)

type GreetConfig struct {
	Name string `arg:"0" help:"who to greet" rules:"required"`
	Loud bool   `arg:"loud" short:"l" help:"shout it"`
}

type GreetCommand struct {
	cli.BaseCommand[GreetConfig]
}

func (c *GreetCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	msg := "hello, " + c.Inputs.Name
	if c.Inputs.Loud {
		msg += "!!!"
	}
	fmt.Println(msg)
	return nil
}

func (c *GreetCommand) Help() string { return "Greet someone" }

func main() {
	app := cli.NewApp(
		cli.Config{Name: "myapp", Version: "1.0.0"},
		cli.GlobalFlags{},
	)

	app.Help(help.NewHelpCommand(app.Config, app.Commands, app.OutputFormats)) // registers the help command
	app.Add("version", version.NewVersionCommand(app.Config))
	app.Add("greet", &GreetCommand{BaseCommand: cli.NewBaseCommand[GreetConfig]()})

	if err := app.Run(os.Args[1:]); err != nil {
		// help and version are reported as sentinel errors, not failures
		if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

`myapp greet Sam -l` prints `hello, Sam!!!`.

## App

`NewApp(Config, GlobalFlags)` returns a value satisfying `App`:

```go
type App interface {
	// Commands returns the registered top-level commands.
	Commands() []Command[any]
	// Config returns the app identity (the serializable DTO).
	Config() Config
	// OutputFormats returns the registered help output codecs.
	OutputFormats() []OutputCodec
	// Resolve appends config Resolvers to the chain used to populate command Options()
	// before Run, and returns the app for chaining. The chain runs lowest-precedence
	// first, then env, then flags. With none registered, only env and flags apply.
	Resolve(resolvers ...Resolver) App
	// HelpOutputs registers additional help output codecs and returns the app for chaining.
	HelpOutputs(formats ...OutputCodec) App
	// Default registers the command run when invoked with no arguments.
	Default(cmd Command[any]) Command[any]
	// Add registers cmd under name and returns it (chain to add subcommands).
	Add(name string, cmd Command[any]) Command[any]
	// Run parses os.Args[1:] and dispatches to the matched command.
	Run(osArgs []string) error
	// Help registers cmd as the help command under the reserved name, so callers
	// never type it themselves. Use instead of Add for the help command.
	Help(cmd Command[any]) Command[any]
}
```

`Config` is a light, serializable DTO (name, version). Config resolution and any
help output codecs are attached separately with the chainable setters:
`cli.NewApp(cfg, cli.GlobalFlags{}).Resolve(config.NewResolver(store, nil)).HelpOutputs(yamlCodec)`.

`Add` returns the command, so subcommands chain:
`app.Add("db", dbCmd).Add("migrate", migrateCmd)`.

`Run` returns sentinel errors `ErrShowingHelp`, `ErrShowingVersion`,
`ErrNoCommands`, `ErrCommandNotFound` (test with `errors.Is`); treat the first two
as clean exits.

## Command

Every command implements `Command[T]`, where `T` is its config struct:

```go
type Command[T any] interface {
	// Name gets (pass "") or sets (pass non-empty) the command name.
	Name(name string) string
	// Add registers a subcommand.
	Add(name string, cmd Command[any])
	// Options returns a pointer to the config struct for flag parsing.
	Options() any
	// Commands returns registered subcommands.
	Commands() []Command[any]
	// Run executes the command with parsed global options and unmatched args.
	Run(options GlobalFlags, unknowns Unknowns) error
	// Validate checks parsed options against the struct's `rules:` tags.
	Validate(options map[string]any) error
	// Help is the one-line summary shown in listings.
	Help() string
	// Description is a longer, multi-line description for detailed help.
	Description() string
	// Examples are usage examples; each is a slice of lines (invocation first,
	// sample output after).
	Examples() [][]string
	// Args are multi-line descriptions for positional args, keyed by index.
	Args() map[int][]string
	// Flags are multi-line descriptions for flags, keyed by the flag as written.
	Flags() map[string][]string
}
```

Embed `BaseCommand[T]` to get everything except `Run` and `Help` for free. It
implements `Name`/`Add`/`Options`/`Commands`/`Validate`, and stubs the four
help-enrichment methods as no-ops, so you override only what you need:

```go
func (c *BaseCommand[T]) Description() string        { return "" }
func (c *BaseCommand[T]) Examples() [][]string       { return nil }
func (c *BaseCommand[T]) Args() map[int][]string     { return nil }
func (c *BaseCommand[T]) Flags() map[string][]string { return nil }
```

Parsed values land on `c.Inputs` (a `*T`) before `Run`, populated from CLI args,
env vars, and `default:` tags.

## Core types

```go
// Config is the serializable app identity (a light DTO). Config resolution and
// help output codecs are attached to the App separately, via the chainable
// app.Resolve(...) and app.HelpOutputs(...) setters.
type Config struct {
	Name    string
	Version string
}

// GlobalFlags are built-in flags available to every command, parsed before
// dispatch and passed to every Run. Add your own fields to extend them.
type GlobalFlags struct {
	Cwd       string `arg:"cwd" short:"c" env:"CWD" help:"Current working directory"`
	Help      bool   `arg:"help" short:"h" env:"HELP" help:"Show help"`
	Version   bool   `arg:"version" short:"v" env:"VERSION" help:"Show version"`
	Verbosity int    `arg:"verbosity" env:"VERBOSITY" help:"Verbosity level (0, 1, 2)"`
	Format    string `arg:"format" help:"Help output format (plain, pretty, md, json, jsonschema)"`
}

// Unknowns carries tokens not matched to any struct field, for pass-through.
type Unknowns struct {
	Args    []string       // positional values with no numeric arg tag
	Options map[string]any // flags not defined on the command
}
```

## Struct tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `arg` | Flag name, or numeric index for a positional arg | `arg:"port"`, `arg:"0"` |
| `short` | Single-char shorthand | `short:"p"` |
| `env` | Environment variable name | `env:"SERVER_PORT"` |
| `help` | One-line description in help output | `help:"Port to listen on"` |
| `default` | Value used when the flag is omitted | `default:"8080"` |
| `rules` | Validation rules (`Validate`) | `rules:"required"`, `rules:"oneof:json,yaml"` |
| `sep` | Separator for `[]T` flags (default `,`) | `sep:"|"` |

Rules combine with `|`: `rules:"required|oneof:json,yaml,toml"`. `oneof` restricts a
field to a fixed set (an enum) and surfaces the allowed values in help and JSON Schema;
pair it with `default:` for a fallback.

A scalar slice field splits a comma-separated value into elements: `--tags=a,b,c`
becomes `["a", "b", "c"]` (same for its env var); override the separator with `sep`.

## Config

Config is optional and fully decoupled: core knows only the `Resolver` seam. Files
and per-command mapping live in the `config` package, which never imports `cli`.
Resolvers compose like middleware - register several and the framework runs them in
order, threading each one's output into the next.

```go
// the only config-shaped thing in core. resolvers run as a chain: each overlays its
// layer on the values from the previous one. the framework seeds struct `default:`
// tags, runs the chain, folds env, then overlays flags (so a typed flag wins).
type Resolver interface {
	Resolve(cmd string, values map[string]any) (map[string]any, error)
}
```

Each `Store` is one config file; build a resolver per store and hand the app the
chain (lowest precedence first). Secrets are just another store:

```go
import "github.com/toaweme/cli/config"

global  := config.NewFileStore("~/.myapp", "config")
project := config.NewFileStore(cwd, "config")
secrets := config.FileSecrets("~/.myapp") // 0600 store named "secrets"

app := cli.NewApp(cli.Config{Name: "myapp"}, cli.GlobalFlags{}).
	Resolve(
		config.NewResolver(global, nil),
		config.NewResolver(project, nil),
		config.NewResolver(secrets, nil),
	)
```

Effective precedence, lowest first: `default:` tags < resolver chain (stores in
registration order) < per-command mapping rules < env (`env:` tag) < flags.

The second `NewResolver` argument is optional per-command field mapping (a `Source`
is a dotted path into the merged config, or a `func() (any, error)`); pass nil for
none. This is where one command inherits another's values:

```go
resolver := config.NewResolver(project, map[string]map[string]config.Source{
	"db migrate": {"steps": "db.steps"},
	"deploy":     {"output": "build.output"}, // deploy inherits build's output dir
})
```

A `Store` is a single config file you can read or write whole, or address one dotted
key within. Inject a store into a command's constructor to use it directly:

```go
global := config.NewFileStore("~/.myapp", "config")
global.Write(defaultConfig)                   // seed ~/.myapp/config.json (whole file)

var cfg AppConfig
global.Read(&cfg)                             // missing file reads as the zero value

global.KeyWrite("build.target", "x86")        // set one dotted key (read-modify-write)
target, _ := global.KeyRead("build.target")   // read one dotted key, nil if absent

var token GitHubToken
secrets.KeyRead("github")                     // secrets, same Store API at 0600
```

The `config.Store` interface (`Read`/`Write`/`Exists` for the whole file,
`KeyRead`/`KeyWrite`/`KeyExists` for a dotted key; codec by file extension) backs each
config; implement it to swap files for memory or a remote store.

## Built-in commands

Import and register the ones you want:

```go
import (
	"github.com/toaweme/cli/commands/help"
	"github.com/toaweme/cli/commands/version"
	"github.com/toaweme/cli/commands/completion"
)

app.Help(help.NewHelpCommand(app.Config, app.Commands, app.OutputFormats)) // help, under the reserved name
app.Add("version", version.NewVersionCommand(app.Config))
app.Add("completion", completion.NewCompletionCommand("myapp"))
```

Help renders in several formats via `--format`: `plain`, `plain-flags`, `pretty`,
`md`, `json`, `jsonschema`. Examples, argument, and flag descriptions you add to a
command show up across all of them.

Register extra output codecs with the chainable `app.HelpOutputs(...)` setter to add
formats. Every extension a codec reports becomes a valid `--format` value (so a YAML
codec accepts both `yml` and `yaml`); the primary `Extension()` is what the
`--format` hint advertises and what the tree is written as. The yaml/toml config
addons satisfy `OutputCodec` structurally, so the core never imports those libraries:

```go
import (
	yamlcodec "github.com/toaweme/cli/config/addons/yaml"
	tomlcodec "github.com/toaweme/cli/config/addons/toml"
)

app := cli.NewApp(cli.Config{Name: "myapp"}, cli.GlobalFlags{}).
	HelpOutputs(yamlcodec.New(), tomlcodec.New())

// myapp help --format yml   (also --format yaml, and --format toml) now work
```

Each codec has a `New(exts ...string)` constructor (`jsoncodec.New`, `yamlcodec.New`,
`tomlcodec.New`). The first extension is the primary, used for writing and as the
`--format` name; a codec may recognize several extensions while writing only the
primary:

```go
jsoncodec.New()                  // recognizes .json
yamlcodec.New()                  // recognizes .yml + .yaml, writes .yml
yamlcodec.New(".yaml")           // writes .yaml, recognizes only .yaml
jsoncodec.New(".json", ".jsonc") // writes .json, also reads .jsonc
```

Pass the codec a store is encoded with to `NewFileStore(dir, name, codec)` (one
codec; a name with an explicit extension is used verbatim, otherwise the codec's
extension is appended). With none it defaults to JSON. Codecs are opt-in and the CLI
does not need any of them - flags, env, `default:` tags, and help's built-in
`--format json` output all work regardless. Pass YAML and the store never touches JSON:

```go
config.NewFileStore(dir, "config")                  // JSON default -> config.json
config.NewFileStore(dir, "config", yamlcodec.New()) // YAML only -> config.yml
```
