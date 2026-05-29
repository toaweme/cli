package help

import (
	"fmt"
	"strings"

	"github.com/toaweme/cli"
	"github.com/toaweme/structs"
)

// HelpDisplayOptions controls how the text help output is formatted.
type HelpDisplayOptions struct {
	ShowFlags bool
	ShowEnv   bool
}

// DisplayHelp renders command help to stdout in text format.
func DisplayHelp(appName string, commands []cli.Command[any], command []string, opts ...HelpDisplayOptions) {
	var displayOpts HelpDisplayOptions
	if len(opts) > 0 {
		displayOpts = opts[0]
	}

	var help []string
	if len(command) == 0 {
		help = displayAllCommandsHelp(appName, commands, displayOpts)
	} else {
		help = displaySingleCommandHelp(appName, commands, command, displayOpts)
	}

	help = append(help, ``, `Global Options:`)

	globalOpts, err := helpOptionsWithEnv(&cli.GlobalOptions{}, displayOpts.ShowEnv)
	if err != nil {
		fmt.Printf("Error printing global options: %v", err)
	}
	help = append(help, globalOpts...)

	fmt.Println(strings.Join(help, "\n"))
}

// findCommandByArgs walks the command tree to find the command matching the arg path.
func findCommandByArgs(commands []cli.Command[any], args []string) cli.Command[any] {
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

func displaySingleCommandHelp(appName string, commands []cli.Command[any], command []string, opts HelpDisplayOptions) []string {
	help := []string{
		fmt.Sprintf(`Usage: ` + appName + ` <command> <subcommand> [args] [options]`),
	}

	cmd := findCommandByArgs(commands, command)
	if cmd == nil {
		fmt.Println("Command not found")
		return []string{}
	}

	cmdHelp := cmd.Help()
	if cmdHelp != "" {
		help = append(help, cmdHelp, ``)
	}
	line := fmt.Sprintf(`$ %s`, strings.Join(command, " "))
	help = append(help, line)

	options, _ := helpOptions(cmd.Options())
	if len(options) > 0 {
		help = append(help, options...)
	}

	if len(cmd.Commands()) > 0 {
		longestName := getLongestName(cmd.Commands())
		for _, subCmd := range cmd.Commands() {
			name := subCmd.Name("")
			help = append(help, fmt.Sprintf(`  %s  %s%s`, name, pad(name, longestName), subCmd.Help()))

			if opts.ShowFlags {
				help = appendCommandFlags(help, subCmd, opts)
			}
		}
	}

	return help
}

func displayAllCommandsHelp(appName string, commands []cli.Command[any], opts HelpDisplayOptions) []string {
	help := []string{
		fmt.Sprintf(`Usage: ` + appName + ` <command> <subcommand> [args] [options]`),
	}
	help = append(help, ``,
		`Options can be passed before or after the command and subcommand.`,
		`Both -[opt] <arg> and --[opt]=<arg> are supported.`,
		`Boolean flags can be passed without an argument to set them to true.`,
		``,
		`Commands:`,
	)

	longestName := getLongestName(commands)

	for _, cmd := range commands {
		name := cmd.Name("")
		help = append(help, fmt.Sprintf(`  %s  %s%s`, name, pad(name, longestName), cmd.Help()))

		if opts.ShowFlags {
			help = appendCommandFlags(help, cmd, opts)
		}

		if len(cmd.Commands()) > 0 {
			for _, subCmd := range cmd.Commands() {
				subName := name + " " + subCmd.Name("")
				help = append(help, `  `+subName+``+pad(subName, longestName)+`  `+subCmd.Help())

				if opts.ShowFlags {
					help = appendCommandFlags(help, subCmd, opts)
				}
			}
		}
	}

	return help
}

func appendCommandFlags(help []string, cmd cli.Command[any], opts HelpDisplayOptions) []string {
	cmdOpts, err := helpOptionsWithEnv(cmd.Options(), opts.ShowEnv)
	if err != nil || len(cmdOpts) == 0 {
		return help
	}

	for _, line := range cmdOpts {
		help = append(help, "    "+line)
	}

	return help
}

func getLongestName(commands []cli.Command[any]) int {
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
	return printableFieldsWithEnv(fields, false)
}

func printableFieldsWithEnv(fields []structs.Field, showEnv bool) []string {
	lines := []string{}
	longestArg := maxLen(fields)

	for _, field := range fields {
		if field.Tags["arg"] == "" && field.Tags["short"] == "" {
			continue
		}
		if isPositionalArg(field.Tags["arg"]) {
			continue
		}
		opt := newHelpOption(field.Tags["arg"], field.Tags["short"], field.Tags["help"])
		padding := pad(opt.Args, longestArg)

		helpText := opt.Help
		if showEnv && field.Tags["env"] != "" {
			helpText += fmt.Sprintf(" [env: %s]", field.Tags["env"])
		}

		var line string
		if len(field.Fields) == 0 {
			line = fmt.Sprintf(`  %s  %s    %s`, opt.Args, padding, helpText)
		} else {
			line = fmt.Sprintf(`  [%s]  %s  %s`, opt.Args, padding, helpText)
		}
		lines = append(lines, line)

		for _, subField := range field.Fields {
			opt := newHelpOption(subField.Tags["arg"], subField.Tags["short"], subField.Tags["help"])
			padding := pad(opt.Args, longestArg)
			subHelp := "  - " + opt.Help
			if showEnv && subField.Tags["env"] != "" {
				subHelp += fmt.Sprintf(" [env: %s]", subField.Tags["env"])
			}
			line := fmt.Sprintf(`    %s  %s%s`, opt.Args, padding, subHelp)
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
	return helpOptionsWithEnv(structure, false)
}

func helpOptionsWithEnv(structure any, showEnv bool) ([]string, error) {
	fields, err := structs.GetStructFields(structure, nil, structs.DefaultEncodingTags)
	if err != nil {
		return nil, fmt.Errorf("failed to get struct fields: %w", err)
	}

	return printableFieldsWithEnv(fields, showEnv), nil
}

func pad(text string, indent int) string {
	return strings.Repeat(" ", indent-len(text))
}

func isPositionalArg(arg string) bool {
	if arg == "" {
		return false
	}
	for _, c := range arg {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// hasRule checks whether a struct field has a specific validation rule (e.g. "required").
func hasRule(field structs.Field, name string) bool {
	for _, r := range field.Rules {
		if r.Name == name {
			return true
		}
	}
	return false
}
