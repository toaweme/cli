package cli

import (
	"cmp"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/toaweme/structs"
)

// tagArg and tagShort are the struct tags a field is matched against:
// arg holds the long flag name (or a numeric index for a positional arg),
// short holds the single-character alias.
const tagArg = "arg"
const tagShort = "short"

// optionPrefix marks an argument as a flag. Both "-x" and "--x" are accepted.
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

// getCommandArgs parses args against fields (the command's struct fields) and globals
// (the GlobalFlags fields). It returns, in order:
//
//   - parsedArgs: positional values whose position matched a numeric arg tag
//   - unknownArgs: positional values with no matching field
//   - parsedOptions: the command's flags, keyed by arg tag so long and short spellings share a key
//   - globalOptions: the global flags, keyed the same way
//   - the first flag the command line gets wrong. Parsing goes on past it so a later --help is still seen
//
// Supported syntax:
//
//   - "--key=value" / "-k=value": value taken from the right of the first "="
//   - "--key value" / "-k value": value taken from the next token unless it is a flag.
//     A negative number such as "-5" is a value, any other value starting with a dash needs "="
//   - bare "--flag": a bool field is set to true. A bool takes a value only with "="
//   - a slice field gathers every occurrence, a map field takes one "key=value" entry per occurrence
//
// Positional arguments are matched in the order they appear, skipping flags: the first
// real argument tries field "0", the second tries "1", and so on. Flags (and the values
// they consume) are not counted, so a positional still lands in the right slot no matter
// how many flags come before it (e.g. "--cwd /tmp deploy" makes "deploy" the first one).
func getCommandArgs(args []string, fields, globals []structs.Field) ([]string, []string, map[string]any, map[string]any, error) {
	parsedArgs := make([]string, 0)
	unknownArgs := make([]string, 0)
	parsedOptions := make(map[string]any)
	globalOptions := make(map[string]any)
	var problem error

	ordinal := 0
	for index := 0; index < len(args); index++ {
		arg := args[index]

		// positional: matched against a numeric arg tag by how many positionals came
		// before it (flags don't count), else unknown.
		if !isFlag(arg) {
			foundField := matchField(fields, strconv.Itoa(ordinal))
			ordinal++
			if foundField != nil {
				parsedArgs = append(parsedArgs, arg)
			} else {
				unknownArgs = append(unknownArgs, arg)
			}
			continue
		}

		flag, value, hasValue := strings.Cut(arg, "=")
		name := strings.TrimPrefix(strings.TrimPrefix(flag, optionPrefix), optionPrefix)
		options := parsedOptions
		field := matchFlag(fields, name)
		if field == nil {
			options, field = globalOptions, matchFlag(globals, name)
		}
		if field == nil {
			problem = cmp.Or(problem, fmt.Errorf("does not take flag %q: %w", flag, ErrUnknownFlag))
			continue
		}

		// a nested field keeps its qualified name, which is how the setter reaches it.
		key := name
		if field.FQN == nil {
			key = cmp.Or(field.Tags[tagArg], field.Tags[tagShort])
		}

		switch {
		case !hasValue && field.Kind == reflect.Bool:
			// a bool flag needs no value, bare presence means true.
			options[key] = true
			continue
		case !hasValue && index+1 < len(args) && isValue(args[index+1]):
			index++
			value = args[index]
		case !hasValue:
			problem = cmp.Or(problem, fmt.Errorf("flag %q needs a value: %w", flag, ErrInvalidFlagValue))
			continue
		}

		switch field.Kind {
		case reflect.Slice:
			// the setter still splits each occurrence on the field's sep tag.
			existing, _ := options[key].(structs.MultiValue)
			options[key] = append(existing, value)
		case reflect.Map:
			entryKey, entryValue, ok := strings.Cut(value, "=")
			if !ok || entryKey == "" {
				problem = cmp.Or(problem, fmt.Errorf("flag %q takes key=value, got %q: %w", flag, value, ErrInvalidFlagValue))
				continue
			}
			entries, _ := options[key].(map[string]string)
			if entries == nil {
				entries = map[string]string{}
				options[key] = entries
			}
			entries[entryKey] = entryValue
		default:
			options[key] = value
		}
	}

	return parsedArgs, unknownArgs, parsedOptions, globalOptions, problem
}

// matchFlag is matchField for a flag name. An empty or numeric name never matches,
// so "--" is never a field without a short tag and "--0" never fills a positional slot.
func matchFlag(fields []structs.Field, name string) *structs.Field {
	if _, err := strconv.Atoi(name); name == "" || err == nil {
		return nil
	}

	return matchField(fields, name)
}

// isFlag reports whether arg is written as a flag. A bare "-" is a positional, the stdin convention.
func isFlag(arg string) bool {
	return strings.HasPrefix(arg, optionPrefix) && arg != optionPrefix
}

// isValue reports whether arg can be a flag's value. A negative number such as "-5" can,
// while words ParseFloat accepts such as "-inf" stay flags.
func isValue(arg string) bool {
	if !isFlag(arg) {
		return true
	}
	_, err := strconv.ParseFloat(arg, 64)

	return err == nil && strings.ContainsAny(arg[1:2], ".0123456789")
}
