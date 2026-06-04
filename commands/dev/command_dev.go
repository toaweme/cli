package dev

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/toaweme/cli"
)

// DevConfig holds the inputs for the dev command.
type DevConfig struct {
	OutputDir string `arg:"output" short:"o" help:"Output directory" default:"_data/outputs"`
}

// DevCommand runs every example binary and captures all possible help outputs.
type DevCommand struct {
	cli.BaseCommand[DevConfig]

	settingsFunc func() cli.Config
}

var _ cli.Command[DevConfig] = (*DevCommand)(nil)

// NewDevCommand creates a dev command for generating example outputs.
func NewDevCommand(settingsFunc func() cli.Config) *DevCommand {
	return &DevCommand{settingsFunc: settingsFunc}
}

func (c *DevCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	outputDir := c.Inputs.OutputDir
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	examples, err := discoverExamples()
	if err != nil {
		return fmt.Errorf("failed to discover examples: %w", err)
	}

	total := 0

	for _, example := range examples {
		cmds := discoverCommands(example)
		runs := buildRuns(example, cmds)

		for _, run := range runs {
			dir := filepath.Join(outputDir, example, run.dir)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "failed to create dir %s: %v\n", dir, err)
				continue
			}

			filename := run.name + ".md"
			path := filepath.Join(dir, filename)

			var content string
			if len(run.completeArgs) > 0 {
				content = formatCompleteOutputs(example, run)
			} else {
				output, exitCode := capture(example, run.args)
				content = formatOutput(example, run, output, exitCode)
			}

			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", path, err)
				continue
			}

			rel, _ := filepath.Rel(outputDir, path)
			total++
			fmt.Printf("  %s\n", rel)
		}
	}

	fmt.Printf("\n%d outputs written to %s/\n", total, outputDir)
	return nil
}

func (c *DevCommand) Help() string {
	return "Generate example outputs for all commands"
}

type run struct {
	dir          string
	name         string
	args         []string
	desc         string
	completeArgs [][]string
}

func discoverExamples() ([]string, error) {
	entries, err := os.ReadDir("examples")
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// discoverCommands runs --help --help-format=json to extract command names and subcommand paths.
func discoverCommands(example string) []string {
	out, _ := exec.Command("go", "run", "./examples/"+example, "--help", "--help-format=json").Output()
	if len(out) == 0 {
		return nil
	}

	type cmdJSON struct {
		Name        string    `json:"name"`
		SubCommands []cmdJSON `json:"subcommands"`
	}

	var commands []cmdJSON
	if err := json.Unmarshal(out, &commands); err != nil {
		return nil
	}

	skip := map[string]bool{"help": true, "version": true, "completion": true, "dev": true}
	var names []string
	for _, cmd := range commands {
		if skip[cmd.Name] {
			continue
		}
		names = append(names, cmd.Name)
		for _, sub := range cmd.SubCommands {
			names = append(names, cmd.Name+" "+sub.Name)
		}
	}
	return names
}

func cmdDir(parts []string) string {
	return filepath.Join(parts...)
}

func buildRuns(example string, cmds []string) []run {
	runs := []run{
		{name: "help", args: []string{"--help"}, desc: "default help output"},
		{name: "help-flags", args: []string{"--help", "--help-format=plain-flags"}, desc: "help with all flags expanded"},
		{name: "format-md", args: []string{"--help", "--help-format=md"}, desc: "comprehensive markdown docs"},
		{name: "format-plain", args: []string{"--help", "--help-format=plain"}, desc: "comprehensive plain text docs"},
		{name: "format-pretty", args: []string{"--help", "--help-format=pretty"}, desc: "comprehensive pretty docs (piped, no ANSI)"},
		{name: "format-json", args: []string{"--help", "--help-format=json"}, desc: "JSON command tree"},
		{name: "format-jsonschema", args: []string{"--help", "--help-format=jsonschema"}, desc: "JSON Schema for all commands"},
		{name: "version", args: []string{"--version"}, desc: "version output"},
	}

	for _, cmd := range cmds {
		parts := strings.Fields(cmd)
		dir := cmdDir(parts)
		helpArgs := append([]string{"help"}, parts...)
		runs = append(runs,
			run{dir: dir, name: "help", args: helpArgs, desc: fmt.Sprintf("help for %q", cmd)},
			run{dir: dir, name: "format-md", args: append(append([]string{}, helpArgs...), "--help-format=md"), desc: fmt.Sprintf("markdown docs for %q", cmd)},
			run{dir: dir, name: "format-plain", args: append(append([]string{}, helpArgs...), "--help-format=plain"), desc: fmt.Sprintf("plain docs for %q", cmd)},
			run{dir: dir, name: "format-json", args: append(append([]string{}, helpArgs...), "--help-format=json"), desc: fmt.Sprintf("JSON for %q", cmd)},
			run{dir: dir, name: "format-jsonschema", args: append(append([]string{}, helpArgs...), "--help-format=jsonschema"), desc: fmt.Sprintf("JSON Schema for %q", cmd)},
		)
	}

	var completionArgs [][]string
	for _, shell := range []string{"bash", "zsh", "fish"} {
		completionArgs = append(completionArgs, []string{"completion", shell})
	}
	runs = append(runs, run{
		name:         "completion",
		desc:         "shell completion scripts (bash, zsh, fish)",
		completeArgs: completionArgs,
	})

	// __complete outputs are collected into a single file per example
	var completeArgs [][]string
	completeArgs = append(completeArgs, []string{"__complete", ""})
	for _, cmd := range cmds {
		parts := strings.Fields(cmd)
		completeArgs = append(completeArgs, append(append([]string{"__complete"}, parts...), ""))
	}
	runs = append(runs, run{
		name:         "complete",
		args:         nil,
		desc:         "tab completion for all commands",
		completeArgs: completeArgs,
	})

	return runs
}

func capture(example string, args []string) (string, int) {
	cmd := exec.Command("go", append([]string{"run", "./examples/" + example}, args...)...)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return string(out), exitCode
}

func formatCompleteOutputs(example string, r run) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s: %s\n\n", example, r.desc))

	for _, args := range r.completeArgs {
		output, exitCode := capture(example, args)
		b.WriteString(fmt.Sprintf("```shell\n❯ %s %s\n```\n\n", example, strings.Join(args, " ")))
		if exitCode != 0 {
			b.WriteString(fmt.Sprintf("exit code: %d\n\n", exitCode))
		}
		b.WriteString("```\n")
		b.WriteString(output)
		if !strings.HasSuffix(output, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("```\n\n")
	}

	return b.String()
}

func formatOutput(example string, r run, output string, exitCode int) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s: %s\n\n", example, r.desc))
	b.WriteString(fmt.Sprintf("```shell\n❯ %s %s\n```\n\n", example, strings.Join(r.args, " ")))

	if exitCode != 0 {
		b.WriteString(fmt.Sprintf("exit code: %d\n\n", exitCode))
	}

	b.WriteString("```\n")
	b.WriteString(output)
	if !strings.HasSuffix(output, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("```\n")

	return b.String()
}
