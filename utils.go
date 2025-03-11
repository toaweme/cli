package cli

import (
	"fmt"
	
	"github.com/toaweme/structs"
)

var defaultTags = structs.DefaultTags

func mapStructToOptions(structure any, vars map[string]any) error {
	manager := structs.New(structure, structs.DefaultRules, defaultTags...)
	// errors, err := manager.Validate(vars)
	// if err != nil {
	// 	return fmt.Errorf("error validating cli args structure: %w", err)
	// }
	//
	// if len(errors) > 0 {
	// 	return fmt.Errorf("validation failed: %v", errors)
	// }
	
	err := manager.Set(vars)
	if err != nil {
		return fmt.Errorf("failed to set fields: %w", err)
	}
	
	return nil
}
