package cli

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/contentforward/structs"
)

const optionPrefix = "-"
const keyValuePairSeparator = "="
const catchAllChar = "-"

func findField(fields []structs.Field, tag string, name string) *structs.Field {
	for _, field := range fields {
		if field.Tags[tag] != "" && field.Tags[tag] == name {
			return &field
		}
	}

	return nil
}

// getCommandArgs parses the args line arguments and returns the parsed args and options and all unknowns
func getCommandArgs(args []string, fields []structs.Field) ([]string, []string, map[string]any, map[string]any) {
	if len(args) < 1 {
		return []string{helpCommand}, []string{}, map[string]any{}, map[string]any{}
	}

	var parsedArgs = make([]string, 0)
	var unknownArgs = make([]string, 0)
	var parsedOptions = make(map[string]any)
	var unknownOptions = make(map[string]any)

	for index := 0; index < len(args); index++ {
		arg := args[index]
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		logger = logger.With("i", index)

		// "-key arg" or "-key=value" syntax
		if strings.HasPrefix(arg, optionPrefix) {
			dePrefixedArg := strings.TrimLeft(arg, optionPrefix)
			var foundField *structs.Field
			for _, field := range fields {
				if field.Tags[tagArg] == dePrefixedArg {
					foundField = &field
					break
				} else if field.Tags[tagShort] == dePrefixedArg {
					foundField = &field
					break
				}
			}

			// TODO: handle key=value syntax
			// if strings.Contains(arg, keyValuePairSeparator) {
			// 	optionName, optionValue := splitKeyValue(arg)
			// 	addOption(parsedOptions, optionName, optionValue)
			// }
			if foundField != nil {
				// empty bool options are set to true
				// logger.Info("field", "opt", dePrefixedArg, "type", foundField.Type)
				if foundField.Type == "bool" {
					parsedOptions[dePrefixedArg] = true
					continue
				}

				// next arg is the value for the current option
				nextArg := args[index+1]
				parsedOptions[dePrefixedArg] = nextArg

				// skip the next arg
				index++
				continue
			} else {
				if len(args) > index+1 {
					nextArg := args[index+1]
					unknownOptions[dePrefixedArg] = nextArg
					index++
					continue
				}
				unknownOptions[dePrefixedArg] = true
			}
		}

		// arg is not an option, it's a command or an argument
		if !strings.HasPrefix(arg, optionPrefix) {
			var foundField *structs.Field
			for _, field := range fields {
				if field.Tags[tagArg] == fmt.Sprintf("%d", index) {
					foundField = &field
					break
				} else if field.Tags[tagShort] == fmt.Sprintf("%d", index) {
					foundField = &field
					break
				}
			}

			if foundField != nil {
				parsedArgs[index] = arg
			} else {
				unknownArgs = append(unknownArgs, arg)
			}
		}
	}

	return parsedArgs, unknownArgs, parsedOptions, unknownOptions
}

// getCommandArgs parses the args line arguments and returns the parsed commands and options
func getCommandArgs2(args []string, fields []structs.Field) ([]string, map[string]any) {
	if len(args) < 1 {
		return []string{helpCommand}, map[string]any{}
	}

	var parsedArgs = make([]string, 0)
	var parsedOptions = make(map[string]any)
	var previousArgIsOption = false

	for index, arg := range args {
		// "-key arg" or "-key=value" syntax
		if strings.HasPrefix(arg, optionPrefix) {
			// "-key=value" syntax
			if strings.Contains(arg, keyValuePairSeparator) {
				optionName, optionValue := splitKeyValue(arg)
				addOption(parsedOptions, optionName, optionValue)
			} else {
				// "-key arg" syntax

				// current arg is an option so we set the previousArgIsOption to true
				// to indicate that the next arg is the value for the current option
				// that's handled in the next iteration !hasPrefix block
				previousArgIsOption = true
				slog.Info("is opt", "opt", arg)

				// if the last arg is an option
				// then this arg is the value for the previous option
				if index == len(args)-1 {
					previousArgName := strings.TrimLeft(arg, "-")
					addOption(parsedOptions, previousArgName, "")
					previousArgIsOption = false
					continue
				}
			}
		}

		// if the arg is not an option
		if !strings.HasPrefix(arg, optionPrefix) {
			// assert previous arg (which is an option) for a field
			field, isFieldArg := getPreviousField(args, fields, index)

			// if the previous arg was an option
			// then this arg is the value for the previous option
			if previousArgIsOption {
				slog.Info("prev is opt", "arg", arg, "is-opt", isFieldArg)

				if isFieldArg && field != nil {
					if field.Type == "bool" {
						arg = "yes"
					}
				}
				previousArgName := strings.TrimLeft(args[index-1], "-")
				addOption(parsedOptions, previousArgName, arg)
				previousArgIsOption = false
				continue
			}

			slog.Info("is arg", "arg", arg)
			parsedArgs = append(parsedArgs, arg)
			continue
		}
	}

	return parsedArgs, parsedOptions
}

func splitKeyValue(arg string) (string, string) {
	pair := strings.SplitN(arg, keyValuePairSeparator, 2)

	optionName := pair[0]
	optionName = strings.TrimLeft(optionName, "-")

	optionValue := ""
	if len(pair) > 1 {
		optionValue = pair[1]
	}

	slog.Info("key=value", "opt", optionName, "value", optionValue)

	return optionName, optionValue
}

func getPreviousField(args []string, fields []structs.Field, index int) (*structs.Field, bool) {
	if index == 0 {
		return nil, false
	}
	field := findField(fields, tagArg, args[index-1])
	isFieldArg := field != nil
	if field == nil {
		field = findField(fields, tagShort, args[index-1])
		isFieldArg = field != nil
	}
	return field, isFieldArg
}

// addOption adds an option to the parsed struct
func addOption(options map[string]any, varName string, varValue string) {
	varName = strings.ToLower(varName)
	if existingVarValue, ok := options[varName]; ok {
		switch val := existingVarValue.(type) {
		case []string:
			options[varName] = append(val, varValue)
		case string:
			options[varName] = []string{val, varValue}
		}
	} else {
		options[varName] = varValue
	}
}

const tagArg = "arg"
const tagShort = "short"

//
// // mergeArgs merges the args and options into a single map
// // args become 0, 1, 2, etc.
// // options are added to a catch-all key
// func mergeArgs(args []string, options map[string]any) map[string]any {
// 	newOptions := make(map[string]any)
// 	for argIndex, arg := range args {
// 		newOptions[fmt.Sprintf("%d", argIndex)] = arg
// 	}
// 	for key := range options {
// 		newOptions[key] = options[key]
// 	}
//
// 	return newOptions
// }
//
// // mergeArgs merges the args and options into a single map
// // args become 0, 1, 2, etc.
// // options are added to a catch-all key
// func catchAllOptions(args []string, options map[string]any, catchAll bool) map[string]any {
// 	newOptions := make(map[string]any)
// 	for argIndex, arg := range args {
// 		newOptions[fmt.Sprintf("%d", argIndex)] = arg
// 	}
// 	newOptions[catchAllChar] = make(map[string]any)
// 	for key := range options {
// 		if !catchAll {
// 			newOptions[key] = options[key]
// 			continue
// 		}
// 		if valMap, ok := newOptions[catchAllChar].(map[string]any); ok {
// 			valMap[key] = options[key]
// 			newOptions[catchAllChar] = valMap
// 		}
// 	}
//
// 	return newOptions
// }
