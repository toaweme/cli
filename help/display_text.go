package help

import (
	"fmt"
	"io"
	"strings"

	"github.com/toaweme/structs"

	"github.com/toaweme/cli"
)

// DisplayOptions controls how the text help output is formatted.
type DisplayOptions struct {
	ShowFlags bool
	ShowEnv   bool
	// ShowValues annotates each command flag with its resolved value (prefix-masked).
	// The values are read from the command's Options() struct, which the app populates
	// before rendering when --help-values is passed.
	ShowValues bool
	// GlobalValues is the populated global flags struct, rendered (with ShowValues) for the
	// Global Options block so flags like --cwd show their set value.
	// Nil falls back to a zero struct.
	GlobalValues *cli.GlobalFlags
	// Formats are extra --help-format values (the codecs registered via App.HelpOutputs)
	// appended to the built-in ones in the global options' --help-format hint.
	Formats []string
	// DefaultCommand is the name of the command run on a bare invocation (App.Default).
	// When set, the all-commands listing notes it above the Commands block. Empty when
	// no default is registered.
	DefaultCommand string
}

// DisplayHelp renders command help as text to w.
func DisplayHelp(w io.Writer, appName string, commands []cli.Command[any], command []string, opts ...DisplayOptions) {
	var displayOpts DisplayOptions
	if len(opts) > 0 {
		displayOpts = opts[0]
	}

	var help []string
	if len(command) == 0 {
		help = displayAllCommandsHelp(appName, commands, displayOpts)
	} else {
		help = displaySingleCommandHelp(w, appName, commands, command, displayOpts)
	}

	help = append(help, ``, `Global Options:`)

	globalOpts, err := helpOptionsWithEnv(globalSource(displayOpts.GlobalValues), displayOpts.ShowEnv, displayOpts.ShowValues, displayOpts.Formats)
	if err != nil {
		fmt.Fprintf(w, "Error printing global options: %v", err)
	}
	help = append(help, globalOpts...)

	fmt.Fprintln(w, strings.Join(help, "\n"))
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

func displaySingleCommandHelp(w io.Writer, appName string, commands []cli.Command[any], command []string, opts DisplayOptions) []string {
	help := []string{
		`Usage: ` + appName + ` <command> <subcommand> [args] [options]`,
	}

	cmd := findCommandByArgs(commands, command)
	if cmd == nil {
		_, _ = fmt.Fprintln(w, "Command not found")
		return []string{}
	}

	cmdHelp := cmd.Help()
	if cmdHelp != "" {
		help = append(help, cmdHelp)
	}
	if desc := commandDescription(cmd); desc != "" {
		help = append(help, ``)
		help = append(help, strings.Split(desc, "\n")...)
	}
	help = append(help, ``)
	line := `$ ` + strings.Join(command, " ")
	help = append(help, line)

	options, _ := helpOptionsWithEnv(cmd.Options(), false, opts.ShowValues, nil)
	if len(options) > 0 {
		help = append(help, options...)
	}

	help = append(help, providerDocLines(cmd, "")...)

	if len(cmd.Commands()) > 0 {
		longestName := getLongestName(cmd.Commands())
		for _, subCmd := range cmd.Commands() {
			name := subCmd.Name("")
			help = append(help, fmt.Sprintf(`  %s  %s%s`, name, pad(name, longestName), firstLine(subCmd.Help())))

			if opts.ShowFlags {
				help = appendCommandFlags(help, subCmd, opts)
			}
		}
	}

	return help
}

func displayAllCommandsHelp(appName string, commands []cli.Command[any], opts DisplayOptions) []string {
	help := []string{
		`Usage: ` + appName + ` <command> <subcommand> [args] [options]`,
	}
	help = append(help, ``,
		`Options can be passed before or after the command and subcommand.`,
		`Both -[opt] <arg> and --[opt]=<arg> are supported.`,
		`Boolean flags can be passed without an argument to set them to true.`,
	)

	if opts.DefaultCommand != "" {
		// show that a bare invocation runs the default command:  full → full build
		help = append(help, ``,
			`Default command:`,
			fmt.Sprintf(`  %s → %s %s`, appName, appName, opts.DefaultCommand),
		)
	}

	help = append(help, ``, `Commands:`)

	longestName := getLongestName(commands)

	for _, cmd := range commands {
		name := cmd.Name("")
		help = append(help, fmt.Sprintf(`  %s  %s%s`, name, pad(name, longestName), firstLine(cmd.Help())))

		if opts.ShowFlags {
			help = appendCommandFlags(help, cmd, opts)
		}

		if len(cmd.Commands()) > 0 {
			for _, subCmd := range cmd.Commands() {
				subName := name + " " + subCmd.Name("")
				help = append(help, `  `+subName+``+pad(subName, longestName)+`  `+firstLine(subCmd.Help()))

				if opts.ShowFlags {
					help = appendCommandFlags(help, subCmd, opts)
				}
			}
		}
	}

	return help
}

func appendCommandFlags(help []string, cmd cli.Command[any], opts DisplayOptions) []string {
	cmdOpts, err := helpOptionsWithEnv(cmd.Options(), opts.ShowEnv, opts.ShowValues, nil)
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

// flagShortPrefix is the leading short segment of a flag's display name: "-h, " for a flag with
// a short, "" for one without. The trailing ", " separates it from the long name.
func flagShortPrefix(short string) string {
	if short == "" {
		return ""
	}
	return "-" + short + ", "
}

// shortColWidth is the width of the reserved short-flag column: the widest "-x, " prefix across
// all fields and their nested sub-fields. Flags without a short are left-padded to this width so
// every "--long" name lines up, the way Cobra/clap/Click render their option lists. Multi-letter
// shorts (e.g. -vv, -vvv from cli.Verbosity) widen the column accordingly. Returns 0 when no field
// has a short, so flags render flush-left with no reserved column.
func shortColWidth(fields []structs.Field) int {
	w := 0
	for _, field := range fields {
		if n := len(flagShortPrefix(field.Tags["short"])); n > w {
			w = n
		}
		for _, sub := range field.Fields {
			if n := len(flagShortPrefix(sub.Tags["short"])); n > w {
				w = n
			}
		}
	}
	return w
}

// flagArgs renders a flag's display name with the short column padded to shortColW, so the
// "--long" names align whether or not the flag has a short. A short-only flag (no long arg)
// renders as just "-x".
func flagArgs(arg, short string, shortColW int) string {
	prefix := flagShortPrefix(short)
	if arg == "" {
		return "-" + short
	}
	return prefix + strings.Repeat(" ", shortColW-len(prefix)) + "--" + arg
}

func printableFieldsWithEnv(fields []structs.Field, showEnv, showValues bool, extraFormats []string) []string {
	lines := []string{}
	shortColW := shortColWidth(fields)
	longestArg := maxLen(fields, shortColW)

	// resolved values get their own aligned column between the flag and the description
	// (rather than trailing after the help text), to match the tables.
	valueColW := 0
	if showValues {
		for _, field := range fields {
			if vc := valueColumn(field); len(vc) > valueColW {
				valueColW = len(vc)
			}
		}
	}

	for _, field := range fields {
		if field.Tags["arg"] == "" && field.Tags["short"] == "" {
			continue
		}
		if isPositionalArg(field.Tags["arg"]) {
			continue
		}
		args := flagArgs(field.Tags["arg"], field.Tags["short"], shortColW)

		helpText := withAllowedValues(field.Tags["help"], field, formatHintExtras(field, extraFormats))
		if showEnv && field.Tags["env"] != "" {
			helpText += fmt.Sprintf(" [env: %s]", field.Tags["env"])
		}

		// a group header (a nested struct with sub-fields) wraps its name in brackets; the
		// reserved short-column padding is trimmed inside the brackets so they read cleanly.
		display := args
		if len(field.Fields) > 0 {
			display = "[" + strings.TrimLeft(args, " ") + "]"
		}
		flagBlock := fmt.Sprintf(`  %s  %s`, display, pad(display, longestArg))

		var line string
		if valueColW > 0 {
			line = fmt.Sprintf(`%s    %-*s  %s`, flagBlock, valueColW, valueColumn(field), helpText)
		} else if len(field.Fields) == 0 {
			line = fmt.Sprintf(`%s    %s`, flagBlock, helpText)
		} else {
			line = fmt.Sprintf(`%s  %s`, flagBlock, helpText)
		}
		lines = append(lines, strings.TrimRight(line, " "))

		for _, subField := range field.Fields {
			subArgs := flagArgs(subField.Tags["arg"], subField.Tags["short"], shortColW)
			padding := pad(subArgs, longestArg)
			subHelp := "  - " + withAllowedValues(subField.Tags["help"], subField, formatHintExtras(subField, extraFormats))
			if showEnv && subField.Tags["env"] != "" {
				subHelp += fmt.Sprintf(" [env: %s]", subField.Tags["env"])
			}
			line := fmt.Sprintf(`    %s  %s%s`, subArgs, padding, subHelp)
			lines = append(lines, line)
		}
	}

	return lines
}

func maxLen(fields []structs.Field, shortColW int) int {
	longestArg := 0

	for _, field := range fields {
		if n := len(flagArgs(field.Tags["arg"], field.Tags["short"], shortColW)); n > longestArg {
			longestArg = n
		}
		for _, subField := range field.Fields {
			if n := len(flagArgs(subField.Tags["arg"], subField.Tags["short"], shortColW)); n > longestArg {
				longestArg = n
			}
		}
	}

	longestArg += 2
	return longestArg
}

func helpOptionsWithEnv(structure any, showEnv, showValues bool, extraFormats []string) ([]string, error) {
	fields, err := structs.GetStructFields(structure, nil, structs.DefaultEncodingTags)
	if err != nil {
		return nil, fmt.Errorf("failed to get struct fields: %w", err)
	}

	return printableFieldsWithEnv(fields, showEnv, showValues, extraFormats), nil
}

func pad(text string, indent int) string {
	return strings.Repeat(" ", indent-len(text))
}
