package main

import (
	"errors"
	"fmt"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/config"
)

// AppConfig is persisted via the config store to ~/.full/config.json
type AppConfig struct {
	DefaultOutput string `json:"default_output"`
	DefaultHost   string `json:"default_host"`
	DefaultPort   int    `json:"default_port"`
}

// ConfigShowConfig holds the inputs for the config show command.
type ConfigShowConfig struct{}

// ConfigShowCommand prints the current application config.
// The store is passed in explicitly via NewConfigShowCommand rather than injected by the framework.
type ConfigShowCommand struct {
	cli.BaseCommand[ConfigShowConfig]
	store config.Store
}

var _ cli.Command[ConfigShowConfig] = (*ConfigShowCommand)(nil)

// NewConfigShowCommand builds the command with the store it reads from.
func NewConfigShowCommand(store config.Store) *ConfigShowCommand {
	return &ConfigShowCommand{BaseCommand: cli.NewBaseCommand[ConfigShowConfig](), store: store}
}

func (c *ConfigShowCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	// no config yet is fine: fall back to the zero value rather than erroring.
	var cfg AppConfig
	if err := c.store.Read(&cfg); err != nil && !errors.Is(err, config.ErrConfigNotFound) {
		return fmt.Errorf("failed to read config: %w", err)
	}
	fmt.Printf("output=%s host=%s port=%d\n", cfg.DefaultOutput, cfg.DefaultHost, cfg.DefaultPort)
	return nil
}

func (c *ConfigShowCommand) Help() string {
	return "Show current config"
}

// ConfigSetConfig holds the inputs for the config set command.
type ConfigSetConfig struct {
	Output string `arg:"output" short:"o" env:"CONFIG_OUTPUT" help:"Default output directory" default:"./bin"`
	Host   string `arg:"host" env:"CONFIG_HOST" help:"Default host" default:"localhost"`
	Port   int    `arg:"port" short:"p" env:"CONFIG_PORT" help:"Default port" default:"8080"`
}

// ConfigSetCommand saves application config.
type ConfigSetCommand struct {
	cli.BaseCommand[ConfigSetConfig]
	store config.Store
}

var _ cli.Command[ConfigSetConfig] = (*ConfigSetCommand)(nil)

// NewConfigSetCommand builds the command with the store it writes to.
func NewConfigSetCommand(store config.Store) *ConfigSetCommand {
	return &ConfigSetCommand{BaseCommand: cli.NewBaseCommand[ConfigSetConfig](), store: store}
}

func (c *ConfigSetCommand) Run(_ cli.GlobalFlags, _ cli.Unknowns) error {
	cfg := AppConfig{
		DefaultOutput: c.Inputs.Output,
		DefaultHost:   c.Inputs.Host,
		DefaultPort:   c.Inputs.Port,
	}
	if err := c.store.Write(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Println("config saved")
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
