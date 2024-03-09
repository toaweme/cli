package app

import (
	"os"
	"strings"
)

type Args struct {
	CommandsOrArgs []string
	Options        map[string]any
}

func getArgs() Args {
	args := os.Args[1:]
	if len(args) < 1 {
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
		}
	}

	return exec
}
