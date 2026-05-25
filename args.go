package cli

import (
	"fmt"
	"strings"

	"github.com/toaweme/structs"
)

const tagArg = "arg"
const tagShort = "short"

const optionPrefix = "-"

func matchField(fields []structs.Field, name string) *structs.Field {
	for _, field := range fields {
		if field.Tags[tagArg] == name || field.Tags[tagShort] == name {
			return &field
		}
	}

	return nil
}

// getCommandArgs parses the args line arguments and returns the parsed args and options and all unknowns
func getCommandArgs(args []string, fields []structs.Field) ([]string, []string, map[string]any, map[string]any) {
	if len(args) < 1 {
		return []string{}, []string{}, map[string]any{}, map[string]any{}
	}

	parsedArgs := make([]string, 0)
	unknownArgs := make([]string, 0)
	parsedOptions := make(map[string]any)
	unknownOptions := make(map[string]any)

	for index := 0; index < len(args); index++ {
		arg := args[index]

		if !strings.HasPrefix(arg, optionPrefix) {
			foundField := matchField(fields, fmt.Sprintf("%d", index))
			if foundField != nil {
				parsedArgs = append(parsedArgs, arg)
			} else {
				unknownArgs = append(unknownArgs, arg)
			}
			continue
		}

		dePrefixedArg := strings.TrimLeft(arg, optionPrefix)

		// handle --key=value syntax
		if strings.Contains(dePrefixedArg, "=") {
			optName, optValue := splitKeyValue(dePrefixedArg)
			foundField := matchField(fields, optName)
			if foundField != nil {
				parsedOptions[optName] = optValue
			} else {
				unknownOptions[optName] = optValue
			}
			continue
		}

		foundField := matchField(fields, dePrefixedArg)
		if foundField != nil {
			if foundField.Type == "bool" {
				parsedOptions[dePrefixedArg] = true
				continue
			}

			nextArg := ""
			if len(args) > index+1 {
				nextArg = args[index+1]
			}
			parsedOptions[dePrefixedArg] = nextArg
			index++
			continue
		}

		if len(args) > index+1 {
			unknownOptions[dePrefixedArg] = args[index+1]
			index++
			continue
		}
		unknownOptions[dePrefixedArg] = true
	}

	return parsedArgs, unknownArgs, parsedOptions, unknownOptions
}

func splitKeyValue(arg string) (string, string) {
	pair := strings.SplitN(arg, "=", 2)

	optionName := pair[0]
	optionValue := ""
	if len(pair) > 1 {
		optionValue = pair[1]
	}

	return optionName, optionValue
}
