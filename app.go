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
//	func (c *GreetCommand) Run(_ cli.GlobalFlags) error {
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
// Every command is strict. A command line that cannot run as written is a [UsageError], which [ExitCode] maps to status 2.
//
// File-backed configuration, output codecs, the help command, and the docs generator live in sub-packages
// so the core stays dependency-light. See the runnable programs under examples/ for complete applications.
package cli

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/toaweme/structs"
)

// ErrCommandNotFound is returned when no registered command matches the args.
var ErrCommandNotFound = errors.New("command not found")

// ErrUsage is matched by every [UsageError].
var ErrUsage = errors.New("invalid usage")

// ErrUnknownFlag is the cause of a [UsageError] for a flag the command does not declare.
var ErrUnknownFlag = errors.New("unknown flag")

// ErrInvalidFlagValue is the cause of a [UsageError] for a flag value that is missing or does not convert.
var ErrInvalidFlagValue = errors.New("invalid flag value")

// ErrUnknownArg is the cause of a [UsageError] for a positional argument the command does not declare.
var ErrUnknownArg = errors.New("unexpected argument")

// ErrNoCommands is returned when the app has no commands registered.
var ErrNoCommands = errors.New("no commands registered")

// ErrDisplaySubCommands signals that sub commands should be printed.
var ErrDisplaySubCommands = errors.New("print sub commands")

// ErrShowingHelp signals that help output was shown instead of running a command.
var ErrShowingHelp = errors.New("showing help")

// ErrShowingVersion signals that the version was shown instead of running a command.
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

// ExitCode maps the error returned by Run to a process exit status.
// It is 2 for a [UsageError], 1 for any other real error and 0 otherwise.
func ExitCode(err error) int {
	if errors.Is(err, ErrUsage) {
		return 2
	}
	if IsRealError(err) {
		return 1
	}

	return 0
}

