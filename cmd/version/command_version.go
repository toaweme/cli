package version

import (
	"fmt"

	"github.com/toaweme/cli"
)

// VersionConfig holds the inputs for the version command.
type VersionConfig struct{}

// VersionCommand prints the application name and version.
type VersionCommand struct {
	cli.BaseCommand[VersionConfig]

	name    string
	version string
}

var _ cli.Command[VersionConfig] = (*VersionCommand)(nil)

// NewVersionCommand creates a version command that prints "name version".
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
