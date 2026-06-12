// Package completion provides a command that generates shell completion scripts.
package completion

import (
	"fmt"
	"strings"

	"github.com/toaweme/cli"
)

// Config holds the inputs for the completion command.
type Config struct {
	// Shell is the shell type to generate completions for.
	Shell string `arg:"0" help:"Shell type: bash, zsh, fish" rules:"required"`
}

// Command generates shell completion scripts.
type Command struct {
	cli.BaseCommand[Config]

	appName string
}

var _ cli.Command[Config] = (*Command)(nil)

// NewCompletionCommand creates a completion command for the given app name.
func NewCompletionCommand(appName string) *Command {
	return &Command{appName: appName}
}

// Run writes the completion script for the requested shell to stdout.
func (c *Command) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	shell := ""
	if c.Inputs != nil {
		shell = c.Inputs.Shell
	}

	var filename string
	switch strings.ToLower(shell) {
	case "bash":
		filename = "scripts/bash.sh"
	case "zsh":
		filename = "scripts/zsh.sh"
	case "fish":
		filename = "scripts/fish.sh"
	default:
		return fmt.Errorf("unsupported shell %q, supported: bash, zsh, fish", shell)
	}

	data, err := scripts.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read completion script for %s: %w", shell, err)
	}

	output := strings.ReplaceAll(string(data), "{{.AppName}}", c.appName)
	fmt.Print(output)
	return nil
}

// Help returns the one-line help summary for the command.
func (c *Command) Help() string {
	return "Generate shell completion scripts"
}

// Description returns the long-form description shown in help output.
func (c *Command) Description() string {
	return strings.Join([]string{
		"Generate a shell completion script and print it to stdout.",
		"",
		"Install (pick your shell):",
		"  bash:  " + c.appName + " completion bash > /etc/bash_completion.d/" + c.appName,
		"  zsh:   " + c.appName + ` completion zsh > "${fpath[1]}/_` + c.appName + `"`,
		"  fish:  " + c.appName + " completion fish > ~/.config/fish/completions/" + c.appName + ".fish",
		"",
		"Then restart your shell or source the file to enable completions.",
	}, "\n")
}

// Examples returns example invocations shown in help output.
func (c *Command) Examples() [][]string {
	return [][]string{
		{c.appName + ` completion bash > /etc/bash_completion.d/` + c.appName},
		{c.appName + ` completion zsh > "${fpath[1]}/_` + c.appName + `"`},
		{c.appName + ` completion fish > ~/.config/fish/completions/` + c.appName + `.fish`},
	}
}

// Args attaches a multi-line description to the positional shell argument.
func (c *Command) Args() map[int][]string {
	return map[int][]string{
		0: {
			"The shell to generate a completion script for.",
			"One of: bash, zsh, fish.",
		},
	}
}
