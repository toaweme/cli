package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/contentforward/structs"
)

// find command or any level of subcommand
func findCommandByArgs(commands []Command[any], args []string) Command[any] {
	if len(args) == 0 {
		return nil
	}

	for _, cmd := range commands {
		if cmd.Name("") == args[0] {
			if len(args) == 1 {
				return cmd
			}
			return findCommandByArgs(cmd.Commands(), args[1:])
		}
	}

	return nil
}

func displaySingleCommandHelp(appName string, commands []Command[any], command []string) []string {
	help := []string{
		fmt.Sprintf(`Usage: ` + appName + ` <command> <subcommand> [args] [options]`),
	}
	help = append(help, ``)

	cmd := findCommandByArgs(commands, command)
	if cmd == nil {
		fmt.Println("Command not found")
		return []string{}
	}

	help = append(help, cmd.Help())
	line := fmt.Sprintf(`$ %s`, strings.Join(command, " "))
	help = append(help, line)

	options, _ := helpOptions(cmd.Options())
	if len(options) > 0 {
		help = append(help, options...)
	}

	return help
}

func displayAllCommandsHelp(appName string, commands []Command[any]) []string {
	help := []string{
		fmt.Sprintf(`Usage: ` + appName + ` <command> <subcommand> [args] [options]`),
	}
	help = append(help, ``)
	help = append(help, `Options can be passed before or after the command and subcommand.`)
	help = append(help, `Both -[opt] <arg> and --[opt]=<arg> are supported.`)
	help = append(help, `Boolean flags can be passed without an argument to set them to true.`)
	help = append(help, ``)
	help = append(help, `Commands:`)

	longestName := getLongestName(commands)

	for _, cmd := range commands {
		name := cmd.Name("")
		if len(cmd.Commands()) > 0 {
			// help = append(help, ``)
		}

		help = append(help, fmt.Sprintf(`  %s  %s%s`, name, paddingRight(name, longestName), cmd.Help()))

		if len(cmd.Commands()) > 0 {
			for _, subCmd := range cmd.Commands() {
				subName := name + " " + subCmd.Name("")
				help = append(help, `  `+subName+``+paddingRight(subName, longestName)+`  `+subCmd.Help())
			}
		}
	}

	return help
}

func DisplayHelp(appName string, commands []Command[any], command []string) {
	help := []string{`Usage: ` + appName + ` <command> <subcommand> [args] [options]`}
	help = append(help, ``)

	if len(command) == 0 {
		help = displayAllCommandsHelp(appName, commands)
	} else {
		help = displaySingleCommandHelp(appName, commands, command)
	}

	help = append(help, ``)
	help = append(help, `Global Options:`)

	opts, err := helpOptions(&GlobalOptions{})
	if err != nil {
		fmt.Printf("Error printing global options: %v", err)
	}
	help = append(help, opts...)

	fmt.Println(strings.Join(help, "\n"))
}

func indentBy(text string, indent int) string {
	indentStr := strings.Repeat(" ", indent)

	return indentStr + text
}

func getLongestName(commands []Command[any]) int {
	longestName := 0

	for _, cmd := range commands {
		name := cmd.Name("")
		if len(name) > longestName {
			longestName = len(name)
		}
		if len(cmd.Commands()) > 0 {
			for _, subCmd := range cmd.Commands() {
				subName := name + " " + subCmd.Name("")
				if len(subName) > longestName {
					longestName = len(subName)
				}
			}
		}
	}

	return longestName
}

func helpOptions(structure any) ([]string, error) {
	fields, err := structs.GetStructFields(structure)
	if err != nil {
		return nil, fmt.Errorf("error getting global option fields: %w", err)
	}

	longestArg := 0

	for _, field := range fields {
		argLen := len(field.Tags["arg"]) + len(field.Tags["short"])
		if argLen > longestArg {
			longestArg = argLen
		}
	}

	options := []string{}

	for _, field := range fields {
		arg := field.Tags["arg"]
		short := field.Tags["short"]
		line := fmt.Sprintf("  -%s, --%s%s   %s", short, arg, paddingRight(arg, longestArg), field.Tags["help"])
		if num, ok := isNumeric(arg); ok {
			line = fmt.Sprintf("  arg%d      %s%s", num+1, paddingRight(arg, longestArg), field.Tags["help"])
		} else {
			if short == "" {
				line = fmt.Sprintf("  -%s%s        %s", arg, paddingRight(arg, longestArg), field.Tags["help"])
			}
		}
		options = append(options, line)
	}
	return options, nil
}

func isNumeric(str string) (int64, bool) {
	i, err := strconv.ParseInt(str, 10, 64)
	return i, err == nil
}

func paddingRight(text string, indent int) string {
	indentStr := strings.Repeat(" ", indent-len(text))

	return indentStr
}
