package main

import (
	"fmt"

	"github.com/toaweme/cli"
)

// BuildConfig demonstrates default:, comma-separated slice, and bool flags. The
// default command runs when no args are given.
type BuildConfig struct {
	Output  string   `arg:"output" short:"o" env:"BUILD_OUTPUT" help:"Output directory" default:"./bin"`
	Tags    []string `arg:"tags" short:"t" env:"BUILD_TAGS" help:"Build tags (comma separated)"`
	Verbose bool     `arg:"verbose" env:"BUILD_VERBOSE" help:"Verbose build output"`
}

type BuildCommand struct {
	cli.BaseCommand[BuildConfig]
}

var _ cli.Command[BuildConfig] = (*BuildCommand)(nil)

func (c *BuildCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	fmt.Printf("building output=%s tags=%v verbose=%v\n", c.Inputs.Output, c.Inputs.Tags, c.Inputs.Verbose)
	return nil
}

func (c *BuildCommand) Help() string { return "Build the project" }

func (c *BuildCommand) Examples() [][]string {
	return [][]string{
		{"full3p build"},
		{"full3p build -o ./dist --tags=integration,e2e --verbose"},
	}
}

// ServeConfig is sourced from a "server:" config section via ConfigStrategy.
type ServeConfig struct {
	Port int    `arg:"port" short:"p" env:"SERVER_PORT" help:"Port to listen on" default:"8080"`
	Host string `arg:"host" env:"SERVER_HOST" help:"Host to bind to" default:"localhost"`
	TLS  bool   `arg:"tls" env:"SERVER_TLS" help:"Enable TLS"`
}

type ServeCommand struct {
	cli.BaseCommand[ServeConfig]
}

var _ cli.Command[ServeConfig] = (*ServeCommand)(nil)

func (c *ServeCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	fmt.Printf("serving host=%s port=%d tls=%v\n", c.Inputs.Host, c.Inputs.Port, c.Inputs.TLS)
	return nil
}

// ConfigStrategy keeps the app-wide merge default but sources this command's
// fields from a "server:" section in the config file.
func (c *ServeCommand) ConfigStrategy() (cli.MergeStrategy, cli.ConfigMapping) {
	return cli.MergeInherit, cli.Namespaced("server")
}

func (c *ServeCommand) Help() string { return "Start the server" }

// DBMigrateConfig is a leaf subcommand under the "db" parent placeholder.
type DBMigrateConfig struct {
	Steps int `arg:"steps" short:"n" help:"Number of migrations to apply" default:"1"`
}

type DBMigrateCommand struct {
	cli.BaseCommand[DBMigrateConfig]
}

var _ cli.Command[DBMigrateConfig] = (*DBMigrateCommand)(nil)

func (c *DBMigrateCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	fmt.Printf("migrating %d step(s)\n", c.Inputs.Steps)
	return nil
}

func (c *DBMigrateCommand) Help() string { return "Apply database migrations" }
