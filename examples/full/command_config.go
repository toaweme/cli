package main

import (
	"fmt"

	"github.com/toaweme/cli"
)

// AppConfig is persisted via config.FileStore to ~/.full/config.json
type AppConfig struct {
	DefaultOutput string `json:"default_output"`
	DefaultHost   string `json:"default_host"`
	DefaultPort   int    `json:"default_port"`
}

// ConfigShowConfig holds the inputs for the config show command.
type ConfigShowConfig struct{}

// ConfigShowCommand prints the current application config. The config is passed
// in explicitly via NewConfigShowCommand rather than injected by the framework.
type ConfigShowCommand struct {
	cli.BaseCommand[ConfigShowConfig]
	cfg cli.Storage
}

var _ cli.Command[ConfigShowConfig] = (*ConfigShowCommand)(nil)

// NewConfigShowCommand builds the command with the config it reads from.
func NewConfigShowCommand(cfg cli.Storage) *ConfigShowCommand {
	return &ConfigShowCommand{BaseCommand: cli.NewBaseCommand[ConfigShowConfig](), cfg: cfg}
}

func (c *ConfigShowCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	var cfg AppConfig
	if err := c.cfg.Load("config", &cfg); err != nil {
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
	cfg cli.Storage
}

var _ cli.Command[ConfigSetConfig] = (*ConfigSetCommand)(nil)

// NewConfigSetCommand builds the command with the config it writes to.
func NewConfigSetCommand(cfg cli.Storage) *ConfigSetCommand {
	return &ConfigSetCommand{BaseCommand: cli.NewBaseCommand[ConfigSetConfig](), cfg: cfg}
}

func (c *ConfigSetCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	cfg := AppConfig{
		DefaultOutput: c.Inputs.Output,
		DefaultHost:   c.Inputs.Host,
		DefaultPort:   c.Inputs.Port,
	}
	if err := c.cfg.Save("config", cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("config saved to %s\n", c.cfg.Dir())
	return nil
}

func (c *ConfigSetCommand) Help() string {
	return "Update config values"
}

func (c *ConfigSetCommand) Examples() [][]string {
	return [][]string{
		{"full config set --output=./dist --host=0.0.0.0 --port=3000"},
	}
}
