package cli_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/toaweme/cli"
)

// helloConfig declares the hello command's flags and positional args via struct tags.
type helloConfig struct {
	Name  string `arg:"0" env:"HELLO_NAME" help:"Who to greet"`
	Shout bool   `arg:"shout" short:"s" help:"Uppercase the greeting"`
}

// helloCommand greets someone by name, optionally shouting.
type helloCommand struct {
	cli.BaseCommand[helloConfig]
}

var _ cli.Command[helloConfig] = (*helloCommand)(nil)

func (c *helloCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	name := c.Inputs.Name
	if name == "" {
		name = "world"
	}
	msg := fmt.Sprintf("hello, %s!", name)
	if c.Inputs.Shout {
		msg = strings.ToUpper(msg)
	}
	fmt.Println(msg)
	return nil
}

func (c *helloCommand) Help() string { return "Greet someone by name" }

// migrateConfig is empty: the migrate command takes no flags or args.
type migrateConfig struct{}

type migrateCommand struct {
	cli.BaseCommand[migrateConfig]
}

var _ cli.Command[migrateConfig] = (*migrateCommand)(nil)

func (c *migrateCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	fmt.Println("running migrations")
	return nil
}

func (c *migrateCommand) Help() string { return "Apply pending migrations" }

// dbCommand is a parent placeholder whose work lives in its subcommands.
type dbCommand struct {
	cli.BaseCommand[struct{}]
}

var _ cli.Command[struct{}] = (*dbCommand)(nil)

func (c *dbCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	// returning ErrDisplaySubCommands tells the framework to print the subcommand list
	return cli.ErrDisplaySubCommands
}

func (c *dbCommand) Help() string { return "Database commands" }

// sayConfig has a default, which the resolver and flags can override.
type sayConfig struct {
	Greeting string `arg:"greeting" default:"hi" help:"Greeting word"`
}

type sayCommand struct {
	cli.BaseCommand[sayConfig]
}

var _ cli.Command[sayConfig] = (*sayCommand)(nil)

func (c *sayCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	fmt.Println(c.Inputs.Greeting)
	return nil
}

func (c *sayCommand) Help() string { return "Say a greeting" }

// staticResolver overlays a fixed set of values, satisfying cli.Resolver structurally.
type staticResolver struct {
	values map[string]any
}

var _ cli.Resolver = staticResolver{}

func (r staticResolver) Resolve(_ string, values map[string]any) (map[string]any, error) {
	for k, v := range r.values {
		values[k] = v
	}
	return values, nil
}

// Example builds a one-command app and dispatches a positional argument to it.
func Example() {
	app := cli.NewApp(
		cli.Config{Name: "greet", Version: "1.0.0"},
		cli.GlobalFlags{},
	)
	app.Add("hello", &helloCommand{BaseCommand: cli.NewBaseCommand[helloConfig]()})

	if err := app.Run([]string{"hello", "Ada"}); cli.IsRealError(err) {
		fmt.Println("error:", err)
	}
	// Output: hello, Ada!
}

// ExampleNewApp shows the minimal app: identity, one command, dispatch.
func ExampleNewApp() {
	app := cli.NewApp(
		cli.Config{Name: "greet", Version: "1.0.0"},
		cli.GlobalFlags{},
	)
	app.Add("hello", &helloCommand{BaseCommand: cli.NewBaseCommand[helloConfig]()})

	_ = app.Run([]string{"hello", "Ada", "--shout"})
	// Output: HELLO, ADA!
}

// ExampleApp_Add builds a subcommand tree: `db migrate` runs the leaf command.
func ExampleApp_Add() {
	app := cli.NewApp(cli.Config{Name: "tool"}, cli.GlobalFlags{})

	db := app.Add("db", &dbCommand{BaseCommand: cli.NewBaseCommand[struct{}]()})
	db.Add("migrate", &migrateCommand{BaseCommand: cli.NewBaseCommand[migrateConfig]()})

	_ = app.Run([]string{"db", "migrate"})
	// Output: running migrations
}

// ExampleApp_Default registers a command to run when no arguments are given.
func ExampleApp_Default() {
	app := cli.NewApp(cli.Config{Name: "greet"}, cli.GlobalFlags{})

	// register the command, then mark the same instance as the default so bare invocation (no args) dispatches to it.
	hello := app.Add("hello", &helloCommand{BaseCommand: cli.NewBaseCommand[helloConfig]()})
	app.Default(hello)

	_ = app.Run(nil)
	// Output: hello, world!
}

// ExampleApp_Resolve overlays config values below env and flags.
// The resolver's "hola" overrides the struct default "hi".
func ExampleApp_Resolve() {
	app := cli.NewApp(cli.Config{Name: "demo"}, cli.GlobalFlags{})
	app.Resolve(staticResolver{values: map[string]any{"greeting": "hola"}})
	app.Add("say", &sayCommand{BaseCommand: cli.NewBaseCommand[sayConfig]()})

	_ = app.Run([]string{"say"})
	// Output: hola
}

// ExampleIsRealError filters the clean-exit sentinels from genuine failures.
func ExampleIsRealError() {
	fmt.Println(cli.IsRealError(nil))
	fmt.Println(cli.IsRealError(cli.ErrShowingHelp))
	fmt.Println(cli.IsRealError(cli.ErrShowingVersion))
	fmt.Println(cli.IsRealError(errors.New("boom")))
	// Output:
	// false
	// false
	// false
	// true
}
