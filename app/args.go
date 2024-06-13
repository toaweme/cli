package app

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

type Args struct {
	CommandsOrArgs []string
	Options        map[string]any
}

const optionPrefix = "-"
const keyValuePairSeparator = "="

func getArgs() *Args {
	args := os.Args[1:]
	if len(args) < 1 {
		log.Trace().Msg("no args provided")
		return &Args{
			CommandsOrArgs: []string{helpCommand},
			Options:        map[string]any{},
		}
	}

	parsed := &Args{
		CommandsOrArgs: make([]string, 0),
		Options:        make(map[string]any),
	}

	previousIsArgIsOption := false
	for index, arg := range args {
		// -key arg
		if strings.HasPrefix(arg, optionPrefix) {
			// -key=value arg
			if strings.Contains(arg, keyValuePairSeparator) {
				pair := strings.Split(arg, keyValuePairSeparator)

				varName := pair[0]
				varName = strings.TrimLeft(varName, "-")

				varValue := pair[1]

				addOption(parsed, varName, varValue)
				log.Trace().Str("varName", varName).Str("varValue", varValue).Msg("adding option")
			} else {
				previousIsArgIsOption = true
			}
		}

		// if the arg is an option
		if !strings.HasPrefix(arg, optionPrefix) {
			log.Trace().Str("arg", arg).Msg("arg is not an option")
			// if the previous arg was an option
			// then this arg is the value for the previous option
			if previousIsArgIsOption {
				log.Trace().Str("arg", arg).Msg("previous arg was an option")
				previousVarName := strings.TrimLeft(args[index-1], "-")
				log.Trace().Str("previousVarName", previousVarName).Msg("adding option")
				addOption(parsed, previousVarName, arg)
				previousIsArgIsOption = false
				continue
			}
			log.Trace().Str("arg", arg).Msg("adding arg")
			parsed.CommandsOrArgs = append(parsed.CommandsOrArgs, arg)
			continue
		}
	}

	return parsed
}

func addOption(parsed *Args, varName string, varValue string) {
	if existingVarValue, ok := parsed.Options[varName]; ok {
		switch val := existingVarValue.(type) {
		case []string:
			parsed.Options[varName] = append(val, varValue)
		case string:
			parsed.Options[varName] = []string{val, varValue}
		}
	} else {
		parsed.Options[varName] = varValue
	}
}
