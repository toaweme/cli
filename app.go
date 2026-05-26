package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/toaweme/structs"
)

var ErrCommandNotFound = fmt.Errorf("command not found")
var ErrNoCommands = fmt.Errorf("no commands registered")
var ErrNoArguments = fmt.Errorf("no arguments provided")
var ErrDisplaySubCommands = fmt.Errorf("print sub commands")
var ErrShowingHelp = fmt.Errorf("showing help")
var ErrShowingVersion = fmt.Errorf("showing version")

const helpCommand = "help"

type App interface {
	Commands() []Command[any]
	Add(name string, cmd Command[any]) Command[any]
	Run(osArgs []string) error
}

type CLI struct {
	settings       Settings
	globalOptions  *GlobalOptions
	commands       []Command[any]
	defaultCommand Command[any]
}

// Unknowns holds arguments and options that were not matched to any defined field.
// Commands receive these to support pass-through or dynamic flag handling.
type Unknowns struct {
	// Args are positional arguments not matched to numbered struct tags.
	Args []string
	// Options are key-value flags not defined in the command's config struct.
	Options map[string]any
}

// Settings configures the CLI application identity.
type Settings struct {
	// Name is the application binary name, shown in help and usage output.
	Name string
	// Version is the semantic version string shown by the version command.
	Version string
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
	Verbosity int `arg:"verbosity" env:"VERBOSITY" help:"Verbosity level (0 - quiet, 1 - normal, 2 - verbose)"`
	// Format controls help output: pretty, plain, md, json, jsonschema.
	Format string `arg:"format" help:"Help output format: pretty, plain, md, json, jsonschema"`
}

func NewApp(settings Settings, opts GlobalOptions) *CLI {
	return &CLI{
		settings:      settings,
		globalOptions: &opts,
		commands:      make([]Command[any], 0),
	}
}

var _ App = (*CLI)(nil)

func (c *CLI) Commands() []Command[any] {
	return c.commands
}

func (c *CLI) Default(cmd Command[any]) Command[any] {
	c.defaultCommand = cmd

	return cmd
}

func (c *CLI) Add(name string, cmd Command[any]) Command[any] {
	cmd.Name(name)
	c.commands = append(c.commands, cmd)

	return cmd
}

func (c *CLI) Run(osArgs []string) error {
	if len(c.commands) < 1 {
		return ErrNoCommands
	}

	if len(osArgs) > 0 && osArgs[0] == "__complete" {
		c.handleComplete(osArgs[1:])
		return nil
	}

	if len(osArgs) < 1 {
		if c.defaultCommand == nil {
			err := c.runHelp(nil)
			if err != nil {
				return fmt.Errorf("failed to run help: %w", err)
			}
			return ErrShowingHelp
		}

		commandInputs := c.defaultCommand.Options()
		commandOptions := make(map[string]any)
		// defaultCommand supports only env as arguments
		env(commandOptions)

		err := mapStructToOptions(commandInputs, commandOptions)
		if err != nil {
			return fmt.Errorf("failed to map command options: %w", err)
		}

		err = c.defaultCommand.Validate(commandOptions)
		if err != nil {
			return fmt.Errorf("failed to validate default command: %w", err)
		}

		err = c.defaultCommand.Run(*c.globalOptions, Unknowns{
			Args:    []string{},
			Options: map[string]any{},
		})
		if err != nil {
			return fmt.Errorf("failed to run default command: %w", err)
		}

		return nil
	}

	globalOptions, globalUnknownOpts, err := c.getGlobalOptions(osArgs)
	if err != nil {
		return fmt.Errorf("failed to get command: %w", err)
	}

	err = mapStructToOptions(c.globalOptions, globalOptions)
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
		if errors.Is(err, ErrCommandNotFound) && c.globalOptions.Help {
			helpErr := c.runHelp(commandArgs, globalUnknownOpts)
			if helpErr != nil {
				return fmt.Errorf("failed to run help: %w", helpErr)
			}

			return fmt.Errorf("%w: %w", err, ErrShowingHelp)
		}

		helpErr := c.runHelp(commandArgs, globalUnknownOpts)
		if helpErr != nil {
			return fmt.Errorf("failed to run help: %w", helpErr)
		}
		return fmt.Errorf("%w: %w", err, ErrShowingHelp)
	}

	commandInputs := command.Options()
	commandFields, err := structs.GetStructFields(commandInputs, nil)
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

	// fill the options map with the args so that commands can use them
	for i, arg := range cmdArgs {
		commandOptions[fmt.Sprintf("%d", i)] = arg
	}
	// add all environment variables to the options map
	env(commandOptions)

	// if --help is passed, show help
	if c.globalOptions.Help {
		err := c.runHelp(commandArgs, globalUnknownOpts)
		if err != nil {
			return fmt.Errorf("failed to run help: %w", err)
		}

		return ErrShowingHelp
	}

	err = mapStructToOptions(commandInputs, commandOptions)
	if err != nil {
		return fmt.Errorf("failed to map command options: %w", err)
	}
	err = command.Validate(commandOptions)
	if err != nil {
		return fmt.Errorf("failed to validate command: %w", err)
	}
	err = command.Run(*c.globalOptions, unknowns)
	if err != nil {
		if errors.Is(err, ErrDisplaySubCommands) {
			return c.runHelp(commandArgs, globalUnknownOpts)
		}
		return fmt.Errorf("failed to run command: %s: %w", command.Name(""), err)
	}

	return nil
}

