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

	settingsFunc func() cli.Settings
}

var _ cli.Command[VersionConfig] = (*VersionCommand)(nil)

// NewVersionCommand creates a version command that reads name and version from the app.
func NewVersionCommand(settingsFunc func() cli.Settings) *VersionCommand {
	return &VersionCommand{settingsFunc: settingsFunc}
}

func (c *VersionCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	s := c.settingsFunc()
	fmt.Printf("%s %s\n", s.Name, s.Version)
	return nil
}

func (c *VersionCommand) Validate(_ map[string]any) error {
	return nil
}

func (c *VersionCommand) Help() string {
	return "Display version"
}
