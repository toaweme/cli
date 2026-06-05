// Package cli is a small, generics-based framework for building command-line applications
// whose flags, positional arguments, environment bindings, and validation rules are declared once as struct tags.
//
// A command is a struct that embeds [BaseCommand] (parameterized by its config type) and implements Run.
// The config type's fields, tagged with arg/short/env/default/help/rules, define everything the framework needs
// to parse, validate, and document the command:
//
//	type GreetConfig struct {
//		Name  string `arg:"0" env:"GREET_NAME" help:"Name to greet" rules:"required"`
//		Shout bool   `arg:"shout" short:"s" help:"Uppercase the greeting"`
//	}
//
//	type GreetCommand struct {
//		cli.BaseCommand[GreetConfig]
//	}
//
//	func (c *GreetCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
//		fmt.Printf("hello, %s!\n", c.Inputs.Name)
//		return nil
//	}
//
// Build an [App] with [NewApp], register commands with [App.Add], [App.Default], and [App.Help],
// then dispatch os.Args with [App.Run]. The framework merges values from struct defaults,
// the resolver chain (see [Resolver]), environment variables, and parsed flags, in that order
// of increasing precedence, before calling Run.
//
// Help (-h/--help) and version (-V/--version) are built in.
// Run returns the [ErrShowingHelp] / [ErrShowingVersion] sentinels once it has handled those requests itself;
// use [IsRealError] to filter them at the call site.
//
// File-backed configuration, output codecs, the help command, and the docs generator live in sub-packages
// so the core stays dependency-light. See the runnable programs under examples/ for complete applications.
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

// Built-in global flag arg names. Go struct tags must be string literals, so each name is written once
// in the matching GlobalFlags `arg:` tag and once here; that is the only duplication, and
// Test_GlobalFlags_ArgNames fails if the two drift apart. Dispatch, parsing, and help reference these consts
// instead of retyping the literal, so the flag names the framework reasons about live in one place.
const (
	argHelp       = "help"
	argHelpValues = "help-values"
	argHelpFormat = "help-format"
	argVersion    = "version"
)

// IsRealError reports whether err is a genuine failure worth surfacing, as opposed to a clean-exit sentinel
// the framework returns once it has already handled the request itself (printing help or the version).
// It returns false for nil and for the ErrShowingHelp / ErrShowingVersion sentinels, and true for anything else,
// so a caller need not enumerate the sentinels by hand:
//
//	if err := app.Run(os.Args[1:]); cli.IsRealError(err) {
//		fmt.Fprintf(os.Stderr, "error: %v\n", err)
//		os.Exit(1)
//	}
func IsRealError(err error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, ErrShowingHelp) && !errors.Is(err, ErrShowingVersion)
}

// App is the top-level CLI application. It owns the command set, global flags,
// and an ordered chain of config Resolvers, and dispatches osArgs to the matched command.
type App interface {
	// Commands returns the registered top-level commands.
	Commands() []Command[any]
	// Config returns the app identity (the serializable DTO).
	Config() Config
	// OutputFormats returns the registered help output codecs, in registration order.
	OutputFormats() []OutputCodec
	// Resolve appends config Resolvers to the chain used to populate each command's Options() before Run,
	// and returns the app for chaining. Resolvers run in the order registered (across all Resolve calls),
	// lowest precedence first, then env, then flags. With none registered, only env and flags apply.
	Resolve(resolvers ...Resolver) App
	// HelpOutputs registers additional help output codecs (e.g. the yaml/toml addons) and returns the app for chaining.
	// Each codec's name, derived from its Extension (".yml" -> "yml"), becomes a valid --help-format value
	// and is advertised in help.
	HelpOutputs(formats ...OutputCodec) App
	// Default sets the command run when no arguments are given; it returns cmd.
	Default(cmd Command[any]) Command[any]
	// Add registers cmd under name and returns it, so subcommands chain off the result.
	Add(name string, cmd Command[any]) Command[any]
	// Run parses osArgs and dispatches to the matched command.
	// Help and version requests surface as the ErrShowingHelp/ErrShowingVersion sentinels.
	Run(osArgs []string) error
	// Help registers cmd as the command that renders help, so callers never have to know the reserved name.
	// Use it instead of Add: app.Help(help.NewHelpCommand(...)).
	Help(cmd Command[any]) Command[any]
}

type app struct {
	config         Config
	resolvers      []Resolver
	formats        []OutputCodec
	globalFlags    *GlobalFlags
	commands       []Command[any]
	defaultCommand Command[any]
}

