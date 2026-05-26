package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/cmd/help"
	"github.com/toaweme/cli/cmd/version"
)

const appName = "full"
const appVersion = "0.1.0"

// BuildConfig holds the inputs for the build command.
type BuildConfig struct {
	Output  string `arg:"output" short:"o" env:"BUILD_OUTPUT" help:"Output directory" default:"./bin"`
	Tags    string `arg:"tags" short:"t" env:"BUILD_TAGS" help:"Build tags (comma separated)"`
	Race    bool   `arg:"race" env:"BUILD_RACE" help:"Enable race detector"`
	Verbose bool   `arg:"verbose" env:"BUILD_VERBOSE" help:"Verbose build output"`
}

// BuildCommand compiles the project with the given options.
type BuildCommand struct {
	cli.BaseCommand[BuildConfig]
}

var _ cli.Command[BuildConfig] = (*BuildCommand)(nil)
var _ cli.ExampleProvider = (*BuildCommand)(nil)

func (c *BuildCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf("building output=%s tags=%s race=%v verbose=%v\n",
		c.Inputs.Output, c.Inputs.Tags, c.Inputs.Race, c.Inputs.Verbose)
	return nil
}

func (c *BuildCommand) Help() string {
	return "Build the project"
}

func (c *BuildCommand) Examples() []string {
	return []string{
		"full build",
		"full build -o ./dist --race",
		"full build --tags=integration,e2e --verbose",
	}
}

// ServeConfig holds the inputs for the serve command.
type ServeConfig struct {
	Port   int    `arg:"port" short:"p" env:"SERVE_PORT" help:"Port to listen on" default:"8080"`
	Host   string `arg:"host" env:"SERVE_HOST" help:"Host to bind to" default:"localhost"`
	TLS    bool   `arg:"tls" env:"SERVE_TLS" help:"Enable TLS"`
	Reload bool   `arg:"reload" short:"r" env:"SERVE_RELOAD" help:"Enable live reload"`
}

// ServeCommand starts the development server.
type ServeCommand struct {
	cli.BaseCommand[ServeConfig]
}

var _ cli.Command[ServeConfig] = (*ServeCommand)(nil)
var _ cli.ExampleProvider = (*ServeCommand)(nil)

func (c *ServeCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf("serving host=%s port=%d tls=%v reload=%v\n",
		c.Inputs.Host, c.Inputs.Port, c.Inputs.TLS, c.Inputs.Reload)
	return nil
}

func (c *ServeCommand) Help() string {
	return "Start the development server"
}

func (c *ServeCommand) Examples() []string {
	return []string{
		"full serve",
		"full serve -p 3000 --reload",
		"full serve --host=0.0.0.0 --tls",
	}
}

// DBMigrateConfig holds the inputs for the db migrate command.
type DBMigrateConfig struct {
	Steps  int  `arg:"steps" short:"n" help:"Number of migrations to run"`
	DryRun bool `arg:"dry-run" help:"Print SQL without executing"`
}

// DBMigrateCommand runs database migrations.
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

// DBSeedConfig holds the inputs for the db seed command.
type DBSeedConfig struct {
	File  string `arg:"file" short:"f" help:"Seed file path" validate:"required"`
	Force bool   `arg:"force" help:"Overwrite existing data"`
}

// DBSeedCommand seeds the database with test data.
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

// DBResetConfig holds the inputs for the db reset command.
type DBResetConfig struct {
	Confirm bool `arg:"confirm" short:"y" help:"Skip confirmation prompt"`
}

// DBResetCommand drops and recreates the database.
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
	app.Add("build", &BuildCommand{BaseCommand: cli.NewBaseCommand[BuildConfig]()})
	app.Add("serve", &ServeCommand{BaseCommand: cli.NewBaseCommand[ServeConfig]()})

	db := help.NewParentPlaceholder()
	app.Add("db", db)
	db.Add("migrate", &DBMigrateCommand{BaseCommand: cli.NewBaseCommand[DBMigrateConfig]()})
	db.Add("seed", &DBSeedCommand{BaseCommand: cli.NewBaseCommand[DBSeedConfig]()})
	db.Add("reset", &DBResetCommand{BaseCommand: cli.NewBaseCommand[DBResetConfig]()})

	err = app.Run(os.Args[1:])
	if err != nil {
		if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
