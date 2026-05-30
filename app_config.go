package cli

import (
	"fmt"
	"os"
	"strings"
)

// loadCommandConfig populates command.Options() according to the resolved merge
// strategy, then validates the merged result. flags are the explicit CLI inputs
// (parsed flags plus positionals keyed by index); pass nil when the command takes
// none.
//
// MergeLayered (with a Store) layers defaults -> config store(s) -> env -> flags,
// so a shared section in the config file (e.g. a `database:` block) fills any
// field tagged to match it, while env and flags override per field. MergeEnvFlags
// (the default, and the fallback when MergeLayered is requested without a Store)
// applies defaults -> env -> flags only. Validation runs after the merge so
// `required` is satisfied by config- or default-provided values, not just flags.
func (c *app) loadCommandConfig(command Command[any], flags map[string]any) error {
	inputs := command.Options()
	cmdStrategy, mapping := command.ConfigStrategy()

	if c.resolveStrategy(cmdStrategy) == MergeLayered && c.store != nil {
		// default layout: shared top-level config (the plain tag match inside
		// Load) plus the command's own "<name>:" section overriding it. A command
		// that declares its own mapping opts out of the name-namespace default.
		if mapping == nil {
			if name := command.Name(""); name != "" {
				mapping = Namespaced(name)
			}
		}
		if err := c.store.Resolve(inputs, LoadOptions{Env: true, Flags: flags, Mapping: mapping}); err != nil {
			return fmt.Errorf("failed to load config for command %q: %w", command.Name(""), err)
		}
	} else if err := mergeConfig(inputs, nil, "", true, flags, nil); err != nil {
		return fmt.Errorf("failed to merge config for command %q: %w", command.Name(""), err)
	}

	// validate against the explicit inputs the user supplied; rules like
	// `required` fall back to the now-populated field values, so values sourced
	// from the config file or defaults still satisfy them.
	validateInputs := map[string]any{}
	env(validateInputs)
	for k, v := range flags {
		validateInputs[k] = v
	}
	if err := command.Validate(validateInputs); err != nil {
		return fmt.Errorf("failed to validate command %q: %w", command.Name(""), err)
	}

	return nil
}

// bindConfigTree hands the app Config to every registered command and subcommand
// that can receive it (anything embedding BaseCommand), plus the default command.
// Walking the tree at Run time avoids the ordering pitfalls of binding at
// registration, when a parent may be added before its config is known.
func (c *app) bindConfigTree() {
	var walk func(cmds []Command[any])
	walk = func(cmds []Command[any]) {
		for _, cmd := range cmds {
			if binder, ok := cmd.(configBinder); ok {
				binder.bindConfig(c.config, c.store)
			}
			walk(cmd.Commands())
		}
	}
	walk(c.commands)

	if c.defaultCommand != nil {
		if binder, ok := c.defaultCommand.(configBinder); ok {
			binder.bindConfig(c.config, c.store)
		}
	}
}

// resolveStrategy resolves the effective merge strategy from a command's declared
// strategy: it wins, falling back to the app-wide Config.Merge, and finally to
// MergeEnvFlags when neither is set (both MergeInherit).
func (c *app) resolveStrategy(cmdStrategy MergeStrategy) MergeStrategy {
	strategy := cmdStrategy
	if strategy == MergeInherit {
		strategy = c.config.Merge
	}
	if strategy == MergeInherit {
		strategy = MergeEnvFlags
	}
	return strategy
}

func env(commandOptions map[string]any) {
	environ := os.Environ()
	for _, env := range environ {
		pair := strings.SplitN(env, "=", 2)
		commandOptions[pair[0]] = pair[1]
	}
}
