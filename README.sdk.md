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
	// Config returns the app identity and merge policy (the serializable DTO).
	Config() Config
	// OutputFormats returns the registered help output codecs.
	OutputFormats() []OutputCodec
	// Store attaches the config storage and returns the app for chaining.
	Store(store Storage) App
	// Formats registers additional help output codecs and returns the app for chaining.
	Formats(formats ...OutputCodec) App
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

`Config` is a light, serializable DTO (name, version, merge strategy). The runtime
objects, the config `Store` and any output `Formats`, are attached separately with
the chainable setters: `cli.NewApp(cfg, cli.GlobalFlags{}).Store(store).Formats(yamlCodec)`.

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
// Config is the serializable app identity and merge policy (a light DTO). The
// config Store and output Formats are attached to the App separately, via the
// chainable app.Store(...) and app.Formats(...) setters.
type Config struct {
	Name    string
	Version string
	Merge   MergeStrategy // optional; how command options are populated (see below)
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

## Storage

Storage is optional. Build one with `NewFileStorage(...)` and attach it with the
chainable `app.Store(...)` setter; the app binds it into the command tree, so any
command can read it via `BaseCommand.Store()`. You can also pass it to commands via
their constructors (explicit injection, no auto-wiring):

```go
type Storage interface {
	config.Store                  // primary kv store promoted: Save / Load / Delete / Exists by key
	Secrets() config.Store        // 0600 files under <dir>/secrets
	Dir() string                  // resolved base directory
	// Resolve merges target from layers, lowest precedence first: struct `default:`
	// tags, config stores (home, then per-project), env (when opts.Env, matched
	// via `env:` tags), then opts.Flags.
	Resolve(target any, opts LoadOptions) error
}

type LoadOptions struct {
	Key   string         // store key per layer; defaults to "config"
	Env   bool           // fold in environment variables
	Flags map[string]any // highest-precedence overrides (e.g. parsed CLI flags)
}

type FileStorage struct {
	Dir        string         // base dir; leading "~" expands; defaults to "~/.<Name>"
	Name       string         // app name; derives the default dir
	PerProject bool           // also use a ".<Name>" dir found by walking up
	Codecs     []config.Codec // JSON built in; add yaml.Codec / toml.Codec
}
```

```go
store := cli.NewFileStorage(cli.FileStorage{Name: "myapp", PerProject: true})

store.Save("config", current)                  // direct key access (promoted)
var settings AppSettings
err := store.Resolve(&settings, cli.LoadOptions{Env: true, Flags: parsedFlags}) // layered
```

To back `Storage` with something other than files, implement the `Storage`
interface directly.

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

Register extra output codecs with the chainable `app.Formats(...)` setter to add
formats. Each codec's name, derived from its `Extension()` (`.yaml` -> `yaml`),
becomes a valid `--format` value, renders the command tree (the same data `json`
emits), and is advertised in the `--format` hint. The yaml/toml config addons
satisfy `OutputCodec` structurally, so the core never imports those libraries:

```go
import (
	yamlcodec "github.com/toaweme/cli/config/addons/yaml"
	tomlcodec "github.com/toaweme/cli/config/addons/toml"
)

app := cli.NewApp(cli.Config{Name: "myapp"}, cli.GlobalFlags{}).
	Formats(&yamlcodec.Codec{}, &tomlcodec.Codec{})

// myapp help --format yaml   (and --format toml) now work
```
