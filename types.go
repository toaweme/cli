package cli

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
	Run(options GlobalFlags, unknowns Unknowns) error
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
}

// Resolver contributes values to a command's Options() before Run. Resolvers
// compose like middleware: the framework registers any number on the App, then runs
// them in order, threading each one's output into the next. After the chain it folds
// in env, then overlays parsed flags, so a typed flag always wins. File-backed
// resolution lives in the config package (config.NewResolver, one per Store), which
// satisfies this interface structurally so core never imports config.
type Resolver interface {
	// Resolve overlays this resolver's layer onto values, the map accumulated by
	// earlier resolvers in the chain, and returns it. cmd is the command path
	// (space-joined, e.g. "db migrate"). The first resolver receives an empty map.
	Resolve(cmd string, values map[string]any) (map[string]any, error)
}

// Config is the serializable application identity: plain values only, so it stays
// a light DTO that round-trips through json/yaml. Config resolution is attached to
// the App separately via the App.Resolve setter; output codecs via App.Formats.
type Config struct {
	// Name is the application binary name, shown in help and usage output.
	Name string `json:"name" yaml:"name"`
	// Version is the semantic version string shown by the version command.
	Version string `json:"version" yaml:"version"`
}

// OutputCodec renders help output for a custom --format value. It is satisfied
// structurally by the yaml/toml/json config addon codecs (which also implement
// config.Codec), so registering one never pulls an encoding library into the core.
// The format name a user types is derived from Extension by trimming the leading
// dot (".yml" -> "yml").
type OutputCodec interface {
	// Marshal encodes v into bytes.
	Marshal(v any) ([]byte, error)
	// Extension returns the file extension for this codec (e.g. ".yml").
	Extension() string
}

// GlobalFlags are built-in flags available to every command.
// These are parsed before command dispatch and passed to every command's Run method.
type GlobalFlags struct {
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
