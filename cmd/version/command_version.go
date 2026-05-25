package version

import (
	"fmt"

	"github.com/toaweme/cli"
)

type VersionVars struct{}

type VersionCommand struct {
	cli.BaseCommand[VersionVars]

	name    string
	version string
}

var _ cli.Command[VersionVars] = (*VersionCommand)(nil)

func NewVersionCommand(name, version string) *VersionCommand {
	return &VersionCommand{name: name, version: version}
}

func (c *VersionCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf("%s %s\n", c.name, c.version)
	return nil
}

func (c *VersionCommand) Validate(_ map[string]any) error {
	return nil
}

func (c *VersionCommand) Help() string {
	return "Display version"
}
