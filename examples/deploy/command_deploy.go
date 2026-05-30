package main

import (
	"fmt"

	"github.com/toaweme/cli"
)

// DeployConfig demonstrates rules:"required" which rejects the command if the arg is missing.
type DeployConfig struct {
	Tag    string `arg:"0" env:"DEPLOY_TAG" help:"Image tag to deploy" rules:"required"`
	Force  bool   `arg:"force" short:"f" env:"DEPLOY_FORCE" help:"Skip confirmation"`
	DryRun bool   `arg:"dry-run" env:"DEPLOY_DRY_RUN" help:"Print what would happen without executing"`
}

// DeployCommand deploys an image tag to a target environment.
type DeployCommand struct {
	cli.BaseCommand[DeployConfig]
}

var _ cli.Command[DeployConfig] = (*DeployCommand)(nil)

func (c *DeployCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	prefix := ""
	if c.Inputs.DryRun {
		prefix = "[dry-run] "
	}

	fmt.Printf("%sdeploying tag=%s force=%v\n", prefix, c.Inputs.Tag, c.Inputs.Force)
	return nil
}

func (c *DeployCommand) Help() string {
	return "Deploy an image tag"
}
