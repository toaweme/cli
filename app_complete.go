package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/toaweme/structs"
)

const (
	shellCompDirectiveNoFileComp = 4
)

func (c *app) handleComplete(args []string) {
	toComplete := ""
	if len(args) > 0 {
		toComplete = args[len(args)-1]
		args = args[:len(args)-1]
	}

	// walk args to find the deepest matching command
	commands := c.commands
	var matched Command[any]
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		found := c.matchCommandByName(arg, commands)
		if found == nil {
			break
		}
		matched = found
		commands = found.Commands()
	}

	if strings.HasPrefix(toComplete, "-") {
		prefix := strings.TrimLeft(toComplete, "-")
		c.completeFlagNames(matched, prefix)
	} else {
		for _, cmd := range commands {
			name := cmd.Name("")
			if strings.HasPrefix(name, toComplete) {
				fmt.Fprintf(os.Stdout, "%s\t%s\n", name, cmd.Help())
			}
		}
	}

	fmt.Fprintf(os.Stdout, ":%d\n", shellCompDirectiveNoFileComp)
}

func (c *app) completeFlagNames(cmd Command[any], prefix string) {
	seen := make(map[string]bool)

	if cmd != nil {
		c.completeFlagsFromOptions(cmd.Options(), prefix, seen)
	}
	c.completeFlagsFromOptions(c.globalFlags, prefix, seen)
}

func (c *app) completeFlagsFromOptions(options any, prefix string, seen map[string]bool) {
	if options == nil {
		return
	}

	fields, err := structs.GetStructFields(options, nil, structs.DefaultEncodingTags)
	if err != nil {
		return
	}

	for _, field := range fields {
		name := field.Tags["arg"]
		if name == "" {
			continue
		}
		if seen[name] {
			continue
		}
		if strings.HasPrefix(name, prefix) {
			seen[name] = true
			fmt.Fprintf(os.Stdout, "--%s\t%s\n", name, field.Tags["help"])
		}
	}
}
