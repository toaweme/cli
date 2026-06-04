package help

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/toaweme/cli"
	"github.com/toaweme/structs"
)

// CommandInfo is the serialized representation of a command. The yaml/toml tags
// let the same struct render cleanly through the yaml/toml output codecs without
// this package importing those libraries. ArgDocs is keyed by the positional index
// as a string ("0", "1") rather than an int so it round-trips through codecs like
// toml, whose table keys must be strings.
type CommandInfo struct {
	Name        string              `json:"name" yaml:"name" toml:"name"`
	Help        string              `json:"help" yaml:"help" toml:"help"`
	Description string              `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Flags       []FlagInfo          `json:"flags,omitempty" yaml:"flags,omitempty" toml:"flags,omitempty"`
	Examples    [][]string          `json:"examples,omitempty" yaml:"examples,omitempty" toml:"examples,omitempty"`
	ArgDocs     map[string][]string `json:"argDescriptions,omitempty" yaml:"argDescriptions,omitempty" toml:"argDescriptions,omitempty"`
	FlagDocs    map[string][]string `json:"flagDescriptions,omitempty" yaml:"flagDescriptions,omitempty" toml:"flagDescriptions,omitempty"`
	SubCommands []CommandInfo       `json:"subcommands,omitempty" yaml:"subcommands,omitempty" toml:"subcommands,omitempty"`
}

// FlagInfo is the serialized representation of a flag.
type FlagInfo struct {
	Name     string `json:"name" yaml:"name" toml:"name"`
	Short    string `json:"short,omitempty" yaml:"short,omitempty" toml:"short,omitempty"`
	Help     string `json:"help,omitempty" yaml:"help,omitempty" toml:"help,omitempty"`
	Type     string `json:"type" yaml:"type" toml:"type"`
	Required bool   `json:"required,omitempty" yaml:"required,omitempty" toml:"required,omitempty"`
	Default  string `json:"default,omitempty" yaml:"default,omitempty" toml:"default,omitempty"`
	Env      string `json:"env,omitempty" yaml:"env,omitempty" toml:"env,omitempty"`
	// Value is the flag's resolved value, populated only under --help-values (secret
	// fields masked). Omitted otherwise so normal help output is unchanged.
	Value string `json:"value,omitempty" yaml:"value,omitempty" toml:"value,omitempty"`
}

// commandsDoc wraps the command list so codecs that cannot encode a top-level
// array (e.g. toml) have a table to anchor to. JSON still emits a bare array via
// DisplayHelpJSON; the keyed wrapper is only used by the generic codec path.
type commandsDoc struct {
	Commands []CommandInfo `json:"commands" yaml:"commands" toml:"commands"`
}

// CommandSchema is the JSON Schema representation of a command's options.
type CommandSchema struct {
	Name       string                 `json:"name"`
	Help       string                 `json:"help"`
	Properties map[string]SchemaField `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// SchemaField is a single field in a JSON Schema.
type SchemaField struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Default     string   `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	// Value is the field's resolved value, populated only under --help-values (secret
	// fields masked). Omitted otherwise so the schema is unchanged in normal use.
	Value string `json:"value,omitempty"`
}

// DisplayHelpJSON writes the command tree as a JSON array to w. Pass showValues to
// include each flag's resolved value (the --help-values mode).
func DisplayHelpJSON(w io.Writer, commands []cli.Command[any], showValues ...bool) {
	info := buildCommandInfoList(commands, optBool(showValues))
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		fmt.Fprintf(w, "failed to marshal help JSON: %v\n", err)
		return
	}
	fmt.Fprintln(w, string(data))
}

// DisplayHelpEncoded renders the command tree through a registered output codec
// (yaml, toml, ...) to w, reusing the same CommandInfo data DisplayHelpJSON builds.
// The list is wrapped in a `commands` table so codecs that reject a top-level array
// (toml) still encode.
func DisplayHelpEncoded(w io.Writer, commands []cli.Command[any], codec cli.OutputCodec, showValues ...bool) error {
	doc := commandsDoc{Commands: buildCommandInfoList(commands, optBool(showValues))}
	data, err := codec.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal help output as %q: %w", formatName(codec), err)
	}
	fmt.Fprintln(w, string(data))
	return nil
}

// formatName is the --help-format value for a codec, derived from its extension
// (".yml" -> "yml").
func formatName(codec cli.OutputCodec) string {
	return strings.TrimPrefix(codec.Extension(), ".")
}

// stringKeyedArgDocs converts the int-indexed positional arg docs into string keys
// ("0", "1") so the result encodes through codecs that require string map keys.
func stringKeyedArgDocs(docs map[int][]string) map[string][]string {
	if len(docs) == 0 {
		return nil
	}
	out := make(map[string][]string, len(docs))
	for idx, lines := range docs {
		out[strconv.Itoa(idx)] = lines
	}
	return out
}

// DisplayHelpJSONSchema writes each command's options as a JSON Schema document to w.
// Pass showValues to include each field's resolved value (the --help-values mode).
func DisplayHelpJSONSchema(w io.Writer, commands []cli.Command[any], showValues ...bool) {
	schemas := buildSchemaList(commands, optBool(showValues))
	data, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		fmt.Fprintf(w, "failed to marshal help JSON schema: %v\n", err)
		return
	}
	fmt.Fprintln(w, string(data))
}

// optBool reads the trailing variadic showValues argument the Display* functions
// accept, defaulting to false so callers that do not pass it keep normal help output.
func optBool(opts []bool) bool {
	return len(opts) > 0 && opts[0]
}

func buildCommandInfoList(commands []cli.Command[any], showValues bool) []CommandInfo {
	var result []CommandInfo
	for _, cmd := range commands {
		result = append(result, buildCommandInfo(cmd, showValues))
	}
	return result
}

func buildCommandInfo(cmd cli.Command[any], showValues bool) CommandInfo {
	info := CommandInfo{
		Name:        cmd.Name(""),
		Help:        cmd.Help(),
		Description: commandDescription(cmd),
		Flags:       extractFlags(cmd.Options(), showValues),
		Examples:    cmd.Examples(),
		ArgDocs:     stringKeyedArgDocs(cmd.Args()),
		FlagDocs:    cmd.Flags(),
	}

	for _, sub := range cmd.Commands() {
		info.SubCommands = append(info.SubCommands, buildCommandInfo(sub, showValues))
	}

	return info
}

func extractFlags(options any, showValues bool) []FlagInfo {
	if options == nil {
		return nil
	}

	fields, err := structs.GetStructFields(options, nil, structs.DefaultEncodingTags)
	if err != nil {
		return nil
	}

	var flags []FlagInfo
	for _, field := range fields {
		if field.Tags["arg"] == "" && field.Tags["short"] == "" {
			continue
		}
		if isPositionalArg(field.Tags["arg"]) {
			continue
		}

		fi := FlagInfo{
			Name:    field.Tags["arg"],
			Short:   field.Tags["short"],
			Help:    field.Tags["help"],
			Type:    field.Type,
			Default: field.Tags["default"],
			Env:     field.Tags["env"],
		}
		if hasRule(field, "required") {
			fi.Required = true
		}
		if showValues {
			fi.Value = fieldRawValue(field)
		}
		flags = append(flags, fi)
	}

	return flags
}

func buildSchemaList(commands []cli.Command[any], showValues bool) []CommandSchema {
	var result []CommandSchema
	for _, cmd := range commands {
		result = append(result, buildSchema(cmd, showValues))
		for _, sub := range cmd.Commands() {
			schema := buildSchema(sub, showValues)
			schema.Name = cmd.Name("") + " " + schema.Name
			result = append(result, schema)
		}
	}
	return result
}

func buildSchema(cmd cli.Command[any], showValues bool) CommandSchema {
	schema := CommandSchema{
		Name:       cmd.Name(""),
		Help:       cmd.Help(),
		Properties: make(map[string]SchemaField),
	}

	options := cmd.Options()
	if options == nil {
		return schema
	}

	fields, err := structs.GetStructFields(options, nil, structs.DefaultEncodingTags)
	if err != nil {
		return schema
	}

	for _, field := range fields {
		argName := field.Tags["arg"]
		if argName == "" || isPositionalArg(argName) {
			continue
		}

		sf := SchemaField{
			Type:        goTypeToSchemaType(field.Type),
			Description: field.Tags["help"],
			Default:     field.Tags["default"],
			Enum:        oneOfValues(field),
		}
		if showValues {
			sf.Value = fieldRawValue(field)
		}
		schema.Properties[argName] = sf

		if hasRule(field, "required") {
			schema.Required = append(schema.Required, argName)
		}
	}

	return schema
}

func goTypeToSchemaType(goType string) string {
	switch goType {
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return "number"
	default:
		return "string"
	}
}
