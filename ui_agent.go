package cli

import (
	"fmt"
	"strings"

	"github.com/toaweme/structs"
)

// AgentOptions controls the agent help output.
type AgentOptions struct {
	AppName  string
	Format   string
	Filter   []string
	Commands []Command[any]
}

// DisplayHelpAgent renders comprehensive markdown help for all commands,
// then passes through the format renderer (pretty, plain, or raw md).
func DisplayHelpAgent(opts AgentOptions) {
	commands := opts.Commands
	if len(opts.Filter) > 0 {
		commands = filterCommands(commands, opts.Filter)
	}

	format := resolveFormat(opts.Format)

	if format == "pretty" && !isTTY() {
		format = "plain"
	}

	// pretty uses the same markdown source as md, then renders with ANSI
	buildFormat := format
	if format == "pretty" {
		buildFormat = "md"
	}

	output := buildAgentOutput(opts.AppName, commands, buildFormat)

	if format == "pretty" {
		fmt.Print(prettyMarkdown(output))
	} else {
		fmt.Print(output)
	}
}

func buildAgentOutput(appName string, commands []Command[any], format string) string {
	var b strings.Builder

	for _, cmd := range commands {
		writeAgentCommand(&b, cmd, "", appName, format)
	}

	if format == "md" || format == "pretty" {
		b.WriteString("## Global Options\n")
	} else {
		b.WriteString("Global Options\n")
	}
	writeAgentFlagBlock(&b, &GlobalOptions{}, "  ", format)

	return b.String()
}

func writeAgentCommand(b *strings.Builder, cmd Command[any], prefix, appName, format string) {
	name := prefix + cmd.Name("")
	help := cmd.Help()

	if format == "md" || format == "pretty" {
		b.WriteString(fmt.Sprintf("## %s\n", name))
	} else {
		b.WriteString(fmt.Sprintf("%s\n", name))
	}
	if help != "" {
		b.WriteString(fmt.Sprintf("  %s\n", help))
	}

	rows := extractFlagRows(cmd.Options())
	if len(rows) > 0 {
		writeAgentFlagRows(b, rows, "  ", format)
	}

	examples := commandExamples(cmd, name, appName)
	if len(examples) > 0 {
		if format == "md" {
			b.WriteString("  ```shell\n")
		}
		for _, ex := range examples {
			b.WriteString(fmt.Sprintf("  ❯ %s\n", ex))
		}
		if format == "md" {
			b.WriteString("  ```\n")
		}
	}

	for _, sub := range cmd.Commands() {
		writeAgentCommand(b, sub, name+" ", appName, format)
	}
}

// flagRow holds the individual columns for a single flag entry.
type flagRow struct {
	Flag     string
	Short    string
	Type     string
	Help     string
	Env      string
	Required bool
	Default  string
}

func extractFlagRows(options any) []flagRow {
	if options == nil {
		return nil
	}

	fields, err := structs.GetStructFields(options, nil)
	if err != nil {
		return nil
	}

	var rows []flagRow
	for _, field := range fields {
		if field.Tags["arg"] == "" && field.Tags["short"] == "" {
			continue
		}

		rows = append(rows, flagRow{
			Flag:     field.Tags["arg"],
			Short:    field.Tags["short"],
			Type:     field.Type,
			Help:     field.Tags["help"],
			Env:      field.Tags["env"],
			Required: field.Tags["validate"] == "required",
			Default:  field.Tags["default"],
		})
	}

	return rows
}

func renderFlagTablePlain(rows []flagRow, indent string) string {
	if len(rows) == 0 {
		return ""
	}

	flagW, typeW, helpW := computeColWidths(rows, false)

	hasEnv := false
	for _, r := range rows {
		if r.Env != "" {
			hasEnv = true
			break
		}
	}

	var b strings.Builder
	for _, r := range rows {
		flag := flagColPlain(r)
		typ := typeCol(r)
		line := fmt.Sprintf("%-*s  %-*s  %-*s", flagW, flag, typeW, typ, helpW, r.Help)
		if hasEnv {
			line += fmt.Sprintf("  %s", r.Env)
		}
		b.WriteString(indent + strings.TrimRight(line, " ") + "\n")
	}
	return b.String()
}

func flagColPlain(r flagRow) string {
	if r.Short != "" {
		return fmt.Sprintf("--%s, -%s", r.Flag, r.Short)
	}
	return fmt.Sprintf("--%s", r.Flag)
}

func computeColWidths(rows []flagRow, markdown bool) (int, int, int) {
	flagW, typeW, helpW := 0, 0, 0
	for _, r := range rows {
		var f string
		if markdown {
			f = flagCol(r)
		} else {
			f = flagColPlain(r)
		}
		if len(f) > flagW {
			flagW = len(f)
		}
		t := typeCol(r)
		if len(t) > typeW {
			typeW = len(t)
		}
		if len(r.Help) > helpW {
			helpW = len(r.Help)
		}
	}
	return flagW, typeW, helpW
}

