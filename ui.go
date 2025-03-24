package cli

import (
	"fmt"
	"strings"

	"github.com/toaweme/structs"
)

func DisplayHelp(appName string, commands []Command[any], command []string) {
	help := []string{`Usage: ` + appName + ` <command> <subcommand> [args] [options]`}
	help = append(help, ``)

	// if len(commands) == 1 {
	// 	help = displaySingleCommandHelp(appName, commands, command)
	// } else if len(command) == 0 {
	// 	help = displayAllCommandsHelp(appName, commands)
	// } else {
	// 	help = displaySingleCommandHelp(appName, commands, command)
	// }
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

	cmdHelp := cmd.Help()
	if cmdHelp != "" {
		help = append(help, cmdHelp)
		help = append(help, ``)
	}
	line := fmt.Sprintf(`$ %s`, strings.Join(command, " "))
	help = append(help, line)

	options, _ := helpOptions(cmd.Options())
	if len(options) > 0 {
		help = append(help, options...)
	}

	if len(cmd.Commands()) > 0 {
		// help = append(help, ``)
		// help = append(help, `Subcommands:`)
		longestName := getLongestName(cmd.Commands())
		for _, subCmd := range cmd.Commands() {
			name := subCmd.Name("")
			help = append(help, fmt.Sprintf(`  %s  %s%s`, name, pad(name, longestName), subCmd.Help()))
		}
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

		help = append(help, fmt.Sprintf(`  %s  %s%s`, name, pad(name, longestName), cmd.Help()))

		if len(cmd.Commands()) > 0 {
			for _, subCmd := range cmd.Commands() {
				subName := name + " " + subCmd.Name("")
				help = append(help, `  `+subName+``+pad(subName, longestName)+`  `+subCmd.Help())
			}
		}
	}

	return help
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

type helpOption struct {
	Args string
	Help string
}

func newHelpOption(arg, short, help string) helpOption {
	args := fmt.Sprintf(`-%s, --%s`, short, arg)
	if short == "" {
		args = fmt.Sprintf(`--%s`, arg)
	} else if arg == "" {
		args = fmt.Sprintf(`-%s`, short)
	}

	return helpOption{
		Args: args,
		Help: help,
	}
}

func printableFields(fields []structs.Field) []string {
	lines := []string{}
	longestArg := maxLen(fields)

	for _, field := range fields {
		if field.Tags["arg"] == "" && field.Tags["short"] == "" {
			continue
		}
		opt := newHelpOption(field.Tags["arg"], field.Tags["short"], field.Tags["help"])
		padding := pad(opt.Args, longestArg)
		line := ""
		if len(field.Fields) == 0 {
			line = fmt.Sprintf(`  %s  %s    %s`, opt.Args, padding, opt.Help)
		} else {
			line = fmt.Sprintf(`  [%s]  %s  %s`, opt.Args, padding, opt.Help)
		}
		lines = append(lines, line)
		// top to bottom right char: └

		for _, subField := range field.Fields {
			opt := newHelpOption(subField.Tags["arg"], subField.Tags["short"], subField.Tags["help"])
			padding := pad(opt.Args, longestArg)
			line := fmt.Sprintf(`    %s  %s%s`, opt.Args, padding, "  - "+opt.Help)
			lines = append(lines, line)
		}
	}

	return lines
}

func maxLen(fields []structs.Field) int {
	longestArg := 0

	for _, field := range fields {
		opt := newHelpOption(field.Tags["arg"], field.Tags["short"], field.Tags["help"])
		if len(opt.Args) > longestArg {
			// slog.Info("longestArg", "len", longestArg, "opt.Args", opt.Args)
			longestArg = len(opt.Args)
		}
		for _, subField := range field.Fields {
			opt := newHelpOption(subField.Tags["arg"], subField.Tags["short"], subField.Tags["help"])
			if len(opt.Args) > longestArg {
				longestArg = len(opt.Args)
			}
		}
	}

	longestArg += 2
	return longestArg
}

func helpOptions(structure any) ([]string, error) {
	fields, err := structs.GetStructFields(structure, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting global option fields: %w", err)
	}

	return printableFields(fields), nil
}

func pad(text string, indent int) string {
	indentStr := strings.Repeat(" ", indent-len(text))

	return indentStr
}
