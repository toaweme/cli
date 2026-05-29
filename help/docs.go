package help

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/toaweme/structs"

	"github.com/toaweme/cli"
)

// AgentOptions controls the comprehensive documentation output.
type AgentOptions struct {
	AppName  string
	Format   string
	Commands []cli.Command[any]
	// Formats are extra --format values (from cli.Config.Formats) appended to the
	// built-in ones in the global options' --format hint.
	Formats []string
}

// DisplayHelpAgent renders comprehensive documentation for all commands,
// including flag tables, env vars, and usage examples.
func DisplayHelpAgent(opts AgentOptions) {
	commands := opts.Commands
	format := resolveFormat(opts.Format)

	if format == "pretty" && !isTTY() {
		format = "plain"
	}

	buildFormat := format
	if format == "pretty" {
		buildFormat = "md"
	}

	output := buildAgentOutput(opts.AppName, commands, buildFormat, opts.Formats)

	if format == "pretty" {
		fmt.Print(prettyMarkdown(output))
	} else {
		fmt.Print(output)
	}
}

// buildAgentOutput generates the full documentation string for all commands.
// format controls whether markdown or plain text is emitted.
func buildAgentOutput(appName string, commands []cli.Command[any], format string, extraFormats []string) string {
	var b strings.Builder

	for _, cmd := range commands {
		writeAgentCommand(&b, cmd, "", appName, format)
	}

	if format == "md" || format == "pretty" {
		b.WriteString("## Global Options\n")
	} else {
		b.WriteString("Global Options\n")
	}
	writeGlobalOptionsBlock(&b, format, extraFormats)

	return b.String()
}

