package help

import (
	"fmt"
	"io"
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
	// Formats are extra --help-format values (from cli.Config.Formats) appended to the
	// built-in ones in the global options' --help-format hint.
	Formats []string
	// ShowValues annotates each flag with its resolved value (secret fields masked),
	// read from the command's Options() struct the app populates before rendering.
	ShowValues bool
	// GlobalValues is the populated global flags struct, rendered (with ShowValues)
	// for the Global Options block so flags like --verbosity show their set value.
	// Nil falls back to a zero struct, so only the flag definitions are shown.
	GlobalValues *cli.GlobalFlags
}

// DisplayHelpAgent renders comprehensive documentation for all commands to w,
// including flag tables, env vars, and usage examples.
func DisplayHelpAgent(w io.Writer, opts AgentOptions) {
	commands := opts.Commands
	format := resolveFormat(opts.Format)

	if format == "pretty" && !isTTY() {
		format = "plain"
	}

	buildFormat := format
	if format == "pretty" {
		buildFormat = "md"
	}

	output := buildAgentOutput(opts.AppName, commands, buildFormat, opts.Formats, opts.ShowValues, opts.GlobalValues)

	if format == "pretty" {
		fmt.Fprint(w, prettyMarkdown(output))
	} else {
		fmt.Fprint(w, output)
	}
}

// buildAgentOutput generates the full documentation string for all commands.
// format controls whether markdown or plain text is emitted.
func buildAgentOutput(appName string, commands []cli.Command[any], format string, extraFormats []string, showValues bool, globalValues *cli.GlobalFlags) string {
	var b strings.Builder

	for _, cmd := range commands {
		writeAgentCommand(&b, cmd, "", appName, format, showValues)
	}

	if format == "md" || format == "pretty" {
		b.WriteString("## Global Options\n")
	} else {
		b.WriteString("Global Options\n")
	}
	writeGlobalFlagsBlock(&b, format, extraFormats, globalValues, showValues)

	return b.String()
}

func writeAgentCommand(b *strings.Builder, cmd cli.Command[any], prefix, appName, format string, showValues bool) {
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

	rows := extractFlagRows(cmd.Options(), showValues)
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
		writeAgentCommand(b, sub, name+" ", appName, format, showValues)
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
	// Value is the bare resolved value for --help-values (e.g. `(8080)`), empty when
	// not in that mode or the flag is unset. Rendered inside the Type column (so the
	// type is not duplicated), dimmed in the pretty path.
	Value string
}

func extractFlagRows(options any, showValues bool) []flagRow {
	return extractFlagRowsWithFormats(options, nil, showValues)
}

// extractFlagRowsWithFormats is extractFlagRows with extra --help-format values to append
// to the format flag's allowed-values hint, used when rendering global options.
func extractFlagRowsWithFormats(options any, extraFormats []string, showValues bool) []flagRow {
	if options == nil {
		return nil
	}

	fields, err := structs.GetStructFields(options, nil, structs.DefaultEncodingTags)
	if err != nil {
		return nil
	}

	var rows []flagRow
	for _, field := range fields {
		rows = appendFlagRows(rows, field, extraFormats, showValues)
	}

	return rows
}

