package cli

import (
	"github.com/toaweme/cli/config"
)

// App is the top-level CLI application. It owns the command set, global options,
// and optional config Storage, and dispatches osArgs to the matched command.
type App interface {
	// Commands returns the registered top-level commands.
	Commands() []Command[any]
	// Config returns the app identity and configuration.
	Config() Config
	// Default sets the command run when no arguments are given; it returns cmd.
	Default(cmd Command[any]) Command[any]
	// Add registers cmd under name and returns it, so subcommands chain off the result.
	Add(name string, cmd Command[any]) Command[any]
	// Run parses osArgs and dispatches to the matched command. Help and version
	// requests surface as the ErrShowingHelp/ErrShowingVersion sentinels.
	Run(osArgs []string) error
	// Help registers cmd as the command that renders help, so callers never have
	// to know the reserved name. Use it instead of Add: app.Help(help.NewHelpCommand(...)).
	Help(cmd Command[any]) Command[any]
}

// Command is the interface every CLI command must implement.
// T is the config struct type whose fields define the command's flags and positional args.
type Command[T any] interface {
	// Name gets or sets the command name. Pass "" to get, non-empty to set.
	Name(name string) string
	// Add registers a subcommand under this command.
	Add(name string, cmd Command[any])
	// Options returns a pointer to the config struct for flag parsing.
	Options() any
	// Commands returns the list of registered subcommands.
	Commands() []Command[any]
	// Run executes the command logic with parsed global options and unknown args.
	Run(options GlobalOptions, unknowns Unknowns) error
	// Validate checks the parsed options map against struct validation rules.
	Validate(options map[string]any) error
	// Help returns a short one-line description shown in command listings.
	Help() string
	// Description returns a longer, multi-line description shown in detailed and
	// agent help. Help stays the one-line listing summary; Description carries the
	// richer body (paragraphs, install instructions, ...). Empty by default.
	Description() string
	// Examples returns usage examples shown in detailed and agent help. Each
	// example is a slice of lines: the first is the invocation, any following
	// lines are sample output shown beneath it. Nil by default.
	Examples() [][]string
	// Args returns multi-line descriptions for positional arguments, keyed by
	// zero-based position. Augments the single-line `help:` tag. Nil by default.
	Args() map[int][]string
	// Flags returns multi-line descriptions for flags, keyed by the flag as
	// written (e.g. "--query, -q"). Augments the single-line `help:` tag. Nil by
	// default.
	Flags() map[string][]string
	// ConfigStrategy selects how this command's Options() are populated before
	// Run and, optionally, how its fields map onto the global config.
	//
	// The MergeStrategy overrides the app-wide Config.Merge: return MergeInherit
	// (the BaseCommand default) to use the app default, or MergeEnvFlags /
	// MergeLayered to force one (e.g. opt out of an app-wide MergeLayered).
	//
	// The ConfigMapping remaps fields onto arbitrary paths in the global config,
	// for when the config is shaped differently than the command struct (e.g. a
	// `http.host` block feeding a flat Host field). It only takes effect under
	// MergeLayered; nil means match fields by their own tags, as usual.
	ConfigStrategy() (MergeStrategy, ConfigMapping)
}

// ConfigMapping remaps a command's fields onto paths in the global config store,
// for the MergeLayered strategy. Each entry is {field: "dotted.path"}: the key is
// matched the same way any input is (the field's arg/short/json/yaml tag or field
// name), the value is a path into the decoded config (nested maps split on ".").
// It is a plain map so mappings compose by merging. An explicit mapping wins over
// the field's incidental tag match for the config-store layer; env and flags are
// unaffected and still override per field.
type ConfigMapping map[string]string

// MergeStrategy selects how a command's Options() struct is populated before Run.
// Set an app-wide default with Config.Merge; a command overrides it (including
// opting out) via Command.ConfigStrategy. The zero value, MergeInherit, defers:
// on a command it means "use the app default", and as the app default it resolves
// to MergeEnvFlags - so an app and command that set nothing keep the original
// env+flags behavior.
type MergeStrategy int

const (
	// MergeInherit defers to the next level: a command inherits Config.Merge, and
	// an unset Config.Merge resolves to MergeEnvFlags. It is the zero value.
	MergeInherit MergeStrategy = iota
	// MergeEnvFlags populates Options() from struct `default:` tags, then
	// environment variables, then CLI flags. No config store is read. This is the
	// effective default and matches the framework's original behavior.
	MergeEnvFlags
	// MergeLayered additionally folds the config store in between defaults and
	// env: `default:` -> Store config -> env -> flags. Requires Config.Store;
	// without one it degrades to MergeEnvFlags.
	MergeLayered
)

