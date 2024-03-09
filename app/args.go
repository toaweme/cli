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

func getArgs() Args {
	args := os.Args[1:]
	if len(args) < 1 {
		log.Trace().Msg("no args provided")
		return Args{
			CommandsOrArgs: []string{helpCommand},
			Options:        map[string]any{},
		}
	}

	exec := Args{
		CommandsOrArgs: make([]string, 0),
		Options:        make(map[string]any),
	}
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			log.Trace().Str("arg", arg).Msg("adding arg")
			exec.CommandsOrArgs = append(exec.CommandsOrArgs, arg)
			continue
		}

		if strings.Contains(arg, "=") {
			pair := strings.Split(arg, "=")

			varName := pair[0]
			varName = strings.TrimLeft(varName, "-")

			varValue := pair[1]

			if existingVarValue, ok := exec.Options[varName]; ok {
				switch val := existingVarValue.(type) {
				case []string:
					exec.Options[varName] = append(val, varValue)
				case string:
					exec.Options[varName] = []string{val, varValue}
				}
			} else {
				exec.Options[varName] = varValue
			}
			log.Trace().Str("varName", varName).Str("varValue", varValue).Msg("adding option")
		}
	}

	return exec
}
