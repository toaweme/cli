package cli

import (
	"fmt"

	"github.com/toaweme/structs"
)

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
	Run(options GlobalOptions, unknowns Unknowns) error
	// Validate checks the parsed options map against struct validation rules.
	Validate(options map[string]any) error
	// Help returns a short one-line description shown in command listings.
	Help() string
}

// ExampleProvider is an optional interface commands can implement to provide
// usage examples shown in agent and detailed help output.
type ExampleProvider interface {
	Examples() []string
}

// BaseCommand provides default implementations for the Command interface.
// Embed it in your command struct to get name management, subcommand registration,
// config struct handling, and validation for free.
type BaseCommand[T any] struct {
	command  string
	commands []Command[any]
	Inputs   *T
}

func NewBaseCommand[T any]() BaseCommand[T] {
	return BaseCommand[T]{
		commands: make([]Command[any], 0),
	}
}

func (c *BaseCommand[T]) Name(name string) string {
	// getter
	if name == "" {
		return c.command
	}

	// setter
	c.command = name
	return name
}

func (c *BaseCommand[T]) Add(name string, cmd Command[any]) {
	cmd.Name(name)
	c.commands = append(c.commands, cmd)
}

func (c *BaseCommand[T]) Validate(options map[string]any) error {
	manager := structs.New(c.Inputs, structs.DefaultRules, defaultTags...)
	errors, err := manager.Validate(options)
	if err != nil {
		return fmt.Errorf("error validating cli args structure: %w", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	return nil
}

func (c *BaseCommand[T]) Options() any {
	if c.Inputs == nil {
		c.Inputs = new(T)
	}

	return c.Inputs
}

func (c *BaseCommand[T]) Commands() []Command[any] {
	return c.commands
}
