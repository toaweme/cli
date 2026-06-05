package main

import (
	"fmt"

	"github.com/toaweme/cli"
)

// BuildConfig demonstrates the default:"./bin" tag which sets the value when the flag is omitted entirely.
type BuildConfig struct {
	Output  string   `arg:"output" short:"o" env:"BUILD_OUTPUT" help:"Output directory" default:"./bin"`
	Tags    []string `arg:"tags" short:"t" env:"BUILD_TAGS" help:"Build tags (comma separated)"`
	Race    bool     `arg:"race" env:"BUILD_RACE" help:"Enable race detector"`
	Verbose bool     `arg:"verbose" env:"BUILD_VERBOSE" help:"Verbose build output"`
}

type BuildCommand struct {
	cli.BaseCommand[BuildConfig]
}

var _ cli.Command[BuildConfig] = (*BuildCommand)(nil)

func (c *BuildCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	fmt.Printf(
		"building output=%s tags=%v race=%v verbose=%v\n",
		c.Inputs.Output, c.Inputs.Tags, c.Inputs.Race, c.Inputs.Verbose,
	)
	return nil
}

func (c *BuildCommand) Help() string {
	return "Build the project"
}

// Examples are shown in --help --help-format=md|pretty|plain output. Each example is a slice of lines:
// the first is the command, the rest are sample output.
func (c *BuildCommand) Examples() [][]string {
	return [][]string{
		{"full build"},
		{"full build -o ./dist --race"},
		{
			"full build --tags=integration,e2e --verbose",
			"building output=./bin tags=[integration e2e] race=false verbose=true",
		},
	}
}

// Flags attaches multi-line descriptions to specific flags, keyed by the flag as it is written.
// These augment the single-line `help:` struct tags.
func (c *BuildCommand) Flags() map[string][]string {
	return map[string][]string{
		"--tags, -t": {
			"Comma-separated list of build tags, e.g. --tags=integration,e2e.",
			"Splits on commas into a []string; override the separator with the",
			"`sep:` struct tag.",
		},
	}
}
