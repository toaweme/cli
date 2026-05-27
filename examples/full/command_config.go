package main

import (
	"fmt"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/config"
)

// AppConfig is persisted via config.FileStore to ~/.full/config.json
type AppConfig struct {
	DefaultOutput string `json:"default_output"`
	DefaultHost   string `json:"default_host"`
	DefaultPort   int    `json:"default_port"`
}

// ConfigShowConfig holds the inputs for the config show command.
type ConfigShowConfig struct{}

// ConfigShowCommand prints the current application config.
// Commands can hold dependencies like a store alongside BaseCommand.
type ConfigShowCommand struct {
	cli.BaseCommand[ConfigShowConfig]
	store *config.FileStore
}

var _ cli.Command[ConfigShowConfig] = (*ConfigShowCommand)(nil)

func (c *ConfigShowCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	var cfg AppConfig
	if err := c.store.Load("config", &cfg); err != nil {
		fmt.Println("no config found, using defaults")
		return nil
	}
	fmt.Printf("output=%s host=%s port=%d\n", cfg.DefaultOutput, cfg.DefaultHost, cfg.DefaultPort)
	return nil
}

func (c *ConfigShowCommand) Help() string {
	return "Show current config"
}

// ConfigSetConfig holds the inputs for the config set command.
type ConfigSetConfig struct {
	Output string `arg:"output" short:"o" help:"Default output directory"`
	Host   string `arg:"host" help:"Default host"`
	Port   int    `arg:"port" short:"p" help:"Default port"`
}

// ConfigSetCommand saves application config.
type ConfigSetCommand struct {
	cli.BaseCommand[ConfigSetConfig]
	store *config.FileStore
}

var _ cli.Command[ConfigSetConfig] = (*ConfigSetCommand)(nil)
var _ cli.ExampleProvider = (*ConfigSetCommand)(nil)

func (c *ConfigSetCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	cfg := AppConfig{
		DefaultOutput: c.Inputs.Output,
		DefaultHost:   c.Inputs.Host,
		DefaultPort:   c.Inputs.Port,
	}
	if err := c.store.Save("config", cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("config saved to %s\n", c.store.Dir())
	return nil
}

func (c *ConfigSetCommand) Help() string {
	return "Update config values"
}

func (c *ConfigSetCommand) Examples() []string {
	return []string{
		"full config set --output=./dist --host=0.0.0.0 --port=3000",
	}
}
