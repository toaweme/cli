package cli

import (
	"fmt"

	"github.com/toaweme/structs"
)

// loadCommandConfig populates command.Options() from ordered layers and then
// validates the result. cmd is the matched command path (e.g. "db migrate"); flags
// are the explicit CLI inputs (parsed flags plus positionals keyed by index), and
// are the highest-precedence layer.
//
// The layers, lowest first:
//  1. struct `default:` tags
//  2. the Resolver chain, each overlaying its layer on the previous (files, mapping)
//  3. env, folded in after the chain so it beats files
//  4. flags, applied as a separate pass so a typed flag always wins
//
// Applying the merged map and the flags as distinct structs.Set passes is what
// makes flags beat env: within a single pass, an `env:` tag match short-circuits,
// so a merged map cannot express "flags over env". Validation runs after the merge
// so `required` is satisfied by config- or default-provided values, not just flags.
func (c *app) loadCommandConfig(command Command[any], cmd string, flags map[string]any) error {
	if err := c.resolveCommandConfig(command, cmd, flags); err != nil {
		return err
	}

	// validate against the explicit inputs the user supplied; rules like `required`
	// fall back to the now-populated field values, so values sourced from config or
	// defaults still satisfy them.
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

// resolveCommandConfig populates command.Options() from the ordered layers without
// validating. It is the merge half of loadCommandConfig, shared with the --help-values
// path, which needs the resolved field values to display but must not fail when a
// required input is absent (the user only asked for help).
func (c *app) resolveCommandConfig(command Command[any], cmd string, flags map[string]any) error {
	inputs := command.Options()
	manager := structs.New(inputs, structs.DefaultRules, structs.WithTags(defaultTags...))

	// run the resolver chain, threading each one's output into the next.
	values := map[string]any{}
	for _, resolver := range c.resolvers {
		next, err := resolver.Resolve(cmd, values)
		if err != nil {
			return fmt.Errorf("failed to resolve config for command %q: %w", command.Name(""), err)
		}
		if next != nil {
			values = next
		}
	}

	// env beats the resolver layers; flags (applied below) still win over env.
	env(values)

	// defaults + resolved layer; an empty map still applies struct `default:` tags.
	if err := manager.Set(values); err != nil {
		return fmt.Errorf("failed to apply resolved config for command %q: %w", command.Name(""), err)
	}

	// flags win, as a separate pass.
	if len(flags) > 0 {
		if err := manager.Set(flags); err != nil {
			return fmt.Errorf("failed to apply flags for command %q: %w", command.Name(""), err)
		}
	}

	return nil
}
