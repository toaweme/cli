package cli

import (
	"fmt"

	"github.com/toaweme/structs"
)

// loadCommandConfig populates command.Options() from three ordered layers and then
// validates the result. cmd is the matched command path (e.g. "db migrate"); flags
// are the explicit CLI inputs (parsed flags plus positionals keyed by index), and
// are the highest-precedence layer.
//
// The layers, lowest first:
//  1. struct `default:` tags
//  2. the Resolver's returned map (files, env, per-command mapping - its business)
//  3. flags, applied as a separate pass so a typed flag always wins
//
// Applying the resolver map and the flags as distinct structs.Set passes is what
// makes flags beat env: within a single pass, an `env:` tag match short-circuits,
// so a merged map cannot express "flags over env". Validation runs after the merge
// so `required` is satisfied by config- or default-provided values, not just flags.
func (c *app) loadCommandConfig(command Command[any], cmd string, flags map[string]any) error {
	inputs := command.Options()
	manager := structs.New(inputs, structs.DefaultRules, structs.WithTags(defaultTags...))

	resolver := c.resolver
	if resolver == nil {
		resolver = ResolverDefault
	}

	values, err := resolver.Resolve(cmd, flags)
	if err != nil {
		return fmt.Errorf("failed to resolve config for command %q: %w", command.Name(""), err)
	}

	// defaults + resolver layer; an empty map still applies struct `default:` tags.
	if values == nil {
		values = map[string]any{}
	}
	if err := manager.Set(values); err != nil {
		return fmt.Errorf("failed to apply resolved config for command %q: %w", command.Name(""), err)
	}

	// flags win, as a separate pass.
	if len(flags) > 0 {
		if err := manager.Set(flags); err != nil {
			return fmt.Errorf("failed to apply flags for command %q: %w", command.Name(""), err)
		}
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