func env(commandOptions map[string]any) {
	environ := os.Environ()
	for _, env := range environ {
		pair := strings.SplitN(env, "=", 2)
		commandOptions[pair[0]] = pair[1]
	}
}

func (c *CLI) printVersion() {
	fmt.Printf("%s %s\n", c.settings.Name, c.settings.Version)
}

func (c *CLI) runHelp(args []string, opts ...map[string]any) error {
	options := map[string]any{}
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	for _, cmd := range c.commands {
		if cmd.Name("") == helpCommand {
			err := cmd.Run(*c.globalOptions, Unknowns{
				Args:    args,
				Options: options,
			})
			if err != nil {
				return fmt.Errorf("failed to run help command: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("help command not found")
}

const (
	shellCompDirectiveNoFileComp = 4
)

func (c *CLI) handleComplete(args []string) {
	toComplete := ""
	if len(args) > 0 {
		toComplete = args[len(args)-1]
		args = args[:len(args)-1]
	}

	// walk args to find the deepest matching command
	commands := c.commands
	var matched Command[any]
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		found := c.matchCommandByName(arg, commands)
		if found == nil {
			break
		}
		matched = found
		commands = found.Commands()
	}

	if strings.HasPrefix(toComplete, "-") {
		prefix := strings.TrimLeft(toComplete, "-")
		c.completeFlagNames(matched, prefix)
	} else {
		for _, cmd := range commands {
			name := cmd.Name("")
			if strings.HasPrefix(name, toComplete) {
				fmt.Fprintf(os.Stdout, "%s\t%s\n", name, cmd.Help())
			}
		}
	}

	fmt.Fprintf(os.Stdout, ":%d\n", shellCompDirectiveNoFileComp)
}

func (c *CLI) completeFlagNames(cmd Command[any], prefix string) {
	seen := make(map[string]bool)

	if cmd != nil {
		c.completeFlagsFromOptions(cmd.Options(), prefix, seen)
	}
	c.completeFlagsFromOptions(c.globalOptions, prefix, seen)
}

func (c *CLI) completeFlagsFromOptions(options any, prefix string, seen map[string]bool) {
	if options == nil {
		return
	}

	fields, err := structs.GetStructFields(options, nil)
	if err != nil {
		return
	}

	for _, field := range fields {
		name := field.Tags["arg"]
		if name == "" {
			continue
		}
		if seen[name] {
			continue
		}
		if strings.HasPrefix(name, prefix) {
			seen[name] = true
			fmt.Fprintf(os.Stdout, "--%s\t%s\n", name, field.Tags["help"])
		}
	}
}

func (c *CLI) getGlobalOptions(osArgs []string) (map[string]any, map[string]any, error) {
	globalFields, err := structs.GetStructFields(c.globalOptions, nil)
	if err != nil {
		return nil, nil, nil
	}

	_, _, globalOptions, unknownOptions := getCommandArgs(osArgs, globalFields)

	return globalOptions, unknownOptions, nil
}

func (c *CLI) matchCommandByArgs(args []string) (Command[any], []string, []string, error) {
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

func (c *CLI) matchCommandByName(arg string, commands []Command[any]) Command[any] {
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