func writeAgentCommand(b *strings.Builder, cmd cli.Command[any], prefix, appName, format string) {
	name := prefix + cmd.Name("")
	help := cmd.Help()

	if format == "md" || format == "pretty" {
		b.WriteString(fmt.Sprintf("## %s\n", name))
	} else {
		b.WriteString(fmt.Sprintf("%s\n", name))
	}
	if help != "" {
		b.WriteString(fmt.Sprintf("  %s\n", firstLine(help)))
	}
	if desc := commandDescription(cmd); desc != "" {
		if format == "md" || format == "pretty" {
			b.WriteString("\n" + desc + "\n")
		} else {
			for _, line := range strings.Split(desc, "\n") {
				if line == "" {
					b.WriteString("\n")
				} else {
					b.WriteString("  " + line + "\n")
				}
			}
		}
	}

	rows := extractFlagRows(cmd.Options())
	if len(rows) > 0 {
		writeAgentFlagRows(b, rows, "  ", format)
	}

	for _, line := range providerDocLines(cmd, "  ") {
		b.WriteString(line + "\n")
	}

	examples := commandExamples(cmd, name, appName)
	if len(examples) > 0 {
		b.WriteString("\n  Examples:\n")
		if format == "md" {
			b.WriteString("  ```shell\n")
		}
		for _, ex := range examples {
			if len(ex) == 0 {
				continue
			}
			b.WriteString(fmt.Sprintf("  ❯ %s\n", ex[0]))
			for _, line := range ex[1:] {
				b.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
		if format == "md" {
			b.WriteString("  ```\n")
		}
	}

	for _, sub := range cmd.Commands() {
		writeAgentCommand(b, sub, name+" ", appName, format)
	}
}

// flagRow holds parsed flag metadata for table rendering.
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
	return extractFlagRowsWithFormats(options, nil)
}

// extractFlagRowsWithFormats is extractFlagRows with extra --format values to append
// to the format flag's allowed-values hint, used when rendering global options.
func extractFlagRowsWithFormats(options any, extraFormats []string) []flagRow {
	if options == nil {
		return nil
	}

	fields, err := structs.GetStructFields(options, nil, structs.DefaultEncodingTags)
	if err != nil {
		return nil
	}

	var rows []flagRow
	for _, field := range fields {
		rows = appendFlagRows(rows, field, extraFormats)
	}

	return rows
}

// appendFlagRows adds a row for field when it carries a flag tag, then recurses
// into nested struct sub-fields. Sub-fields are addressed by their dotted FQN tag
// (e.g. "database.host") and may carry their own oneof rule, so they render in the
// flag table the same way top-level flags do. extraFormats rides along on the
// --format field's allowed-values hint (see formatHintExtras).
func appendFlagRows(rows []flagRow, field structs.Field, extraFormats []string) []flagRow {
	if (field.Tags["arg"] != "" || field.Tags["short"] != "") && !isPositionalArg(field.Tags["arg"]) {
		rows = append(rows, flagRow{
			Flag:     flagArg(field),
			Short:    field.Tags["short"],
			Type:     displayType(field),
			Help:     withAllowedValues(field.Tags["help"], field, formatHintExtras(field, extraFormats)),
			Env:      flagEnv(field),
			Required: hasRule(field, "required"),
			Default:  field.Default,
		})
	}

	for _, sub := range field.Fields {
		rows = appendFlagRows(rows, sub, extraFormats)
	}

	return rows
}

// flagArg returns the flag name a user types for field: the dotted FQN tag for a
// nested sub-field (e.g. "database.host"), or the plain arg tag for a top-level field.
func flagArg(field structs.Field) string {
	if field.FQN != nil && field.FQN.Tags["arg"] != "" {
		return field.FQN.Tags["arg"]
	}
	return field.Tags["arg"]
}

// flagEnv returns the env var name for field, preferring the underscore-joined FQN
// env tag for nested sub-fields (e.g. "DATABASE_HOST") over the bare leaf tag.
func flagEnv(field structs.Field) string {
	if field.FQN != nil && field.FQN.Tags["env"] != "" {
		return field.FQN.Tags["env"]
	}
	return field.Tags["env"]
}

// displayType renders a field's type for help output, preferring the concrete Go
// type for slices ("[]string") over the bare reflect kind ("slice").
func displayType(field structs.Field) string {
	if field.Value.IsValid() && field.Value.Kind() == reflect.Slice {
		return field.Value.Type().String()
	}
	return field.Type
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

	header := fmt.Sprintf("| %-*s | %-*s | %-*s |", flagW, "Flag", typeW, "Type", helpW, "Description")
	if hasEnv {
		envW := envColWidth(rows)
		header = fmt.Sprintf("| %-*s | %-*s | %-*s | %-*s |", flagW, "Flag", typeW, "Type", helpW, "Description", envW, "Env")
	}
	b.WriteString(indent + header + "\n")

	sep := fmt.Sprintf("| %s | %s | %s |", strings.Repeat("-", flagW), strings.Repeat("-", typeW), strings.Repeat("-", helpW))
	if hasEnv {
		envW := envColWidth(rows)
		sep = fmt.Sprintf("| %s | %s | %s | %s |", strings.Repeat("-", flagW), strings.Repeat("-", typeW), strings.Repeat("-", helpW), strings.Repeat("-", envW))
	}
	b.WriteString(indent + sep + "\n")

	for _, r := range rows {
		row := fmt.Sprintf("| %-*s | %-*s | %-*s |", flagW, flagCol(r), typeW, typeCol(r), helpW, r.Help)
		if hasEnv {
			envW := envColWidth(rows)
			envVal := envColValue(r)
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

func envColValue(r flagRow) string {
	if r.Env == "" {
		return ""
	}
	if r.Default != "" {
		return fmt.Sprintf("`%s`=*%s*", r.Env, r.Default)
	}
	return "`" + r.Env + "`"
}

func envColWidth(rows []flagRow) int {
	w := 3
	for _, r := range rows {
		v := envColValue(r)
		if len(v) > w {
			w = len(v)
		}
	}
	return w
}

func extractExampleFlags(options any) string {
	if options == nil {
		return ""
	}

	fields, err := structs.GetStructFields(options, nil, structs.DefaultEncodingTags)
	if err != nil {
		return ""
	}

	var parts []string
	for _, field := range fields {
		arg := field.Tags["arg"]
		if arg == "" {
			continue
		}

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
			parts = append(parts, fmt.Sprintf("--%s=<%s>", arg, displayType(field)))
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

func writeGlobalOptionsBlock(b *strings.Builder, format string, extraFormats []string) {
	indent := "  "
	rows := extractFlagRowsWithFormats(&cli.GlobalOptions{}, extraFormats)
	if len(rows) == 0 {
		return
	}

	// allowed values for flags like --format ride along in the Help column as a
	// "(one of: ...)" hint sourced from the field's oneof rule, so the block is
	// just the flag table.
	writeAgentFlagRows(b, rows, indent, format)
}

func writeAgentFlagRows(b *strings.Builder, rows []flagRow, indent, format string) {
	if format == "md" {
		b.WriteString(renderFlagTableMd(rows, indent))
		return
	}
	b.WriteString(renderFlagTablePlain(rows, indent))
}

// FilterCommands returns only the commands matching the filter list.
// Supports top-level names ("build") and subcommand paths ("db migrate").
// Parent commands are included with only their matched subcommands.
func FilterCommands(commands []cli.Command[any], filters []string) []cli.Command[any] {
	filterSet := make(map[string]bool, len(filters))
	for _, f := range filters {
		filterSet[strings.TrimSpace(f)] = true
	}

	var result []cli.Command[any]
	for _, cmd := range commands {
		name := cmd.Name("")

		if filterSet[name] {
			result = append(result, cmd)
			continue
		}

		var matchedSubs []cli.Command[any]
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

// filteredCommand wraps a parent command to expose only a subset of its subcommands.
// Used by filterCommands to narrow output without mutating the original command tree.
type filteredCommand struct {
	command cli.Command[any]
	subs    []cli.Command[any]
}

var _ cli.Command[any] = (*filteredCommand)(nil)

func (f *filteredCommand) Name(name string) string                       { return f.command.Name(name) }
func (f *filteredCommand) Add(name string, cmd cli.Command[any])         { f.command.Add(name, cmd) }
func (f *filteredCommand) Options() any                                  { return f.command.Options() }
func (f *filteredCommand) Commands() []cli.Command[any]                  { return f.subs }
func (f *filteredCommand) Run(o cli.GlobalOptions, u cli.Unknowns) error { return f.command.Run(o, u) }
func (f *filteredCommand) Validate(o map[string]any) error               { return f.command.Validate(o) }
func (f *filteredCommand) Help() string                                  { return f.command.Help() }
func (f *filteredCommand) Description() string                           { return f.command.Description() }
func (f *filteredCommand) Examples() [][]string                          { return f.command.Examples() }
func (f *filteredCommand) Args() map[int][]string                        { return f.command.Args() }
func (f *filteredCommand) Flags() map[string][]string                    { return f.command.Flags() }
func (f *filteredCommand) ConfigStrategy() (cli.MergeStrategy, cli.ConfigMapping) {
	return f.command.ConfigStrategy()
}
