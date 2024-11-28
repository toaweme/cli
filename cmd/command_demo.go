package main

import (
	"log/slog"

	"github.com/awee-ai/cli"
)

type Features struct {
	ModeName string `arg:"mode" short:"m" help:"Feature mode name" default:"Demo Basic Inc."`
	Billing  bool   `arg:"billing" short:"b" help:"Billing feature" default:"true"`
	Admin    bool   `arg:"admin" short:"a" help:"Admin feature" default:"false"`
	Users    bool   `arg:"users" short:"u" help:"Users feature" default:"true"`
	Preview  bool   `arg:"preview" short:"p" help:"Preview feature" default:"false"`
}

type DemoAppVars struct {
	Copyright string   `arg:"copy" help:"Copy right" default:"Demo Basic Inc. ©️2021-{{ year }}"`
	Features  Features `arg:"features" short:"f" help:"Features"`
}

type DemoAppCommand struct {
	cli.BaseCommand[DemoAppVars]
}

var _ cli.Command[DemoAppVars] = (*DemoAppCommand)(nil)

func NewDemoCommand() *DemoAppCommand {
	return &DemoAppCommand{
		BaseCommand: cli.NewBaseCommand[DemoAppVars](),
	}
}

func (c *DemoAppCommand) Run(options cli.GlobalOptions, unknowns cli.Unknowns) error {
	features := c.Inputs.Features
	if features.Billing {
		slog.Info("billing feature enabled")
	}
	if features.Admin {
		slog.Info("admin feature enabled")
	}
	if features.Users {
		slog.Info("users feature enabled")
	}
	if features.Preview {
		slog.Info("preview feature enabled")
	}
	// start http, grpc, ws, etc. servers
	// connect to databases
	// do whatever you need to do under "$ ./binary demo -f.b no -f.a yes -f.u yes -f.p no"
	slog.Info("running application", "copy", c.Inputs.Copyright, "mode", features.ModeName)

	return nil
}

func (c *DemoAppCommand) Validate(vars map[string]any) error {
	return nil
}

// Help returns the help message for the command
func (c *DemoAppCommand) Help() string {
	return "Demo app command"
}
