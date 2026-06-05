package gendocs

import (
	"fmt"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/help/gendocs"
)

// GenDocsConfig holds the inputs for the gendocs command.
type GenDocsConfig struct {
	OutputDir  string `arg:"out" short:"o" env:"GENDOCS_OUT" help:"Output directory" default:"docs"`
	PerCommand bool   `arg:"per-command" help:"Also write one file per command"`
}

// GenDocsCommand renders the application's own command tree to documentation files,
// in every help format the app supports. It introspects the live command tree in process,
// so the docs match the binary exactly and never go stale.
type GenDocsCommand struct {
	cli.BaseCommand[GenDocsConfig]

	settingsFunc    func() cli.Config
	commandListFunc func() []cli.Command[any]
	formatsFunc     func() []cli.OutputCodec
}

var _ cli.Command[GenDocsConfig] = (*GenDocsCommand)(nil)

// NewGenDocsCommand creates a gendocs command. It takes the same getters as the help command
// (App.Config, App.Commands, App.OutputFormats) so it can render the host app's command tree
// and custom formats without re-running the binary.
func NewGenDocsCommand(settingsFunc func() cli.Config, commandList func() []cli.Command[any], formats func() []cli.OutputCodec) *GenDocsCommand {
	return &GenDocsCommand{
		BaseCommand:     cli.NewBaseCommand[GenDocsConfig](),
		settingsFunc:    settingsFunc,
		commandListFunc: commandList,
		formatsFunc:     formats,
	}
}

func (c *GenDocsCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	written, err := gendocs.Generate(gendocs.Options{
		AppName:    c.settingsFunc().Name,
		Commands:   c.commandListFunc(),
		Codecs:     c.formatsFunc(),
		Dir:        c.Inputs.OutputDir,
		PerCommand: c.Inputs.PerCommand,
	})
	if err != nil {
		return fmt.Errorf("failed to generate docs: %w", err)
	}

	for _, path := range written {
		fmt.Printf("  %s\n", path)
	}
	fmt.Printf("\n%d files written to %s/\n", len(written), c.Inputs.OutputDir)
	return nil
}

func (c *GenDocsCommand) Help() string {
	return "Generate documentation for all commands"
}

func (c *GenDocsCommand) Validate(_ map[string]any) error {
	return nil
}
