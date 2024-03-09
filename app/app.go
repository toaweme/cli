package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/contentforward/structs"

	"github.com/contentforward/cli"
)

var ErrNoEnvFile = fmt.Errorf("no .env file found")

const helpCommand = "help"

type CLI struct {
	cwd      string
	envFile  string
	commands map[string]cli.Command
}

func NewApp(cwd, envFile string) *CLI {
	return &CLI{
		cwd:      cwd,
		envFile:  envFile,
		commands: make(map[string]cli.Command),
	}
}

var _ cli.App = (*CLI)(nil)

func (c *CLI) Init() error {
	writer := zerolog.ConsoleWriter{Out: os.Stderr}
	logger := log.With().Caller().Logger().Output(writer)
	log.Logger = logger

	err := godotenv.Overload(c.envFile)
	if err != nil {
		return fmt.Errorf("failed to load .env file: %s: %w: %w", c.envFile, err, ErrNoEnvFile)
	}

	return nil
}

func (c *CLI) GetCommands() map[string]cli.Command {
	cmds := make(map[string]cli.Command)
	for name, cmd := range c.commands {
		cmds[name] = cmd
	}
	return cmds
}

func (c *CLI) AddCommand(name string, cmd cli.Command) error {
	c.commands[name] = cmd
	return nil
}

func (c *CLI) commandExists(name string) bool {
	_, ok := c.commands[name]
	log.Trace().Str("command", name).Bool("exists", ok).Msg("checking if command exists")
	return ok
}

func (c *CLI) Run() error {
	exec := getArgs()
	if len(exec.CommandsOrArgs) == 0 {
		return fmt.Errorf("no commands to run")
	}

	commandNameKey, args := getCommand(exec, c.commandExists)
	log.Trace().Str("command", commandNameKey).Any("args", args).Msg("found command")
	cmd, ok := c.commands[commandNameKey]
	if !ok {
		log.Debug().Str("command", commandNameKey).Msg("command not found")
		cmd, ok = c.commands["help"]
		if !ok {
			return fmt.Errorf("please register a 'help' command")
		}
		commandNameKey = "help"
	}

	log.Trace().Str("command", commandNameKey).Msg("validating cli command args and options")

	err := cmd.Validate(exec.Options)
	if err != nil {
		return fmt.Errorf("command '%s' validation failed: %w", commandNameKey, err)
	}

	log.Trace().Str("command", commandNameKey).Msg("running cli command")

	commandStructure := cmd.Structure()
	err = mapStructure(commandStructure, merge(args, exec.Options))
	if err != nil {
		return fmt.Errorf("failed to map cli command structure to args: %w", err)
	}

	err = cmd.Run(exec.Options)
	if err != nil {
		return fmt.Errorf("cli command '%s' execution failed: %w", commandNameKey, err)
	}

	return nil
}

func getCommand(givenArgs Args, existsFunc func(string) bool) (string, []string) {
	log.Trace().Any("args", givenArgs).Msg("getting command")
	var args []string
	var foundCommandName string

	// find the longest command that exists
	if len(givenArgs.CommandsOrArgs) > 1 {
		commandName := strings.Join(givenArgs.CommandsOrArgs, " ")
		if existsFunc(commandName) {
			foundCommandName = commandName
			log.Trace().Str("command", commandName).Any("args", args).Msg("found command")
		}

		argOffset := strings.Count(commandName, " ")
		if argOffset > 0 {
			args = givenArgs.CommandsOrArgs[argOffset:]
		}
	} else {
		// if there is only one command, use it
		foundCommandName = strings.Join(givenArgs.CommandsOrArgs, " ")
	}

	log.Trace().Str("command", foundCommandName).Any("args", args).Msg("longest command found")

	if !existsFunc(foundCommandName) {
		log.Trace().Str("command", foundCommandName).Any("args", args).Msg("command not found, using help command")
		foundCommandName = helpCommand
	}
	return foundCommandName, args
}

func mapStructure(structure any, vars map[string]any) error {
	manager := structs.NewManager(structure, structs.DefaultRules, structs.DefaultTags...)
	errors, err := manager.Validate(vars)
	if err != nil {
		return fmt.Errorf("error validating cli command structure: %w", err)
	}

	if len(errors) > 0 {
		for field, rules := range errors {
			log.Error().Str("field", field).Any("rules", rules).Msg("validation error")
		}
		return fmt.Errorf("validation failed: %v", errors)
	}

	err = manager.SetFields(vars)
	if err != nil {
		return fmt.Errorf("failed to set fields: %w", err)
	}

	return nil
}

const catchAllChar = "-"

func merge(args []string, options map[string]any) map[string]any {
	newVars := make(map[string]any)
	newVars[catchAllChar] = map[string]any{}
	for key := range options {
		// add options to newVars via catch-all key
		if val, ok := newVars[catchAllChar]; ok {
			// safely assert val as map[string]any with a check on the assertion result
			if valMap, ok := val.(map[string]any); ok {
				valMap[key] = options[key]
				newVars[catchAllChar] = valMap
			}
		} else {
			newVars[catchAllChar] = map[string]any{key: options[key]}
		}
	}
	for argIndex, arg := range args {
		newVars[fmt.Sprintf("%d", argIndex)] = arg
	}

	return newVars
}
