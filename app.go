package cli

import (
	"errors"
	"fmt"

	"github.com/toaweme/structs"
)

var ErrCommandNotFound = errors.New("command not found")
var ErrNoCommands = errors.New("no commands registered")
var ErrDisplaySubCommands = errors.New("print sub commands")
var ErrShowingHelp = errors.New("showing help")
var ErrShowingVersion = errors.New("showing version")

const helpCommand = "help"

type app struct {
	config         Config
	globalOptions  *GlobalOptions
	commands       []Command[any]
	defaultCommand Command[any]
}

// NewApp creates an application from config (identity, optional Store, merge
// strategy, and output Formats) and the default values for global options. Register
// commands with Add, Default, and Help, then dispatch with Run.
func NewApp(config Config, opts GlobalOptions) App {
	return newApp(config, opts)
}

func newApp(config Config, opts GlobalOptions) *app {
	return &app{
		config:        config,
		globalOptions: &opts,
		commands:      make([]Command[any], 0),
	}
}

var _ App = (*app)(nil)

// Commands returns the registered top-level commands, in registration order.
func (c *app) Commands() []Command[any] {
	return c.commands
}

// Config returns the application's Config (name, version, Store, Merge, Formats).
func (c *app) Config() Config {
	return c.config
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

	// hand every command (and subcommand) the app Config so it can read global
	// configuration via BaseCommand.Config()/Store(); ordering-independent, unlike
	// binding at registration time.
	c.bindConfigTree()

	if len(osArgs) > 0 && osArgs[0] == "__complete" {
		c.handleComplete(osArgs[1:])
		return nil
	}

	globalOptions, globalUnknownOpts := c.getGlobalOptions(osArgs)

	// --format spans the built-in formats plus any output codecs registered in
	// Config.Formats, so it is validated here (against the full set) rather than by
	// the static oneof rule on GlobalOptions.Format, which only knows the built-ins.
	if err := c.validateFormat(globalOptions["format"]); err != nil {
		return err
	}

	err := mapStructToOptions(c.globalOptions, globalOptions, "format")
	if err != nil {
		return fmt.Errorf("failed to update global options struct: %w", err)
	}

	if c.globalOptions.Version {
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
		if errors.Is(err, ErrCommandNotFound) && c.defaultCommand != nil && !c.globalOptions.Help {
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
	if c.globalOptions.Help {
		err := c.runHelp(commandArgs, globalUnknownOpts)
		if err != nil {
			return fmt.Errorf("failed to run help: %w", err)
		}

		return ErrShowingHelp
	}

	if err := c.loadCommandConfig(command, flags); err != nil {
		return err
	}

	err = command.Run(*c.globalOptions, unknowns)
	if err != nil {
		if errors.Is(err, ErrDisplaySubCommands) {
			return c.runHelp(commandArgs, globalUnknownOpts)
		}
		return fmt.Errorf("failed to run command %q: %w", command.Name(""), err)
	}

	return nil
}
