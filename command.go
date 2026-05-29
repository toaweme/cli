package cli

import (
	"errors"
	"fmt"

	"github.com/toaweme/structs"
)

// ErrValidationFailed is returned by Validate when one or more struct rules fail.
var ErrValidationFailed = errors.New("validation failed")

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
	// Description returns a longer, multi-line description shown in detailed and
	// agent help. Help stays the one-line listing summary; Description carries the
	// richer body (paragraphs, install instructions, ...). Empty by default.
	Description() string
	// Examples returns usage examples shown in detailed and agent help. Each
	// example is a slice of lines: the first is the invocation, any following
	// lines are sample output shown beneath it. Nil by default.
	Examples() [][]string
	// Args returns multi-line descriptions for positional arguments, keyed by
	// zero-based position. Augments the single-line `help:` tag. Nil by default.
	Args() map[int][]string
	// Flags returns multi-line descriptions for flags, keyed by the flag as
	// written (e.g. "--query, -q"). Augments the single-line `help:` tag. Nil by
	// default.
	Flags() map[string][]string
}

// BaseCommand provides default implementations for the Command interface.
// Embed it in your command struct to get name management, subcommand registration,
// config struct handling, validation, and no-op help providers for free. Override
// Description/Examples/Args/Flags to enrich help output.
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
	manager := structs.New(c.Inputs, structs.DefaultRules, structs.WithTags(defaultTags...))
	validationErrs, err := manager.Validate(options)
	if err != nil {
		return fmt.Errorf("failed to validate cli args structure: %w", err)
	}

	if len(validationErrs) > 0 {
		return fmt.Errorf("validation failed: %v: %w", validationErrs, ErrValidationFailed)
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

// Description returns no long-form description by default. Override to provide one.
func (c *BaseCommand[T]) Description() string { return "" }

// Examples returns no usage examples by default. Override to provide them.
func (c *BaseCommand[T]) Examples() [][]string { return nil }

// Args returns no positional-argument descriptions by default. Override to provide them.
func (c *BaseCommand[T]) Args() map[int][]string { return nil }

// Flags returns no flag descriptions by default. Override to provide them.
func (c *BaseCommand[T]) Flags() map[string][]string { return nil }
