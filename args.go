package cli

import (
	"fmt"
	"strings"

	"github.com/toaweme/structs"
)

// tagArg and tagShort are the struct tags a field is matched against: arg holds
// the long flag name (or a numeric index for a positional arg), short holds the
// single-character alias.
const tagArg = "arg"
const tagShort = "short"

// optionPrefix marks an argument as a flag. Both "-x" and "--x" are accepted;
// any run of leading dashes is trimmed before matching.
const optionPrefix = "-"

// matchField returns the field whose arg or short tag equals name, or nil when
// no field claims that name. name is the bare flag (no dashes) or, for a
// positional argument, its index rendered as a string ("0", "1", ...).
//
// Nested struct fields are matched by their fully-qualified tag, built by gluing
// parent tags with "." (e.g. a `database.Connection` field tagged `arg:"database"`
// whose host field is tagged `arg:"host"` is reached as "database.host"). This is
// what lets `--database.host` target a nested config field.
func matchField(fields []structs.Field, name string) *structs.Field {
	for i := range fields {
		field := fields[i]
		if field.Tags[tagArg] == name || field.Tags[tagShort] == name {
			return &field
		}
		if found := matchNestedField(field.Fields, name); found != nil {
			return found
		}
	}

	return nil
}

// matchNestedField matches name against the fully-qualified tags of each nested
// field, recursing through deeper nesting. Unlike top-level matching it consults
// the full tag priority (arg, short, json, yaml), not just arg/short: a shared
// config type embedded as a field (e.g. database.Connection) typically carries
// only json/env tags, so "--database.host" must resolve through the json tag.
func matchNestedField(fields []structs.Field, name string) *structs.Field {
	if name == "" {
		return nil
	}
	for i := range fields {
		field := fields[i]
		if field.FQN != nil {
			for _, tag := range defaultTags {
				if field.FQN.Tags[tag] == name {
					return &field
				}
			}
		}
		if found := matchNestedField(field.Fields, name); found != nil {
			return found
		}
	}

	return nil
}

// getCommandArgs splits args into the four buckets a command needs, matching
// each token against fields (the command's struct fields). It returns, in order:
//
//   - parsedArgs: positional values whose index matched a numeric arg tag
//   - unknownArgs: positional values with no matching field (pass-through)
//   - parsedOptions: flags matched to a field, keyed by the name as written
//   - unknownOptions: flags with no matching field (pass-through)
//
// Supported syntax:
//
//   - "--key=value" / "-key=value": value taken from the right of the "="
//   - "-key value" / "--key value": value taken from the next token, which is
//     then consumed (skipped)
//   - bare "--flag": a bool field is set to true; an unknown bare flag with a
//     following value consumes it, otherwise it is recorded as true
//
// Positional arguments are matched by their index within args against numeric
// arg tags: the token at args[0] tries field "0", args[1] tries "1", and so on.
// Flags occupy indices too, so a positional value is found at the index it sits
// at in args, not at its position among non-flag tokens only.
func getCommandArgs(args []string, fields []structs.Field) ([]string, []string, map[string]any, map[string]any) {
	if len(args) < 1 {
		return []string{}, []string{}, map[string]any{}, map[string]any{}
	}

	parsedArgs := make([]string, 0)
	unknownArgs := make([]string, 0)
	parsedOptions := make(map[string]any)
	unknownOptions := make(map[string]any)

	for index := 0; index < len(args); index++ {
		arg := args[index]

		// positional: matched by index against a numeric arg tag, else unknown.
		if !strings.HasPrefix(arg, optionPrefix) {
			foundField := matchField(fields, fmt.Sprintf("%d", index))
			if foundField != nil {
				parsedArgs = append(parsedArgs, arg)
			} else {
				unknownArgs = append(unknownArgs, arg)
			}
			continue
		}

		dePrefixedArg := strings.TrimLeft(arg, optionPrefix)

		// "--key=value": value is on the right of the "=".
		if strings.Contains(dePrefixedArg, "=") {
			optName, optValue := splitKeyValue(dePrefixedArg)
			foundField := matchField(fields, optName)
			if foundField != nil {
				parsedOptions[optName] = optValue
			} else {
				unknownOptions[optName] = optValue
			}
			continue
		}

		foundField := matchField(fields, dePrefixedArg)
		if foundField != nil {
			// a bool flag needs no value: bare presence means true.
			if foundField.Type == "bool" {
				parsedOptions[dePrefixedArg] = true
				continue
			}

			// otherwise the value is the next token, which is then consumed.
			nextArg := ""
			if len(args) > index+1 {
				nextArg = args[index+1]
			}
			parsedOptions[dePrefixedArg] = nextArg
			index++
			continue
		}

		// unknown flag: take the next token as its value when present (consuming
		// it), otherwise record it as a bare boolean true.
		if len(args) > index+1 {
			unknownOptions[dePrefixedArg] = args[index+1]
			index++
			continue
		}
		unknownOptions[dePrefixedArg] = true
	}

	return parsedArgs, unknownArgs, parsedOptions, unknownOptions
}

// splitKeyValue splits a de-prefixed "key=value" token into its name and value.
// A token with no "=" yields an empty value; the split keeps only the first "="
// so values may themselves contain "=".
func splitKeyValue(arg string) (string, string) {
	pair := strings.SplitN(arg, "=", 2)

	optionName := pair[0]
	optionValue := ""
	if len(pair) > 1 {
		optionValue = pair[1]
	}

	return optionName, optionValue
}