var _ App = (*app)(nil)

// NewApp creates an application from config (the serializable identity and merge strategy)
// and the default values for global flags. Attach config Resolvers and output Formats with the chainable setters
// (Resolve and HelpOutputs), then register commands with Add, Default, and Help, and dispatch with Run.
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

// Resolve appends config Resolvers to the chain and returns the app for chaining.
func (c *app) Resolve(resolvers ...Resolver) App {
	c.resolvers = append(c.resolvers, resolvers...)

	return c
}

// HelpOutputs registers additional help output codecs and returns the app for chaining.
func (c *app) HelpOutputs(formats ...OutputCodec) App {
	c.formats = append(c.formats, formats...)

	return c
}

// Default registers the command Run dispatches to when invoked with no arguments. It returns cmd.
func (c *app) Default(cmd Command[any]) Command[any] {
	c.defaultCommand = cmd

	return cmd
}

// Add registers cmd under name and returns it. Attach subcommands by calling Add on the returned command:
// app.Add("db", db).Add("migrate", migrate).
func (c *app) Add(name string, cmd Command[any]) Command[any] {
	cmd.Name(name)
	c.commands = append(c.commands, cmd)

	return cmd
}

// Run parses osArgs (typically os.Args[1:]): it binds Config into the command tree,
// resolves and validates global options, then dispatches to the matched command.
// A --help or --version request, and an unknown command, surface as the
// ErrShowingHelp / ErrShowingVersion sentinels - test with errors.Is and treat them as clean exits.
func (c *app) Run(osArgs []string) error {
	if len(c.commands) < 1 {
		return ErrNoCommands
	}

	if len(osArgs) > 0 && osArgs[0] == "__complete" {
		c.handleComplete(osArgs[1:])
		return nil
	}

	globalFlags, globalUnknownOpts := c.getGlobalFlags(osArgs)

	// --help-format spans the built-in formats plus any output codecs registered via HelpOutputs,
	// so it is validated here (against the full set) rather than by the static oneof rule on
	// GlobalFlags.HelpFormat, which only knows the built-ins.
	if err := c.validateFormat(globalFlags[argHelpFormat]); err != nil {
		return err
	}

	err := mapStructToOptions(c.globalFlags, globalFlags, argHelpFormat)
	if err != nil {
		return fmt.Errorf("failed to update global options struct: %w", err)
	}

	// -h/--help and -V/--version must trigger regardless of position,
	// even directly after a value-taking flag the global parse would let swallow them.
	// Detect them with a direct scan and OR into whatever the parse already set.
	if !c.globalFlags.Help {
		c.globalFlags.Help = boolFlagRequested(osArgs, globalBoolFlagNames(argHelp))
	}
	if !c.globalFlags.Version {
		c.globalFlags.Version = boolFlagRequested(osArgs, globalBoolFlagNames(argVersion))
	}
	if !c.globalFlags.HelpValues {
		c.globalFlags.HelpValues = boolFlagRequested(osArgs, globalBoolFlagNames(argHelpValues))
	}
	// --help-values is a help mode, so it implies --help.
	if c.globalFlags.HelpValues {
		c.globalFlags.Help = true
	}

	if c.globalFlags.Version {
		c.printVersion()
		return ErrShowingVersion
	}

	// commandArgs holds the osArgs that are commands
	// allArgs holds the osArgs that are not commands
	command, commandArgs, allArgs, err := c.matchCommandByArgs(osArgs)
	if err != nil {
		// no command matched. with a default command set (and not an explicit --help), dispatch to it
		// with the args parsed against it, so `app --flag` behaves like `app <default> --flag`
		// and bare `app` runs the default. otherwise show help.
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

	// commandOptions holds the parsed flags; fold in positions keyed by index
	// so the two together form the highest-precedence flags layer.
	flags := commandOptions
	for i, arg := range cmdArgs {
		flags[fmt.Sprintf("%d", i)] = arg
	}

	// if --help is passed, show help
	if c.globalFlags.Help {
		// with --help-values, populate the matched command's struct so help can show resolved values.
		// Skip validation (the resolve-only path) so --help still works when required inputs are absent.
		if c.globalFlags.HelpValues {
			if err := c.resolveCommandConfig(command, strings.Join(commandArgs, " "), flags); err != nil {
				return err
			}
		}

		err := c.runHelp(commandArgs, globalUnknownOpts)
		if err != nil {
			return fmt.Errorf("failed to run help: %w", err)
		}

		return ErrShowingHelp
	}

	// cmdPath is the matched command path (e.g. "db migrate"),
	// handed to the resolver so it can apply per-command rules.
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

func (c *app) matchCommandByArgs(args []string) (Command[any], []string, []string, error) {
	var command Command[any]
	var commandNameIndexes []int

	for a := 0; a < len(args); a++ {
		// previous arg is a command
		// assert if this arg is a sub command
		if command != nil {
			subCommand := c.matchCommandByName(args[a], command.Commands())
			if subCommand != nil {
				command = subCommand
				commandNameIndexes = append(commandNameIndexes, a)
				continue
			}

			// no sub command found, but
			// we have a command already so let's keep it and try further args
			break
		}

		cmd := c.matchCommandByName(args[a], c.commands)
		if cmd != nil {
			command = cmd
			commandNameIndexes = append(commandNameIndexes, a)
			continue
		}
	}

	if command == nil {
		return nil, nil, nil, ErrCommandNotFound
	}

	// create a new slice that excludes the command args
	// we don't need the command args anymore
	allOtherArgs := make([]string, 0)
	commandNameArgs := make([]string, 0)
	for i := 0; i < len(args); i++ {
		if exists(commandNameIndexes, i) {
			commandNameArgs = append(commandNameArgs, args[i])
			continue
		}
		allOtherArgs = append(allOtherArgs, args[i])
	}

	return command, commandNameArgs, allOtherArgs, nil
}

func exists(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}

	return false
}

func (c *app) matchCommandByName(arg string, commands []Command[any]) Command[any] {
	var command Command[any]
	for i := 0; i < len(commands); i++ {
		cmd := commands[i]
		if cmd.Name("") == arg {
			command = cmd
			break
		}
	}

	return command
}

// loadCommandConfig populates command.Options() from ordered layers and then validates the result.
// cmd is the matched command path (e.g. "db migrate"); flags are the explicit CLI inputs
// (parsed flags plus positionals keyed by index), and are the highest-precedence layer.
//
// The layers, lowest first:
//  1. struct `default:` tags
//  2. the Resolver chain, each overlaying its layer on the previous (files, mapping)
//  3. env, folded in after the chain so it beats files
//  4. flags, applied as a separate pass so a typed flag always wins
//
// Applying the merged map and the flags as distinct structs.Set passes is what makes flags beat env:
// within a single pass, an `env:` tag match short-circuits, so a merged map cannot express "flags over env".
// Validation runs after the merge so `required` is satisfied by config- or default-provided values, not just flags.
func (c *app) loadCommandConfig(command Command[any], cmd string, flags map[string]any) error {
	if err := c.resolveCommandConfig(command, cmd, flags); err != nil {
		return err
	}

	// validate against the explicit inputs the user supplied; rules like `required` fall back to the
	// now-populated field values, so values sourced from config or defaults still satisfy them.
	validateInputs := map[string]any{}
	env(validateInputs)
	for k, v := range flags {
		validateInputs[k] = v
	}
	if err := command.Validate(validateInputs); err != nil {
		return fmt.Errorf("failed to validate command %q: %w", command.Name(""), err)
	}

	return nil
}

// resolveCommandConfig populates command.Options() from the ordered layers without validating.
// It is the merge half of loadCommandConfig, shared with the --help-values path,
// which needs the resolved field values to display but must not fail when a required input is absent
// (the user only asked for help).
func (c *app) resolveCommandConfig(command Command[any], cmd string, flags map[string]any) error {
	inputs := command.Options()
	manager := structs.New(inputs, structs.DefaultRules, structs.WithTags(defaultTags...))

	// run the resolver chain, threading each one's output into the next.
	values := map[string]any{}
	for _, resolver := range c.resolvers {
		next, err := resolver.Resolve(cmd, values)
		if err != nil {
			return fmt.Errorf("failed to resolve config for command %q: %w", command.Name(""), err)
		}
		if next != nil {
			values = next
		}
	}

	// env beats the resolver layers; flags (applied below) still win over env.
	env(values)

	// defaults + resolved layer; an empty map still applies struct `default:` tags.
	if err := manager.Set(values); err != nil {
		return fmt.Errorf("failed to apply resolved config for command %q: %w", command.Name(""), err)
	}

	// flags win, as a separate pass.
	if len(flags) > 0 {
		if err := manager.Set(flags); err != nil {
			return fmt.Errorf("failed to apply flags for command %q: %w", command.Name(""), err)
		}
	}

	return nil
}
