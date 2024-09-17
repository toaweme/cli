package cli

import (
	"fmt"

	"github.com/contentforward/structs"
)

type Command[T any] interface {
	// BaseCommand[T]

	Name(name string) string
	Add(name string, cmd Command[any])
	Options() any
	Commands() []Command[any]

	// Command[T]

	Run(options GlobalOptions) error
	Validate(options map[string]any) error
	Help() string
}

type BaseCommand[T any] struct {
	command  string
	commands []Command[any]
	Inputs   *T
}

func NewBaseCommand[T any]() BaseCommand[T] {
	return BaseCommand[T]{
		commands: make([]Command[any], 0),
		Inputs:   new(T),
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
	manager := structs.NewManager(c.Inputs, structs.DefaultRules, defaultTags...)
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
