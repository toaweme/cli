package main

import (
	"fmt"

	"github.com/toaweme/cli"
)

type DBMigrateConfig struct {
	Steps  int  `arg:"steps" short:"n" help:"Number of migrations to run"`
	DryRun bool `arg:"dry-run" help:"Print SQL without executing"`
}

type DBMigrateCommand struct {
	cli.BaseCommand[DBMigrateConfig]
}

var _ cli.Command[DBMigrateConfig] = (*DBMigrateCommand)(nil)
var _ cli.ExampleProvider = (*DBMigrateCommand)(nil)

func (c *DBMigrateCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf("migrating steps=%d dry-run=%v\n", c.Inputs.Steps, c.Inputs.DryRun)
	return nil
}

func (c *DBMigrateCommand) Help() string {
	return "Run database migrations"
}

func (c *DBMigrateCommand) Examples() []string {
	return []string{
		"full db migrate",
		"full db migrate -n 3",
		"full db migrate --dry-run",
	}
}

type DBSeedConfig struct {
	File  string `arg:"file" short:"f" help:"Seed file path" rules:"required"`
	Force bool   `arg:"force" help:"Overwrite existing data"`
}

type DBSeedCommand struct {
	cli.BaseCommand[DBSeedConfig]
}

var _ cli.Command[DBSeedConfig] = (*DBSeedCommand)(nil)

func (c *DBSeedCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf("seeding file=%s force=%v\n", c.Inputs.File, c.Inputs.Force)
	return nil
}

func (c *DBSeedCommand) Help() string {
	return "Seed the database with test data"
}

type DBResetConfig struct {
	Confirm bool `arg:"confirm" short:"y" help:"Skip confirmation prompt"`
}

type DBResetCommand struct {
	cli.BaseCommand[DBResetConfig]
}

var _ cli.Command[DBResetConfig] = (*DBResetCommand)(nil)

func (c *DBResetCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf("resetting confirm=%v\n", c.Inputs.Confirm)
	return nil
}

func (c *DBResetCommand) Help() string {
	return "Drop and recreate the database"
}
