package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/toaweme/structs"
)

var ErrCommandNotFound = errors.New("command not found")
var ErrNoCommands = errors.New("no commands registered")
var ErrDisplaySubCommands = errors.New("print sub commands")
var ErrShowingHelp = errors.New("showing help")
var ErrShowingVersion = errors.New("showing version")

const helpCommand = "help"

// App is the top-level CLI application. It owns the command set, global flags,
// and an optional config Resolver, and dispatches osArgs to the matched command.
//
// Config carries only the serializable identity (name, version); config resolution
// and any output Formats are attached separately via the chainable Resolve and
// Formats setters:
//
//	app := cli.NewApp(cli.Config{Name: "app"}, cli.GlobalFlags{}).
//		Resolve(config.NewFileResolver(cfg)).
//		Formats(yamlCodec, tomlCodec)
type App interface {
	// Commands returns the registered top-level commands.
	Commands() []Command[any]
	// Config returns the app identity (the serializable DTO).
	Config() Config
	// OutputFormats returns the registered help output codecs, in registration order.
	OutputFormats() []OutputCodec
	// Resolve attaches the config Resolver used to populate each command's Options()
	// before Run, and returns the app for chaining. When unset, ResolverDefault
	// (env + flags, no files) is used.
	Resolve(resolver Resolver) App
	// Formats registers additional help output codecs (e.g. the yaml/toml addons)
	// and returns the app for chaining. Each codec's name, derived from its Extension
	// (".yml" -> "yml"), becomes a valid --format value and is advertised in help.
	Formats(formats ...OutputCodec) App
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

type app struct {
	config         Config
	resolver       Resolver
	formats        []OutputCodec
	globalFlags    *GlobalFlags
	commands       []Command[any]
	defaultCommand Command[any]
}

var _ App = (*app)(nil)

// NewApp creates an application from config (the serializable identity and merge
// strategy) and the default values for global flags. Attach a config Store and any
// output Formats with the chainable setters, then register commands with Add,
// Default, and Help, and dispatch with Run.
func NewApp(config Config, opts GlobalFlags) App {
	return &app{
		config:      config,
		globalFlags: &opts,
		commands:    make([]Command[any], 0),
	}
}

// Commands returns the registered top-level commands, in registration order.
func (c *app) Commands() []Command[any] {
	return c.commands
}

// Config returns the application's Config (name, version, merge strategy).
func (c *app) Config() Config {
	return c.config
}

// OutputFormats returns the registered help output codecs, in registration order.
func (c *app) OutputFormats() []OutputCodec {
	return c.formats
}

// Resolve attaches the config Resolver and returns the app for chaining.
func (c *app) Resolve(resolver Resolver) App {
	c.resolver = resolver

	return c
}

// Formats registers additional help output codecs and returns the app for chaining.
func (c *app) Formats(formats ...OutputCodec) App {
	c.formats = append(c.formats, formats...)

	return c
}

// Default registers the command Run dispatches to when invoked with no arguments.
// It returns cmd.
func (c *app) Default(cmd Command[any]) Command[any] {
	c.defaultCommand = cmd

	return cmd
}

// Add registers cmd under name and returns it. Attach subcommands by calling Add on
// the returned command: app.Add("db", db).Add("migrate", migrate).
func (c *app) Add(name string, cmd Command[any]) Command[any] {
	cmd.Name(name)
	c.commands = append(c.commands, cmd)

	return cmd
}

// Run parses osArgs (typically os.Args[1:]): it binds Config into the command tree,
// resolves and validates global options, then dispatches to the matched command. A
// --help or --version request, and an unknown command, surface as the ErrShowingHelp
// / ErrShowingVersion sentinels - test with errors.Is and treat them as clean exits.
func (c *app) Run(osArgs []string) error {
	if len(c.commands) < 1 {
		return ErrNoCommands
	}

	if len(osArgs) > 0 && osArgs[0] == "__complete" {
		c.handleComplete(osArgs[1:])
		return nil
	}

	globalFlags, globalUnknownOpts := c.getGlobalFlags(osArgs)

	// --format spans the built-in formats plus any output codecs registered in
	// Config.Formats, so it is validated here (against the full set) rather than by
	// the static oneof rule on GlobalFlags.Format, which only knows the built-ins.
	if err := c.validateFormat(globalFlags["format"]); err != nil {
		return err
	}

	err := mapStructToOptions(c.globalFlags, globalFlags, "format")
	if err != nil {
		return fmt.Errorf("failed to update global options struct: %w", err)
	}

	if c.globalFlags.Version {
		c.printVersion()
		return ErrShowingVersion
	}

	// commandArgs holds the osArgs that are commands
	// allArgs holds the osArgs that are not commands
	command, commandArgs, allArgs, err := c.matchCommandByArgs(osArgs)
	if err != nil {
		// no command matched. with a default command set (and not an explicit
		// --help), dispatch to it with the args parsed against it, so `app --flag`
		// behaves like `app <default> --flag` and bare `app` runs the default.
		// otherwise show help.
		if errors.Is(err, ErrCommandNotFound) && c.defaultCommand != nil && !c.globalFlags.Help {
			command = c.defaultCommand
			commandArgs = nil
			allArgs = osArgs
		} else {
			helpErr := c.runHelp(commandArgs, globalUnknownOpts)
			if helpErr != nil {
				return fmt.Errorf("failed to run help: %w", helpErr)
			}
			return fmt.Errorf("%w: %w", err, ErrShowingHelp)
		}
	}

	commandInputs := command.Options()
	commandFields, err := structs.GetStructFields(commandInputs, nil, structs.DefaultEncodingTags)
	if err != nil {
		return fmt.Errorf("failed to get struct fields: %w", err)
	}

	// cmdArgs are the args defined as numeric tags in the struct e.g. `arg:"0"`
	// cmdUnknownArgs are the args that are not defined in the struct
	// commandOptions are the options defined in the struct e.g. `arg:"cwd"`
	// cmdUnknownOptions are the options that are not defined in the struct
	cmdArgs, cmdUnknownArgs, commandOptions, cmdUnknownOptions := getCommandArgs(allArgs, commandFields)
	unknowns := Unknowns{
		Args:    cmdUnknownArgs,
		Options: cmdUnknownOptions,
	}

	// commandOptions holds the parsed flags; fold in positionals keyed by index
	// so the two together form the highest-precedence flags layer.
	flags := commandOptions
	for i, arg := range cmdArgs {
		flags[fmt.Sprintf("%d", i)] = arg
	}

	// if --help is passed, show help
	if c.globalFlags.Help {
		err := c.runHelp(commandArgs, globalUnknownOpts)
		if err != nil {
			return fmt.Errorf("failed to run help: %w", err)
		}

		return ErrShowingHelp
	}

	// cmdPath is the matched command path (e.g. "db migrate"), handed to the
	// resolver so it can apply per-command rules.
	cmdPath := strings.Join(commandArgs, " ")
	if err := c.loadCommandConfig(command, cmdPath, flags); err != nil {
		return err
	}

	err = command.Run(*c.globalFlags, unknowns)
	if err != nil {
		if errors.Is(err, ErrDisplaySubCommands) {
			return c.runHelp(commandArgs, globalUnknownOpts)
		}
		return fmt.Errorf("failed to run command %q: %w", command.Name(""), err)
	}

	return nil
}
