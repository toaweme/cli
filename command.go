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
// config struct handling, validation, and no-op help providers for free. Override
// Description/Examples/Args/Flags to enrich help output.
type BaseCommand[T any] struct {
	command  string
	commands []Command[any]
	Inputs   *T
	config   Config
}

// configBinder is the unexported capability the app uses to hand each command the
// application Config when it is registered. BaseCommand satisfies it, so every
// command that embeds BaseCommand can read the global config without any wiring;
// the app binds the whole command tree in Run. It is unexported so only the
// framework can bind, never callers.
type configBinder interface {
	bindConfig(cfg Config)
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

// Config returns the application Config bound when the command was registered
// (name, version, the storage, the merge default). It lets a command read global
// configuration beyond the fields merged into its own Options(). Zero value until
// the command is registered with an app.
func (c *BaseCommand[T]) Config() Config { return c.config }

// Store returns the application config storage, or nil when none was configured.
// Shorthand for Config().Store, the common case for reading or persisting config.
func (c *BaseCommand[T]) Store() Storage { return c.config.Store }

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

// Options returns the pointer the parser fills (and the merge populates). It
// allocates a fresh T only when Inputs is unset, so an app can make a command
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

// ConfigStrategy defers to the app-wide Config.Merge (MergeInherit) and declares
// no field mappings by default. Override to force a specific MergeStrategy and/or
// to remap fields onto the global config (see ConfigMapping).
func (c *BaseCommand[T]) ConfigStrategy() (MergeStrategy, ConfigMapping) {
	return MergeInherit, nil
}

func (c *BaseCommand[T]) bindConfig(cfg Config) { c.config = cfg }