// App is the top-level CLI application. It owns the command set, global flags,
// and an ordered chain of config Resolvers, and dispatches osArgs to the matched command.
type App interface {
	// Commands returns the registered top-level commands.
	Commands() []Command[any]
	// DefaultCommand returns the command registered via Default, or nil when none is set.
	DefaultCommand() Command[any]
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

// DefaultCommand returns the command registered via Default, or nil when none is set.
func (c *app) DefaultCommand() Command[any] {
	return c.defaultCommand
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
// A --help or --version request surfaces as the ErrShowingHelp / ErrShowingVersion sentinels,
// which are clean exits. A command line that cannot run as written is a *UsageError, and Run prints nothing for it.
func (c *app) Run(osArgs []string) error {
	if len(c.commands) < 1 {
		return ErrNoCommands
	}

	if len(osArgs) > 0 && osArgs[0] == "__complete" {
		c.handleComplete(osArgs[1:])
		return nil
	}

	// c.globalFlags is always a non-nil *GlobalFlags (set once in NewApp),
	// so GetStructFields cannot return an error here.
	globalFields, _ := structs.GetStructFields(c.globalFlags, nil, structs.DefaultEncodingTags)

	// commandArgs holds the osArgs that are commands
	// allArgs holds the osArgs that are not commands
	command, commandArgs, allArgs := c.matchCommandByArgs(osArgs, globalFields)
	if command == nil {
		// with no command matched, `app --flag` behaves like `app <default> --flag` and bare `app` runs the default.
		command = c.defaultCommand
	}
	// cmdPath is the matched command path (e.g. "db migrate"),
	// handed to the resolver so it can apply per-command rules.
	cmdPath := strings.Join(commandArgs, " ")

	var commandFields []structs.Field
	if command != nil {
		fields, err := structs.GetStructFields(command.Options(), nil, structs.DefaultEncodingTags)
		if err != nil {
			return fmt.Errorf("failed to get struct fields: %w", err)
		}
		commandFields = fields
	}

	// cmdArgs are the args defined as numeric tags in the struct e.g. `arg:"0"`
	// cmdUnknownArgs are the args that are not defined in the struct
	cmdArgs, cmdUnknownArgs, commandOptions, globalFlags, flagErr := getCommandArgs(allArgs, commandFields, globalFields)

	// --help-format spans the built-in formats plus any output codecs registered via HelpOutputs,
	// so it is validated here (against the full set) rather than by the static oneof rule on
	// GlobalFlags.HelpFormat, which only knows the built-ins.
	if err := c.validateFormat(globalFlags[argHelpFormat]); err != nil {
		return &UsageError{Command: cmdPath, Err: err}
	}

	err := mapStructToOptions(c.globalFlags, globalFlags, argHelpFormat)
	if err != nil {
		return c.usageError(cmdPath, flagValueError(c.globalFlags, globalFlags, err))
	}
	// --help-values is a help mode, so it implies --help.
	if c.globalFlags.HelpValues {
		c.globalFlags.Help = true
	}

	if c.globalFlags.Version {
		c.printVersion()
		return ErrShowingVersion
	}

	// the help command takes a command path as its positional arguments.
	if help, ok := command.(HelpCommand); ok {
		help.SetPath(cmdUnknownArgs)
		cmdUnknownArgs = nil
	}

	// a leftover positional is a mistyped command, reported even beside --help so a typo is never hidden,
	// or a stray argument, reported unless a bad flag it likely belongs to is reported instead.
	if len(cmdUnknownArgs) > 0 && ((len(commandArgs) == 0 && len(cmdArgs) == 0) || len(command.Commands()) > 0) {
		return c.usageError(cmdPath, fmt.Errorf("has no command %q: %w", cmdUnknownArgs[0], ErrCommandNotFound))
	}
	if len(cmdUnknownArgs) > 0 && flagErr == nil {
		return c.usageError(cmdPath, fmt.Errorf("does not take argument %q: %w", cmdUnknownArgs[0], ErrUnknownArg))
	}
	if flagErr != nil && !c.globalFlags.Help {
		return c.usageError(cmdPath, flagErr)
	}

	// commandOptions holds the parsed flags; fold in positions keyed by index
	// so the two together form the highest-precedence flags layer.
	flags := commandOptions
	for i, arg := range cmdArgs {
		flags[strconv.Itoa(i)] = arg
	}

	// if --help is passed, or no command matched, show help
	if c.globalFlags.Help || command == nil {
		// with --help-values, populate the matched command's struct so help can show resolved values.
		// Skip validation (the resolve-only path) so --help still works when required inputs are absent.
		if c.globalFlags.HelpValues && command != nil && flagErr == nil {
			if err := c.resolveCommandConfig(command, cmdPath, flags); err != nil {
				return err
			}
		}

		err := c.runHelp(commandArgs)
		if err != nil {
			return fmt.Errorf("failed to run help: %w", err)
		}

		return ErrShowingHelp
	}

	if err := c.loadCommandConfig(command, cmdPath, flags); err != nil {
		return err
	}

	err = command.Run(*c.globalFlags)
	if err != nil {
		if errors.Is(err, ErrDisplaySubCommands) {
			return c.runHelp(commandArgs)
		}
		return fmt.Errorf("failed to run command %q: %w", command.Name(""), err)
	}

	return nil
}

// matchCommandByArgs walks args for the command path. A positional naming a subcommand
// of the command matched so far descends into it, and the first one that does not ends the path.
// A flag's value is skipped so it is never read as a command name.
func (c *app) matchCommandByArgs(args []string, globals []structs.Field) (Command[any], []string, []string) {
	var command Command[any]
	commandNameArgs := make([]string, 0)
	allOtherArgs := make([]string, 0)
	children := c.commands

	for a := 0; a < len(args); a++ {
		if isFlag(args[a]) {
			allOtherArgs = append(allOtherArgs, args[a])
			if a+1 < len(args) && isValue(args[a+1]) && flagTakesValue(args[a], command, children, globals) {
				a++
				allOtherArgs = append(allOtherArgs, args[a])
			}
			continue
		}

		if next := c.matchCommandByName(args[a], children); next != nil {
			command, children = next, next.Commands()
			commandNameArgs = append(commandNameArgs, args[a])
			continue
		}
		children = nil
		allOtherArgs = append(allOtherArgs, args[a])
	}

	return command, commandNameArgs, allOtherArgs
}

// flagTakesValue reports whether flag consumes the next argument. It is looked up on command,
// then the global flags, then every command under children, since a flag may precede its subcommand.
func flagTakesValue(flag string, command Command[any], children []Command[any], globals []structs.Field) bool {
	if strings.Contains(flag, "=") {
		return false
	}
	var fields []structs.Field
	if command != nil {
		fields, _ = structs.GetStructFields(command.Options(), nil, structs.DefaultEncodingTags)
	}
	name := strings.TrimPrefix(strings.TrimPrefix(flag, optionPrefix), optionPrefix)
	if field := matchFlag(append(fields, globals...), name); field != nil {
		return field.Kind != reflect.Bool
	}
	for _, child := range children {
		if flagTakesValue(flag, child, child.Commands(), nil) {
			return true
		}
	}

	return false
}

// usageError wraps err, a predicate about the command path, into a UsageError that names it ("app db migrate").
func (c *app) usageError(cmdPath string, err error) error {
	title := cmp.Or(strings.TrimSpace(c.config.Name+" "+cmdPath), "app")

	return &UsageError{Command: cmdPath, Err: fmt.Errorf("%s %w", title, err)}
}

func (c *app) matchCommandByName(arg string, commands []Command[any]) Command[any] {
	var command Command[any]
	for i := range commands {
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
		return &UsageError{Command: cmd, Err: fmt.Errorf("failed to validate command %q: %w", command.Name(""), err)}
	}

	return nil
}

// env folds the process environment into commandOptions, keyed by variable name,
// so fields are matched by their `env:` tag during the merge. The framework folds
// env in after the resolver chain and before flags, so env beats files but loses to a typed flag.
func env(commandOptions map[string]any) {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			commandOptions[pair[0]] = pair[1]
		}
	}
}

