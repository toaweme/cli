package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/awee-ai/structs"
)

var ErrCommandNotFound = fmt.Errorf("command not found")
var ErrNoCommands = fmt.Errorf("no commands registered")
var ErrNoArguments = fmt.Errorf("no arguments provided")
var ErrDisplaySubCommands = fmt.Errorf("print sub commands")
var ErrShowingHelp = fmt.Errorf("showing help")

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

// Unknowns is a struct that holds unknown args and options
// it's a struct for the user to have the ability to
type Unknowns struct {
	Args    []string
	Options map[string]any
}

type Settings struct {
}

type GlobalOptions struct {
	Cwd       string `arg:"cwd" short:"c" help:"Current working directory"`
	Help      bool   `arg:"help" short:"h" help:"Show help"`
	Verbosity int    `arg:"verbosity" short:"v" help:"Verbosity level (0 - quiet, 1 - normal, 2 - verbose)"`
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
		// default command supports only env as arguments
		env(commandOptions)

		err := mapStructToOptions(commandInputs, commandOptions)
		if err != nil {
			return fmt.Errorf("failed to map command options: %w", err)
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

	globalOptions, err := c.getGlobalOptions(osArgs)
	if err != nil {
		return fmt.Errorf("failed to get command: %w", err)
	}

	err = mapStructToOptions(c.globalOptions, globalOptions)
	if err != nil {
		return fmt.Errorf("failed to update global options struct: %w", err)
	}

	// commandArgs holds the osArgs that are commands
	// allArgs holds the osArgs that are not commands
	command, commandArgs, allArgs, err := c.matchCommandByArgs(osArgs)
	if err != nil {
		if errors.Is(err, ErrCommandNotFound) && c.globalOptions.Help {
			helpErr := c.runHelp(commandArgs)
			if helpErr != nil {
				return fmt.Errorf("failed to run help: %w", helpErr)
			}

			return fmt.Errorf("%w: %w", err, ErrShowingHelp)
		}

		slog.Error("failed to match command by args", "error", err)
		helpErr := c.runHelp(commandArgs)
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
		err := c.runHelp(commandArgs)
		if err != nil {
			return fmt.Errorf("failed to run help: %w", err)
		}

		return ErrShowingHelp
	}

	err = mapStructToOptions(commandInputs, commandOptions)
	if err != nil {
		return fmt.Errorf("failed to map command options: %w", err)
	}
	err = command.Run(*c.globalOptions, unknowns)
	if err != nil {
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

func (c *CLI) runHelp(args []string) error {
	// slog.Info("running help", "args", args)
	for _, cmd := range c.commands {
		if cmd.Name("") == helpCommand {
			err := cmd.Run(*c.globalOptions, Unknowns{
				Args:    args,
				Options: map[string]any{},
			})
			if err != nil {
				return fmt.Errorf("failed to run help command: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("help command not found")
}

func (c *CLI) getGlobalOptions(osArgs []string) (map[string]any, error) {
	globalFields, err := structs.GetStructFields(c.globalOptions, nil)
	if err != nil {
		return nil, nil
	}

	// GlobalOptions struct does not accept arguments
	// so we use unknownArgs as the processed arguments
	_, _, globalOptions, _ := getCommandArgs(osArgs, globalFields)

	return globalOptions, nil
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
