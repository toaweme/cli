package cli

import (
	"fmt"
	"log/slog"

	"github.com/contentforward/structs"
)

var ErrCommandNotFound = fmt.Errorf("command not found")
var ErrNoCommands = fmt.Errorf("no commands registered")
var ErrNoArguments = fmt.Errorf("no arguments provided")

const helpCommand = "help"

type App interface {
	Commands() []Command[any]
	Add(name string, cmd Command[any]) Command[any]
	Run(osArgs []string) error
}

type CLI struct {
	settings      Settings
	globalOptions *GlobalOptions
	commands      []Command[any]
}

type Settings struct {
}

type GlobalOptions struct {
	Cwd       string `arg:"cwd" short:"c" help:"Current working directory"`
	Help      bool   `arg:"help" short:"h" help:"Show help"`
	Verbosity int    `arg:"v" help:"Verbosity level (0 - quiet, 1 - normal, 2 - verbose)"`
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
		return ErrNoArguments
	}

	globalOptions, err := c.getGlobalOptions(osArgs)
	if err != nil {
		return fmt.Errorf("failed to get command: %w", err)
	}

	err = mapStructToOptions(c.globalOptions, globalOptions)
	if err != nil {
		return fmt.Errorf("failed to update global options struct: %w", err)
	}

	// if --help is passed, show help
	if c.globalOptions.Help {
		err := c.runHelp()
		if err != nil {
			return fmt.Errorf("failed to run help: %w", err)
		}
		return nil
	}
	//
	// if len(allArgs) == 0 {
	// 	err := c.runHelp()
	// 	if err != nil {
	// 		return fmt.Errorf("failed to run help: %w", err)
	// 	}
	// 	return nil
	// }

	command, allArgs, err := c.matchCommandByArgs(osArgs)
	if err != nil {
		return fmt.Errorf("failed to match command by args: %v: %w", osArgs, err)
	}

	commandInputs := command.Options()
	commandFields, err := structs.GetStructFields(commandInputs)
	if err != nil {
		return fmt.Errorf("failed to get struct fields: %w", err)
	}
	cmdArgs, cmdUnknownArgs, commandOptions, cmdUnknownOptions := getCommandArgs(allArgs, commandFields)
	unknowns := Unknowns{
		Args:    cmdUnknownArgs,
		Options: cmdUnknownOptions,
	}

	// fill the options map with the args so that commands can use them
	for i, arg := range cmdArgs {
		commandOptions[fmt.Sprintf("%s", i)] = arg
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

func (c *CLI) runHelp() error {
	for _, cmd := range c.commands {
		if cmd.Name("") == helpCommand {
			err := cmd.Run(*c.globalOptions, Unknowns{
				Args:    []string{},
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

// Unknowns is a struct that holds unknown args and options
// it's a struct for the user to have the ability to
type Unknowns struct {
	Args    []string
	Options map[string]any
}

func (c *CLI) getGlobalOptions(osArgs []string) (map[string]any, error) {
	globalFields, err := structs.GetStructFields(c.globalOptions)
	if err != nil {
		return nil, nil
	}

	// GlobalOptions struct does not accept arguments
	// so we use unknownArgs as the processed arguments
	_, _, globalOptions, _ := getCommandArgs(osArgs, globalFields)

	return globalOptions, nil
}

func (c *CLI) matchCommandByArgs(args []string) (Command[any], []string, error) {
	var command Command[any]
	var commandArgs []int

	for a := 0; a < len(args); a++ {
		// previous arg is a command
		// assert if this arg is a sub command
		if command != nil {
			subCommand := c.matchCommandByName(args[a], command.Commands())
			if subCommand != nil {
				command = subCommand
				commandArgs = append(commandArgs, a)
				continue
			}

			// no sub command found, but
			// we have a command already so let's keep it and try further args
			break
		}

		cmd := c.matchCommandByName(args[a], c.commands)
		if cmd != nil {
			command = cmd
			commandArgs = append(commandArgs, a)
			continue
		}
	}

	slog.Info("command", "command", command, "args", commandArgs)

	if command == nil {
		return nil, nil, ErrCommandNotFound
	}

	// create a new slice that excludes the command args
	// we don't need the command args anymore
	excludedArgs := make([]string, 0)
	for i := 0; i < len(args); i++ {
		for _, ca := range commandArgs {
			if i == ca {
				continue
			}
		}
		excludedArgs = append(excludedArgs, args[i])
	}

	// argIndex+1 to exclude the matched command
	return command, excludedArgs, nil
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
