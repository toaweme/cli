package cli

import (
	"fmt"
	"strings"

	"github.com/contentforward/structs"
)

const tagArg = "arg"
const tagShort = "short"

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
		return []string{}, []string{}, map[string]any{}, map[string]any{}
	}

	var parsedArgs = make([]string, 0)
	var unknownArgs = make([]string, 0)
	var parsedOptions = make(map[string]any)
	var unknownOptions = make(map[string]any)

	for index := 0; index < len(args); index++ {
		arg := args[index]
		// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		// logger = logger.With("i", index)

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
				nextArg := ""
				if len(args) > index+1 {
					nextArg = args[index+1]
				}
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

func splitKeyValue(arg string) (string, string) {
	pair := strings.SplitN(arg, keyValuePairSeparator, 2)

	optionName := pair[0]
	optionName = strings.TrimLeft(optionName, "-")

	optionValue := ""
	if len(pair) > 1 {
		optionValue = pair[1]
	}

	// slog.Info("key=value", "opt", optionName, "value", optionValue)

	return optionName, optionValue
}
