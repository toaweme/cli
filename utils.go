package cli

import (
	"fmt"
	"slices"

	"github.com/toaweme/structs"
)

var defaultTags = structs.DefaultCLITags

// mapStructToOptions validates vars against structure's rules, then sets them on the struct.
// Keys named in skipValidate are exempt from validation but still set: used for --help-format,
// whose valid set is extended at runtime by registered output codecs the static oneof rule cannot know about.
func mapStructToOptions(structure any, vars map[string]any, skipValidate ...string) error {
	manager := structs.New(structure, structs.WithTags(defaultTags...))

	validateVars := vars
	if len(skipValidate) > 0 {
		validateVars = make(map[string]any, len(vars))
		for k, v := range vars {
			if !slices.Contains(skipValidate, k) {
				validateVars[k] = v
			}
		}
	}

	errors, err := manager.Validate(validateVars)
	if err != nil {
		return fmt.Errorf("error validating cli args structure: %w", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	err = manager.Set(vars)
	if err != nil {
		return fmt.Errorf("failed to set fields: %w", err)
	}

	return nil
}