func renderFlagTableMd(rows []flagRow, indent string) string {
	if len(rows) == 0 {
		return ""
	}

	flagW, typeW, helpW := computeColWidths(rows, true)

	hasEnv := false
	for _, r := range rows {
		if r.Env != "" {
			hasEnv = true
			break
		}
	}

	var b strings.Builder

	// header
	header := fmt.Sprintf("| %-*s | %-*s | %-*s |", flagW, "Flag", typeW, "Type", helpW, "Description")
	if hasEnv {
		envW := envColWidth(rows)
		header = fmt.Sprintf("| %-*s | %-*s | %-*s | %-*s |", flagW, "Flag", typeW, "Type", helpW, "Description", envW, "Env")
	}
	b.WriteString(indent + header + "\n")

	// separator
	sep := fmt.Sprintf("| %s | %s | %s |", strings.Repeat("-", flagW), strings.Repeat("-", typeW), strings.Repeat("-", helpW))
	if hasEnv {
		envW := envColWidth(rows)
		sep = fmt.Sprintf("| %s | %s | %s | %s |", strings.Repeat("-", flagW), strings.Repeat("-", typeW), strings.Repeat("-", helpW), strings.Repeat("-", envW))
	}
	b.WriteString(indent + sep + "\n")

	// rows
	for _, r := range rows {
		row := fmt.Sprintf("| %-*s | %-*s | %-*s |", flagW, flagCol(r), typeW, typeCol(r), helpW, r.Help)
		if hasEnv {
			envW := envColWidth(rows)
			envVal := r.Env
			if envVal != "" {
				envVal = "`" + envVal + "`"
			}
			row = fmt.Sprintf("| %-*s | %-*s | %-*s | %-*s |", flagW, flagCol(r), typeW, typeCol(r), helpW, r.Help, envW, envVal)
		}
		b.WriteString(indent + row + "\n")
	}

	return b.String()
}

func flagCol(r flagRow) string {
	if r.Short != "" {
		return fmt.Sprintf("`--%s`, `-%s`", r.Flag, r.Short)
	}
	return fmt.Sprintf("`--%s`", r.Flag)
}

func typeCol(r flagRow) string {
	t := r.Type
	if r.Required {
		t += ", required"
	}
	if r.Default != "" {
		t += ", default: " + r.Default
	}
	return t
}

func envColWidth(rows []flagRow) int {
	w := 3 // "Env"
	for _, r := range rows {
		v := r.Env
		if v != "" {
			v = "`" + v + "`"
		}
		if len(v) > w {
			w = len(v)
		}
	}
	return w
}

func commandExamples(cmd Command[any], fullName, appName string) []string {
	if ep, ok := cmd.(ExampleProvider); ok {
		return ep.Examples()
	}

	flags := extractExampleFlags(cmd.Options())
	if len(flags) == 0 {
		return nil
	}

	return []string{appName + " " + fullName + flags}
}

func extractExampleFlags(options any) string {
	if options == nil {
		return ""
	}

	fields, err := structs.GetStructFields(options, nil)
	if err != nil {
		return ""
	}

	var parts []string
	for _, field := range fields {
		arg := field.Tags["arg"]
		if arg == "" {
			continue
		}

		// positional args show as <name>
		if arg == "0" || arg == "1" || arg == "2" {
			helpTag := field.Tags["help"]
			if helpTag != "" {
				parts = append(parts, "<"+strings.ToLower(strings.ReplaceAll(helpTag, " ", "-"))+">")
			} else {
				parts = append(parts, "<arg>")
			}
			continue
		}

		switch field.Type {
		case "bool":
			parts = append(parts, "--"+arg)
		case "string":
			parts = append(parts, "--"+arg+"=<value>")
		default:
			parts = append(parts, fmt.Sprintf("--%s=<%s>", arg, field.Type))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return " " + strings.Join(parts, " ")
}

func writeAgentFlagBlock(b *strings.Builder, options any, indent, format string) {
	rows := extractFlagRows(options)
	if len(rows) > 0 {
		writeAgentFlagRows(b, rows, indent, format)
	}
}

func writeAgentFlagRows(b *strings.Builder, rows []flagRow, indent, format string) {
	if format == "md" {
		b.WriteString(renderFlagTableMd(rows, indent))
		return
	}
	b.WriteString(renderFlagTablePlain(rows, indent))
}

func filterCommands(commands []Command[any], filters []string) []Command[any] {
	filterSet := make(map[string]bool, len(filters))
	for _, f := range filters {
		filterSet[strings.TrimSpace(f)] = true
	}

	var result []Command[any]
	for _, cmd := range commands {
		name := cmd.Name("")

		if filterSet[name] {
			result = append(result, cmd)
			continue
		}

		// check if any filter matches a subcommand path
		var matchedSubs []Command[any]
		for _, sub := range cmd.Commands() {
			subPath := name + " " + sub.Name("")
			if filterSet[subPath] || filterSet[sub.Name("")] {
				matchedSubs = append(matchedSubs, sub)
			}
		}

		if len(matchedSubs) > 0 {
			filtered := &filteredCommand{
				command: cmd,
				subs:    matchedSubs,
			}
			result = append(result, filtered)
		}
	}

	return result
}

// filteredCommand wraps a command to expose only a subset of its subcommands.
type filteredCommand struct {
	command Command[any]
	subs    []Command[any]
}

var _ Command[any] = (*filteredCommand)(nil)

func (f *filteredCommand) Name(name string) string        { return f.command.Name(name) }
func (f *filteredCommand) Add(name string, cmd Command[any]) { f.command.Add(name, cmd) }
func (f *filteredCommand) Options() any                    { return f.command.Options() }
func (f *filteredCommand) Commands() []Command[any]        { return f.subs }
func (f *filteredCommand) Run(o GlobalOptions, u Unknowns) error { return f.command.Run(o, u) }
func (f *filteredCommand) Validate(o map[string]any) error { return f.command.Validate(o) }
func (f *filteredCommand) Help() string                    { return f.command.Help() }
