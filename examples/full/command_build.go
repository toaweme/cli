package main

import (
	"fmt"

	"github.com/toaweme/cli"
)

// BuildConfig demonstrates the default:"./bin" tag which sets the value when the flag is omitted entirely.
type BuildConfig struct {
	Output  string `arg:"output" short:"o" env:"BUILD_OUTPUT" help:"Output directory" default:"./bin"`
	Tags    string `arg:"tags" short:"t" env:"BUILD_TAGS" help:"Build tags (comma separated)"`
	Race    bool   `arg:"race" env:"BUILD_RACE" help:"Enable race detector"`
	Verbose bool   `arg:"verbose" env:"BUILD_VERBOSE" help:"Verbose build output"`
}

type BuildCommand struct {
	cli.BaseCommand[BuildConfig]
}

var _ cli.Command[BuildConfig] = (*BuildCommand)(nil)

// ExampleProvider is optional; commands without it get auto-generated examples
var _ cli.ExampleProvider = (*BuildCommand)(nil)

func (c *BuildCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf(
		"building output=%s tags=%s race=%v verbose=%v\n",
		c.Inputs.Output, c.Inputs.Tags, c.Inputs.Race, c.Inputs.Verbose,
	)
	return nil
}

func (c *BuildCommand) Help() string {
	return "Build the project"
}

// Examples are shown in --help --format=md|pretty|plain output.
func (c *BuildCommand) Examples() []string {
	return []string{
		"full build",
		"full build -o ./dist --race",
		"full build --tags=integration,e2e --verbose",
	}
}
