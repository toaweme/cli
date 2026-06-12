package cli

import (
	"errors"
	"fmt"

	"github.com/toaweme/structs"
)

// ErrValidationFailed is returned by Validate when one or more struct rules fail.
var ErrValidationFailed = errors.New("validation failed")

// BaseCommand provides default implementations for the Command interface.
// Embed it in your command struct to get name management, subcommand registration,
// config struct handling, validation, and no-op help providers for free.
// Override Description/Examples/Args/Flags to enrich help output.
type BaseCommand[T any] struct {
	command  string
	commands []Command[any]
	Inputs   *T
}

// NewBaseCommand returns a BaseCommand with an initialized subcommand slice.
func NewBaseCommand[T any]() BaseCommand[T] {
	return BaseCommand[T]{
		commands: make([]Command[any], 0),
	}
}

// Name returns the command's name when called with an empty string, or sets and
// returns it otherwise.
func (c *BaseCommand[T]) Name(name string) string {
	// getter
	if name == "" {
		return c.command
	}

	// setter
	c.command = name
	return name
}

// Add registers cmd as a subcommand under the given name.
func (c *BaseCommand[T]) Add(name string, cmd Command[any]) {
	cmd.Name(name)
	c.commands = append(c.commands, cmd)
}

// Validate checks the parsed options against the struct rules on the command's
// config type, returning ErrValidationFailed when any rule fails.
func (c *BaseCommand[T]) Validate(options map[string]any) error {
	manager := structs.New(c.Inputs, structs.WithTags(defaultTags...))
	validationErrs, err := manager.Validate(options)
	if err != nil {
		return fmt.Errorf("failed to validate cli args structure: %w", err)
	}

	if len(validationErrs) > 0 {
		return fmt.Errorf("validation failed: %v: %w", validationErrs, ErrValidationFailed)
	}

	return nil
}

// Options returns the pointer the parser fills (and the merge populates).
// It allocates a fresh T only when Inputs is unset, so an app can make a command
// operate on a slice of a larger config struct by assigning Inputs before Run:
//
//	cmd.Inputs = &appCfg.Server // flags, env, and merge now write into appCfg
//
// That is the top-down "single source of truth" pattern: one app config struct,
// with each command viewing the field it owns. Leaving Inputs nil keeps the
// command's config independent and portable, which is the default.
func (c *BaseCommand[T]) Options() any {
	if c.Inputs == nil {
		c.Inputs = new(T)
	}

	return c.Inputs
}

// Commands returns the registered subcommands.
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
