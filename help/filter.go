package help

import (
	"strings"

	"github.com/toaweme/cli"
)

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
				Command: cmd,
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
		return []cli.Command[any]{&filteredCommand{Command: cmd, subs: subs}}
	}
	return nil
}

// filteredCommand is a command that reports only a subset of its subcommands. It
// embeds the real command (delegating every method) and overrides Commands(), so
// FilterCommands can hand the renderers a narrowed view without mutating the shared
// command tree.
type filteredCommand struct {
	cli.Command[any]
	subs []cli.Command[any]
}

var _ cli.Command[any] = (*filteredCommand)(nil)

func (f *filteredCommand) Commands() []cli.Command[any] { return f.subs }