// resolveCommandConfig populates command.Options() from the ordered layers without validating.
// It is the merge half of loadCommandConfig, shared with the --help-values path,
// which needs the resolved field values to display but must not fail when a required input is absent
// (the user only asked for help).
func (c *app) resolveCommandConfig(command Command[any], cmd string, flags map[string]any) error {
	inputs := command.Options()
	manager := structs.New(inputs, structs.WithTags(defaultTags...))

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
			return c.usageError(cmd, flagValueError(inputs, flags, err))
		}
	}

	return nil
}

// flagValueError names the flag behind err, the setter's error for flags on inputs, which names the struct field instead.
// The flag is the first one, in sorted order, that cannot be set alone on a zero copy of inputs.
func flagValueError(inputs any, flags map[string]any, err error) error {
	name := "flags"
	keys := make([]string, 0, len(flags))
	for key := range flags {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		probe := structs.New(reflect.New(reflect.TypeOf(inputs).Elem()).Interface(), structs.WithTags(defaultTags...))
		if probe.Set(map[string]any{key: flags[key]}) == nil {
			continue
		}
		if _, err := strconv.Atoi(key); err == nil {
			name = fmt.Sprintf("argument %q", flags[key])
		} else if len(key) == 1 {
			name = fmt.Sprintf("flag %q", optionPrefix+key)
		} else {
			name = fmt.Sprintf("flag %q", optionPrefix+optionPrefix+key)
		}
		break
	}

	return fmt.Errorf("%s got a value it cannot take, %w: %w", name, err, ErrInvalidFlagValue)
}
