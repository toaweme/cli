package completion

import (
	"fmt"
	"strings"

	"github.com/toaweme/cli"
)

// CompletionConfig holds the inputs for the completion command.
type CompletionConfig struct {
	// Shell is the shell type to generate completions for.
	Shell string `arg:"0" help:"Shell type: bash, zsh, fish" validate:"required"`
}

// CompletionCommand generates shell completion scripts.
type CompletionCommand struct {
	cli.BaseCommand[CompletionConfig]

	appName string
}

var _ cli.Command[CompletionConfig] = (*CompletionCommand)(nil)

// NewCompletionCommand creates a completion command for the given app name.
func NewCompletionCommand(appName string) *CompletionCommand {
	return &CompletionCommand{appName: appName}
}

func (c *CompletionCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	shell := ""
	if c.Inputs != nil {
		shell = c.Inputs.Shell
	}

	filename := ""
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

func (c *CompletionCommand) Help() string {
	return "Generate shell completion scripts"
}
