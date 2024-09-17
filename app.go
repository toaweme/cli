package cli

import (
	"fmt"
	"strings"

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

func joinOptionMap(options map[string]any) string {
	var opts []string
	for k, v := range options {
		opts = append(opts, fmt.Sprintf("%s=%v", k, v))
	}

	return strings.Join(opts, " ")
}

func (c *CLI) Run(osArgs []string) error {
	if len(c.commands) < 1 {
		return ErrNoCommands
	}
	if len(osArgs) < 1 {
		return ErrNoArguments
	}

	allArgs, globalOptions, err := c.getGlobalOptions(osArgs)
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

	if len(allArgs) == 0 {
		err := c.runHelp()
		if err != nil {
			return fmt.Errorf("failed to run help: %w", err)
		}
		return nil
	}

	command, commandArgs, err := c.matchCommandByArgs(allArgs)
	if err != nil {
		return fmt.Errorf("failed to match command by args: %v: %w", allArgs, err)
	}

	_ = commandArgs

	// opts := mergeArgs(commandArgs, globalOptions)
	// commandOptions := cmd.Options()
	// err = mapStructToOptions(commandOptions, opts)

	err = command.Run(*c.globalOptions)
	if err != nil {
		return fmt.Errorf("failed to run command: %s: %w", command.Name(""), err)
	}

	return nil

	// return command, commandArgs, nil
	// commandArgs, options,
	// 	appArgs := getCommandArgs()

	// get the commandArgs name and commandArgs by passing an existsFunc and finding the longest commandArgs that exists
	// commandNameKey, commandArgs := findExistingCommand(*appArgs, func(name string) bool {
	// 	_, ok := c.commands[name]
	// 	return ok
	// })
	//
	// opts := mergeArgs(commandArgs, appArgs.Options, false)
	// // spew.Dump("opts", opts)
	// globalOpts, err := fillBooleans(&c.globalOptions, opts, c.settings.EmptyFlagsTrue)
	// if err != nil {
	// 	return fmt.Errorf("failed to fill booleans: %w", err)
	// }
	//
	// err = mapStructToOptions(&c.globalOptions, globalOpts)
	// if err != nil {
	// 	return fmt.Errorf("failed to map global options: %w", err)
	// }
	//
	// var cmd Command[any]
	// cmd, ok := c.commands[commandNameKey]
	// if !ok {
	// 	cmd, ok = c.commands[helpCommand]
	// 	if !ok {
	// 		return fmt.Errorf("please register a 'help' commandArgs")
	// 	}
	// }
	//
	// // have cmd here
	//
	// if !c.globalOptions.Help {
	// 	err = cmd.Validate(appArgs.Options)
	// 	if err != nil {
	// 		return fmt.Errorf("commandArgs '%s' validation failed: %w", commandNameKey, err)
	// 	}
	// }
	//
	// commandOptions := cmd.Options()
	// err = mapStructToOptions(commandOptions, opts)
	// if err != nil {
	// 	return fmt.Errorf("failed to map cli commandArgs structure to commandArgs: %w", err)
	// }
	//
	// if c.globalOptions.Help {
	// 	err := PrintOptions(commandOptions)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to print commandArgs options: %w", err)
	// 	}
	// }
	// err = cmd.Run(c.globalOptions)
	// if err != nil {
	// 	return fmt.Errorf("cli commandArgs '%s' execution failed: %w", commandNameKey, err)
	// }

	return nil
}

func (c *CLI) runHelp() error {
	for _, cmd := range c.commands {
		if cmd.Name("") == helpCommand {
			err := cmd.Run(*c.globalOptions)
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

func (c *CLI) getGlobalOptions(osArgs []string) ([]string, map[string]any, error) {
	globalFields, err := structs.GetStructFields(c.globalOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get struct fields: %w", err)
	}

	// GlobalOptions struct does not accept arguments
	// so we use unknownArgs as the processed arguments
	_, unknownArgs, options, _ := getCommandArgs(osArgs, globalFields)

	return unknownArgs, options, nil
}

func (c *CLI) matchCommandByArgs(args []string) (Command[any], []string, error) {
	var command Command[any]
	var argIndex int = -1

	for a := 0; a < len(args); a++ {
		// previous arg is a command
		// assert if this arg is a sub command
		if command != nil {
			subCommand := c.matchCommandByName(args[a], command.Commands())
			if subCommand != nil {
				command = subCommand
				argIndex = a
				continue
			}

			// no sub command found, but
			// we have a command already so let's keep it and try further args
			break
		}

		cmd := c.matchCommandByName(args[a], c.commands)
		if cmd != nil {
			command = cmd
			argIndex = a
			continue
		}
	}

	if command == nil || argIndex == -1 {
		return nil, nil, ErrCommandNotFound
	}

	return command, args[argIndex:], nil
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
