package config

import (
	"fmt"
	"strings"
)

// Type identifies which config a value is read from or written to: a machine-wide,
// per-user, or per-project file. The constants are conventional; an app may define
// its own since a Type is just a string.
type Type string

const (
	System  Type = "system"  // machine-wide, e.g. /etc/app
	Global  Type = "global"  // per-user, e.g. ~/.app
	Project Type = "project" // per-directory, e.g. ./
)

// Config is an ordered set of config files (lowest precedence first) plus an
// optional secrets backend. Register files with Add, then read the merged view with
// a Resolver (see NewFileResolver) or address a single file with From.
type Config struct {
	handlers []*handler
	secrets  SecretBackend
}

// New returns an empty Config. Register config files with Add.
func New() *Config {
	return &Config{}
}

// Add registers a config handler of the given Type, backed by store and read from name -
// the file's base name within the store, to which the store appends the codec
// extension (so "config" becomes e.g. config.yml). An empty name defaults to
// "config". Files are ordered lowest precedence first, so add the global config
// before the project one. Returns the Config for chaining.
func (c *Config) Add(configType Type, store Store, name string) *Config {
	if name == "" {
		name = "config"
	}
	c.handlers = append(c.handlers, &handler{configType: configType, store: store, name: name})
	return c
}

// WithSecrets attaches a secrets backend and returns the Config for chaining.
func (c *Config) WithSecrets(backend SecretBackend) *Config {
	c.secrets = backend
	return c
}

// From returns the config handler of the given Type. It errors if that Type was never
// registered, so a typo surfaces here rather than silently doing nothing.
func (c *Config) From(configType Type) (*handler, error) {
	for _, f := range c.handlers {
		if f.configType == configType {
			return f, nil
		}
	}
	return nil, fmt.Errorf("config %q is not registered", configType)
}

// Secret reads the secret at key into target via the configured backend.
func (c *Config) Secret(key string, target any) error {
	if c.secrets == nil {
		return fmt.Errorf("no secrets backend configured")
	}
	if err := c.secrets.Load(key, target); err != nil {
		return fmt.Errorf("failed to read secret %q: %w", key, err)
	}
	return nil
}

// handler is a single config file (selected by Type): read it whole, write it, or get
// and set a single dotted field. Obtain one from Config.From.
type handler struct {
	configType Type
	store      Store
	name       string
}

// Read decodes the file into target. A missing file is not an error.
func (f *handler) Read(target any) error {
	if !f.store.Exists(f.name) {
		return nil
	}
	if err := f.store.Load(f.name, target); err != nil {
		return fmt.Errorf("failed to read config %q: %w", f.name, err)
	}
	return nil
}

// Write persists cfg as the whole file.
func (f *handler) Write(cfg any) error {
	if err := f.store.Save(f.name, cfg); err != nil {
		return fmt.Errorf("failed to write config %q: %w", f.name, err)
	}
	return nil
}

// Set updates a single dotted path within the file (read-modify-write).
func (f *handler) Set(path string, value any) error {
	values := map[string]any{}
	if f.store.Exists(f.name) {
		if err := f.store.Load(f.name, &values); err != nil {
			return fmt.Errorf("failed to read config %q: %w", f.name, err)
		}
	}
	setPath(values, path, value)
	if err := f.store.Save(f.name, values); err != nil {
		return fmt.Errorf("failed to set %q in config %q: %w", path, f.name, err)
	}
	return nil
}

// Get returns the value at a dotted path, or nil when absent.
func (f *handler) Get(path string) (any, error) {
	if !f.store.Exists(f.name) {
		return nil, nil
	}
	values := map[string]any{}
	if err := f.store.Load(f.name, &values); err != nil {
		return nil, fmt.Errorf("failed to read config %q: %w", f.name, err)
	}
	v, _ := getPath(values, path)
	return v, nil
}

// setPath writes value at a dotted path within m, creating nested maps as needed.
func setPath(m map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	for _, p := range parts[:len(parts)-1] {
		next, ok := m[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[p] = next
		}
		m = next
	}
	m[parts[len(parts)-1]] = value
}

// getPath walks a dotted path through nested map[string]any values.
func getPath(m map[string]any, path string) (any, bool) {
	var current any = m
	for _, p := range strings.Split(path, ".") {
		mm, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = mm[p]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
