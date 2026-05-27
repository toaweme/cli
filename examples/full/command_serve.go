package main

import (
	"fmt"

	"github.com/toaweme/cli"
)

type ServeConfig struct {
	Port   int    `arg:"port" short:"p" env:"SERVER_PORT" help:"Port to listen on" default:"8080"`
	Host   string `arg:"host" env:"SERVER_HOST" help:"Host to bind to" default:"localhost"`
	TLS    bool   `arg:"tls" env:"SERVER_TLS" help:"Enable TLS"`
	Reload bool   `arg:"reload" short:"r" env:"SERVER_RELOAD" help:"Enable live reload"`
}

type ServeCommand struct {
	cli.BaseCommand[ServeConfig]
}

var _ cli.Command[ServeConfig] = (*ServeCommand)(nil)
var _ cli.ExampleProvider = (*ServeCommand)(nil)

func (c *ServeCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	fmt.Printf(
		"serving host=%s port=%d tls=%v reload=%v\n",
		c.Inputs.Host, c.Inputs.Port, c.Inputs.TLS, c.Inputs.Reload,
	)
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
