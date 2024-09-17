package cli

import (
	"fmt"
	"strings"

	"github.com/contentforward/structs"
)

func PrintCommands(commands []Command[any]) {
	longestName := 0

	for _, cmd := range commands {
		name := cmd.Name("")
		if len(name) > longestName {
			longestName = len(name)
		}
	}

	for _, cmd := range commands {
		fmt.Printf("  %s%s  : %v\n", cmd.Name(""), indent(cmd.Name(""), longestName), cmd.Help())
	}
}

func PrintOptions(structure any) error {
	fields, err := structs.GetStructFields(structure)
	if err != nil {
		return fmt.Errorf("error getting global option fields: %w", err)
	}

	longestArg := 0

	for _, field := range fields {
		argLen := len(field.Tags["arg"]) + len(field.Tags["short"])
		if argLen > longestArg {
			longestArg = argLen
		}
	}

	for _, field := range fields {
		arg := field.Tags["arg"]
		short := field.Tags["short"]
		line := fmt.Sprintf("  -%s, --%s%s: %s", short, arg, indent(arg, longestArg), field.Tags["help"])
		if short == "" {
			line = fmt.Sprintf("  -%s%s     : %s", arg, indent(arg, longestArg), field.Tags["help"])
		}
		fmt.Println(line)
	}
	return nil
}

func indent(text string, indent int) string {
	l := len(text)

	indentStr := strings.Repeat(" ", indent-l)

	return indentStr
}
