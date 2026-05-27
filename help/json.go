package help

import (
	"encoding/json"
	"fmt"

	"github.com/toaweme/cli"
	"github.com/toaweme/structs"
)

// CommandInfo is the JSON representation of a command.
type CommandInfo struct {
	Name        string        `json:"name"`
	Help        string        `json:"help"`
	Flags       []FlagInfo    `json:"flags,omitempty"`
	SubCommands []CommandInfo `json:"subcommands,omitempty"`
}

// FlagInfo is the JSON representation of a flag.
type FlagInfo struct {
	Name     string `json:"name"`
	Short    string `json:"short,omitempty"`
	Help     string `json:"help,omitempty"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
	Default  string `json:"default,omitempty"`
	Env      string `json:"env,omitempty"`
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
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
}

// DisplayHelpJSON outputs the command tree as a JSON array.
func DisplayHelpJSON(commands []cli.Command[any]) {
	info := buildCommandInfoList(commands)
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshal help JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// DisplayHelpJSONSchema outputs each command's options as a JSON Schema document.
func DisplayHelpJSONSchema(commands []cli.Command[any]) {
	schemas := buildSchemaList(commands)
	data, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshal help JSON schema: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func buildCommandInfoList(commands []cli.Command[any]) []CommandInfo {
	var result []CommandInfo
	for _, cmd := range commands {
		result = append(result, buildCommandInfo(cmd))
	}
	return result
}

func buildCommandInfo(cmd cli.Command[any]) CommandInfo {
	info := CommandInfo{
		Name:  cmd.Name(""),
		Help:  cmd.Help(),
		Flags: extractFlags(cmd.Options()),
	}

	for _, sub := range cmd.Commands() {
		info.SubCommands = append(info.SubCommands, buildCommandInfo(sub))
	}

	return info
}

func extractFlags(options any) []FlagInfo {
	if options == nil {
		return nil
	}

	fields, err := structs.GetStructFields(options, nil)
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
		flags = append(flags, fi)
	}

	return flags
}

func buildSchemaList(commands []cli.Command[any]) []CommandSchema {
	var result []CommandSchema
	for _, cmd := range commands {
		result = append(result, buildSchema(cmd))
		for _, sub := range cmd.Commands() {
			schema := buildSchema(sub)
			schema.Name = cmd.Name("") + " " + schema.Name
			result = append(result, schema)
		}
	}
	return result
}

func buildSchema(cmd cli.Command[any]) CommandSchema {
	schema := CommandSchema{
		Name:       cmd.Name(""),
		Help:       cmd.Help(),
		Properties: make(map[string]SchemaField),
	}

	options := cmd.Options()
	if options == nil {
		return schema
	}

	fields, err := structs.GetStructFields(options, nil)
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