// Config configures the application identity and optional storage.
type Config struct {
	// Name is the application binary name, shown in help and usage output.
	Name string
	// Version is the semantic version string shown by the version command.
	Version string
	// Store is the optional configuration storage. It is available to commands
	// for direct access via constructor injection, and is the source the
	// MergeLayered strategy reads from. Setting it does NOT, on its own, change
	// how command options are populated - that is governed by Merge.
	Store Storage
	// Merge is the app-wide default strategy for populating each command's
	// Options() before Run. The zero value (MergeInherit) resolves to
	// MergeEnvFlags: env + flags only, the original behavior. Set it to
	// MergeLayered to fold the config Store in by default. Individual commands
	// override it via Command.ConfigStrategy.
	Merge MergeStrategy
	// Formats registers additional help output codecs (e.g. the yaml/toml config
	// addons). Each one's name, derived from its Extension (".yaml" -> "yaml"),
	// becomes a valid --format value and is advertised in help. The core stays
	// free of the underlying encoding libraries: codecs satisfy OutputCodec
	// structurally, so nothing here imports yaml or toml.
	Formats []OutputCodec
}

// OutputCodec renders help output for a custom --format value. It is satisfied
// structurally by the yaml/toml/json config addon codecs (which also implement
// config.Codec), so registering one never pulls an encoding library into the core.
// The format name a user types is derived from Extension by trimming the leading
// dot (".yaml" -> "yaml").
type OutputCodec interface {
	// Marshal encodes v into bytes.
	Marshal(v any) ([]byte, error)
	// Extension returns the file extension for this codec (e.g. ".yaml").
	Extension() string
}

// GlobalOptions are built-in flags available to every command.
// These are parsed before command dispatch and passed to every command's Run method.
type GlobalOptions struct {
	// Cwd overrides the working directory for the command.
	Cwd string `arg:"cwd" short:"c" env:"CWD" help:"Current working directory"`
	// Help triggers help display instead of running the matched command.
	Help bool `arg:"help" short:"h" env:"HELP" help:"Show help"`
	// Version prints the application version and exits.
	Version bool `arg:"version" short:"v" env:"VERSION" help:"Show version"`
	// Verbosity controls log output level (0=quiet, 1=normal, 2=verbose).
	Verbosity int `arg:"verbosity" env:"VERBOSITY" help:"Verbosity level (0, 1, 2)"`
	// Format controls help output. The allowed values come from the oneof rule,
	// which also drives the "(one of: ...)" hint shown in help.
	Format string `arg:"format" help:"Help output format" rules:"oneof:plain,plain-flags,pretty,md,json,jsonschema"`
}

// Unknowns holds arguments and options that were not matched to any defined field.
// Commands receive these to support pass-through or dynamic flag handling.
type Unknowns struct {
	// Args are positional arguments not matched to numbered struct tags.
	Args []string
	// Options are key-value flags not defined in the command's config struct.
	Options map[string]any
}

// Storage is the configuration accessor available to commands. It wraps a
// primary key-value store, a separate secrets store (0600-permission files),
// and a layered Load that merges a typed config from several sources.
//
// NewFileStorage returns the file-backed implementation; implement Storage
// directly to back it with something else (in-memory, remote, ...).
type Storage interface {
	// Store returns the primary key-value store for direct Load/Save by key.
	Store() config.Store
	// Secrets returns the secrets store, backed by 0600-permission files under
	// a "secrets" subdirectory of the config directory.
	Secrets() config.Store
	// Dir returns the resolved base directory backing this storage.
	Dir() string
	// Load merges configuration into target from layered sources, lowest
	// precedence first: struct `default:` tags, then each config store (home,
	// then any per-project store discovered by walking up), then environment
	// variables (when opts.Env is set, matched via `env:` tags), then opts.Flags.
	// Later layers override earlier ones field by field; absent layers are skipped.
	Load(target any, opts LoadOptions) error
}

// LoadOptions configures a layered load. See Storage.Load.
type LoadOptions struct {
	// Key is the store key (file name) read from each config store layer.
	// Defaults to "config" when empty.
	Key string
	// Env, when set, folds environment variables in as a layer above the config
	// stores. Fields are matched by their `env:` tag.
	Env bool
	// Flags is the highest-precedence layer, typically the parsed CLI options.
	// Keys are matched the same way struct inputs are (arg/short/json/yaml tags).
	Flags map[string]any
	// Mapping optionally remaps target fields onto arbitrary paths in the config
	// store layers (not env/flags). See ConfigMapping.
	Mapping ConfigMapping
}
