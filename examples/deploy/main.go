package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/cmd/help"
	"github.com/toaweme/cli/cmd/version"
)

const appName = "deploy"
const appVersion = "0.1.0"

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	app := cli.NewApp(
		cli.Settings{Name: appName, Version: appVersion},
		cli.GlobalOptions{Cwd: cwd},
	)

	app.Add("help", help.NewHelpCommand(appName, app.Commands))
	app.Add("version", version.NewVersionCommand(appName, appVersion))

	parent := help.NewParentPlaceholder()
	app.Add("deploy", parent)
	parent.Add("staging", &DeployCommand{BaseCommand: cli.NewBaseCommand[DeployConfig]()})
	parent.Add("production", &DeployCommand{BaseCommand: cli.NewBaseCommand[DeployConfig]()})

	err = app.Run(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// DeployConfig holds the inputs for the deploy command.
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

func (c *DeployCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
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
