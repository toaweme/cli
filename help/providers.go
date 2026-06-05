package help

import (
	"fmt"
	"sort"
	"strings"

	"github.com/toaweme/structs"

	"github.com/toaweme/cli"
)

// commandExamples returns usage examples for a command. The command's own Examples are used when present;
// otherwise a single example is auto-generated from the flag definitions. Each example is a slice of lines:
// line 0 is the invocation, the rest are sample output.
func commandExamples(cmd cli.Command[any], fullName, appName string) [][]string {
	if examples := cmd.Examples(); len(examples) > 0 {
		return examples
	}

	flags := extractExampleFlags(cmd.Options())
	if len(flags) == 0 {
		return nil
	}

	return [][]string{{appName + " " + fullName + flags}}
}

// extractExampleFlags builds the trailing arg/flag placeholders for an auto-generated usage example
// (e.g. " <name> --shout") from a command's option struct tags.
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

// docEntry is one labelled multi-line description rendered by docBlock.
type docEntry struct {
	label string
	lines []string
}

// providerDocLines renders the Arguments and Flag details blocks a command exposes via its Args/Flags methods,
// indented by indent. Returns nil when the command provides neither. Flag entries are sorted by their label
// so output is deterministic regardless of map iteration order.
func providerDocLines(cmd cli.Command[any], indent string) []string {
	var lines []string

	if docs := cmd.Args(); len(docs) > 0 {
		keys := make([]int, 0, len(docs))
		for k := range docs {
			keys = append(keys, k)
		}
		sort.Ints(keys)

		entries := make([]docEntry, 0, len(keys))
		for _, k := range keys {
			entries = append(entries, docEntry{label: fmt.Sprintf("[%d]", k), lines: docs[k]})
		}
		lines = append(lines, docBlock("Arguments", entries, indent)...)
	}

	if docs := cmd.Flags(); len(docs) > 0 {
		keys := make([]string, 0, len(docs))
		for k := range docs {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		entries := make([]docEntry, 0, len(keys))
		for _, k := range keys {
			entries = append(entries, docEntry{label: k, lines: docs[k]})
		}
		lines = append(lines, docBlock("Flag details", entries, indent)...)
	}

	return lines
}

// docBlock renders a titled block of labelled multi-line descriptions. The first line of each entry sits
// next to its label; continuation lines are aligned under it. Returns nil when entries is empty.
func docBlock(title string, entries []docEntry, indent string) []string {
	if len(entries) == 0 {
		return nil
	}

	labelW := 0
	for _, e := range entries {
		if len(e.label) > labelW {
			labelW = len(e.label)
		}
	}

	lines := []string{"", indent + title + ":"}
	cont := indent + "  " + strings.Repeat(" ", labelW+2)
	for _, e := range entries {
		first := ""
		if len(e.lines) > 0 {
			first = e.lines[0]
		}
		lines = append(lines, fmt.Sprintf("%s  %-*s  %s", indent, labelW, e.label, first))
		for _, line := range e.lines[1:] {
			lines = append(lines, cont+line)
		}
	}

	return lines
}