// appendFlagRows adds a row for field when it carries a flag tag, then recurses
// into nested struct sub-fields. Sub-fields are addressed by their dotted FQN tag
// (e.g. "database.host") and may carry their own oneof rule, so they render in the
// flag table the same way top-level flags do. extraFormats rides along on the
// --help-format field's allowed-values hint (see formatHintExtras).
func appendFlagRows(rows []flagRow, field structs.Field, extraFormats []string, showValues bool) []flagRow {
	if (field.Tags["arg"] != "" || field.Tags["short"] != "") && !isPositionalArg(field.Tags["arg"]) {
		value := ""
		if showValues {
			value = valueText(field)
		}
		rows = append(rows, flagRow{
			Flag:     flagArg(field),
			Short:    field.Tags["short"],
			Type:     displayType(field),
			Help:     withAllowedValues(field.Tags["help"], field, formatHintExtras(field, extraFormats)),
			Env:      flagEnv(field),
			Required: hasRule(field, "required"),
			Default:  field.Default,
			Value:    value,
		})
	}

	for _, sub := range field.Fields {
		rows = appendFlagRows(rows, sub, extraFormats, showValues)
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

// flagTableColumns assembles the columns of the flag table in display order:
// Flag, [Env], Type, [Value], Description. Env and Value are present only when some
// row carries one. markdown selects the rendered cell form (backticks, emphasis).
// It returns the column headers, the per-row cells, and each column's width.
func flagTableColumns(rows []flagRow, markdown bool) (headers []string, cells [][]string, widths []int) {
	hasEnv := anyRow(rows, func(r flagRow) bool { return r.Env != "" })
	hasValue := anyRow(rows, func(r flagRow) bool { return r.Value != "" })

	headers = []string{"Flag"}
	if hasEnv {
		headers = append(headers, "Env")
	}
	headers = append(headers, "Type")
	if hasValue {
		headers = append(headers, "Value")
	}
	headers = append(headers, "Description")

	for _, r := range rows {
		row := []string{flagColCell(r, markdown)}
		if hasEnv {
			row = append(row, envColCell(r, markdown))
		}
		row = append(row, typeCol(r))
		if hasValue {
			row = append(row, valueColCell(r, markdown))
		}
		row = append(row, descCol(r, markdown))
		cells = append(cells, row)
	}

	widths = make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range cells {
		for i, c := range row {
			if len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}
	return headers, cells, widths
}

func renderFlagTablePlain(rows []flagRow, indent string) string {
	if len(rows) == 0 {
		return ""
	}
	_, cells, widths := flagTableColumns(rows, false)

	var b strings.Builder
	for _, row := range cells {
		b.WriteString(indent + strings.TrimRight(padCells(row, widths, "  "), " ") + "\n")
	}
	return b.String()
}

func renderFlagTableMd(rows []flagRow, indent string) string {
	if len(rows) == 0 {
		return ""
	}
	headers, cells, widths := flagTableColumns(rows, true)

	var b strings.Builder
	b.WriteString(indent + "| " + padCells(headers, widths, " | ") + " |\n")

	seps := make([]string, len(widths))
	for i, w := range widths {
		seps[i] = strings.Repeat("-", w)
	}
	b.WriteString(indent + "| " + strings.Join(seps, " | ") + " |\n")

	for _, row := range cells {
		b.WriteString(indent + "| " + padCells(row, widths, " | ") + " |\n")
	}
	return b.String()
}

// padCells left-pads each cell to its column width and joins them with sep.
func padCells(cells []string, widths []int, sep string) string {
	parts := make([]string, len(cells))
	for i, c := range cells {
		w := 0
		if i < len(widths) {
			w = widths[i]
		}
		parts[i] = fmt.Sprintf("%-*s", w, c)
	}
	return strings.Join(parts, sep)
}

func anyRow(rows []flagRow, pred func(flagRow) bool) bool {
	for _, r := range rows {
		if pred(r) {
			return true
		}
	}
	return false
}

func flagColCell(r flagRow, markdown bool) string {
	if markdown {
		return flagCol(r)
	}
	return flagColPlain(r)
}

func flagColPlain(r flagRow) string {
	if r.Short != "" {
		return fmt.Sprintf("--%s, -%s", r.Flag, r.Short)
	}
	return fmt.Sprintf("--%s", r.Flag)
}

func envColCell(r flagRow, markdown bool) string {
	if markdown {
		return envColValue(r)
	}
	return r.Env
}

// valueColCell is the Value column for a row: the bare resolved value, wrapped in
// emphasis in the markdown path so the pretty renderer dims it.
func valueColCell(r flagRow, markdown bool) string {
	if r.Value == "" {
		return ""
	}
	if markdown {
		return "*" + r.Value + "*"
	}
	return r.Value
}

// globalSource returns the global flags struct to render the Global Options block
// from: the populated one when provided (so --help-values shows set values), else a
// zero struct (just the flag definitions).
func globalSource(globalValues *cli.GlobalFlags) any {
	if globalValues == nil {
		return &cli.GlobalFlags{}
	}
	return globalValues
}

func flagCol(r flagRow) string {
	if r.Short != "" {
		return fmt.Sprintf("`--%s`, `-%s`", r.Flag, r.Short)
	}
	return fmt.Sprintf("`--%s`", r.Flag)
}

// typeCol is the Type column for a row: just the type, plus a "required" marker when
// the flag is mandatory. The default value is not shown here (it reads as a trailing
// "(default: x)" hint on the description, see descCol).
func typeCol(r flagRow) string {
	t := r.Type
	if r.Required {
		t += ", required"
	}
	return t
}

// envColValue is the Env column for a row: just the variable name. The default value
// is not appended here (it lives in the description's "(default: x)" hint), so the
// column reads as the plain env var a user would export.
func envColValue(r flagRow) string {
	if r.Env == "" {
		return ""
	}
	return "`" + r.Env + "`"
}

// descCol is the Description column for a row: the help text with a trailing
// "(default: x)" hint when the field carries a non-zero default. markdown wraps the
// hint in emphasis so the pretty renderer dims it. Zero-value defaults (a bool
// "false", a numeric "0") are implied and suppressed to avoid noise (Cobra's rule).
func descCol(r flagRow, markdown bool) string {
	desc := r.Help
	def := defaultHint(r)
	if def == "" {
		return desc
	}
	if markdown {
		def = "*(default: " + r.Default + ")*"
	}
	if desc != "" {
		desc += " "
	}
	return desc + def
}

// defaultHint returns the "(default: x)" suffix for a row, or "" when the field has
// no default tag or its default equals the type's zero value (which is implied).
func defaultHint(r flagRow) string {
	if r.Default == "" || isZeroDefault(r.Type, r.Default) {
		return ""
	}
	return "(default: " + r.Default + ")"
}

// isZeroDefault reports whether def is the zero value for a field of the given type,
// so an explicit `default:"false"` on a bool (or `default:"0"` on a number) is not
// shown as a default the user needs to know about.
func isZeroDefault(typ, def string) bool {
	switch typ {
	case "bool":
		return def == "false"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return def == "0"
	case "float32", "float64":
		return def == "0" || def == "0.0"
	}
	return false
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
	rows := extractFlagRows(options, false)
	if len(rows) > 0 {
		writeAgentFlagRows(b, rows, indent, format)
	}
}

func writeGlobalFlagsBlock(b *strings.Builder, format string, extraFormats []string, globalValues *cli.GlobalFlags, showValues bool) {
	indent := "  "
	rows := extractFlagRowsWithFormats(globalSource(globalValues), extraFormats, showValues)
	if len(rows) == 0 {
		return
	}

	// allowed values for flags like --help-format ride along in the Help column as a
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
//
// The args are first tried as a single command path (e.g. ["db", "migrate"] narrows
// to db's migrate subcommand only, not all of db's subcommands). When they do not
// resolve to one path, each arg is treated as an independent name filter, so
// ["build", "deploy"] still lists both top-level commands.
func FilterCommands(commands []cli.Command[any], filters []string) []cli.Command[any] {
	if path := filterByPath(commands, filters); path != nil {
		return path
	}

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

// filterByPath interprets path as one command path (["db", "migrate"]) and returns
// the tree narrowed to exactly it: each ancestor rendered with only the matched
// child, the target shown with its own subcommands. Returns nil when path does not
// resolve to a command, so FilterCommands can fall back to independent name filters.
func filterByPath(commands []cli.Command[any], path []string) []cli.Command[any] {
	if len(path) == 0 {
		return nil
	}
	for _, cmd := range commands {
		if cmd.Name("") != path[0] {
			continue
		}
		if len(path) == 1 {
			return []cli.Command[any]{cmd}
		}
		subs := filterByPath(cmd.Commands(), path[1:])
		if subs == nil {
			return nil
		}
		return []cli.Command[any]{&filteredCommand{command: cmd, subs: subs}}
	}
	return nil
}

// filteredCommand wraps a parent command to expose only a subset of its subcommands.
// Used by filterCommands to narrow output without mutating the original command tree.
type filteredCommand struct {
	command cli.Command[any]
	subs    []cli.Command[any]
}

var _ cli.Command[any] = (*filteredCommand)(nil)

func (f *filteredCommand) Name(name string) string                     { return f.command.Name(name) }
func (f *filteredCommand) Add(name string, cmd cli.Command[any])       { f.command.Add(name, cmd) }
func (f *filteredCommand) Options() any                                { return f.command.Options() }
func (f *filteredCommand) Commands() []cli.Command[any]                { return f.subs }
func (f *filteredCommand) Run(o cli.GlobalFlags, u cli.Unknowns) error { return f.command.Run(o, u) }
func (f *filteredCommand) Validate(o map[string]any) error             { return f.command.Validate(o) }
func (f *filteredCommand) Help() string                                { return f.command.Help() }
func (f *filteredCommand) Description() string                         { return f.command.Description() }
func (f *filteredCommand) Examples() [][]string                        { return f.command.Examples() }
func (f *filteredCommand) Args() map[int][]string                      { return f.command.Args() }
func (f *filteredCommand) Flags() map[string][]string                  { return f.command.Flags() }
