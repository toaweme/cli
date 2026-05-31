package config

import (
	"fmt"
	"strings"
)

// Type identifies a config scope: which store a value is read from or written to.
// The constants below are conventional; an app may define its own scopes since a
// Type is just a string.
type Type string

const (
	System  Type = "system"  // machine-wide, e.g. /etc/app
	Global  Type = "global"  // per-user, e.g. ~/.app
	Project Type = "project" // per-directory, e.g. ./
)

// scope binds a Type to the store backing it and the key (file base name) read
// within that store.
type scopeBinding struct {
	typ   Type
	store Store
	key   string
}

// Config is an ordered set of config scopes (lowest precedence first) plus an
// optional secrets backend. Declare scopes with Add, then read the merged view
// with a Resolver (see NewFileResolver) or address a single scope with Scope.
type Config struct {
	scopes  []scopeBinding
	secrets SecretBackend
}

// New returns an empty Config. Declare scopes with Add.
func New() *Config {
	return &Config{}
}

// Add registers a config scope of type t backed by store and read under key (an
// empty key defaults to "config"), and returns the Config for chaining. Scopes are
// ordered lowest precedence first, so add the global scope before the project one.
func (c *Config) Add(t Type, store Store, key string) *Config {
	if key == "" {
		key = "config"
	}
	c.scopes = append(c.scopes, scopeBinding{typ: t, store: store, key: key})
	return c
}

// WithSecrets attaches a secrets backend and returns the Config for chaining.
func (c *Config) WithSecrets(backend SecretBackend) *Config {
	c.secrets = backend
	return c
}

// Scope returns a handle to a single config scope for reading, seeding, or setting
// a field. Addressing an unregistered scope returns a handle whose methods error,
// so a typo surfaces at the call rather than silently doing nothing.
func (c *Config) Scope(t Type) Scope {
	for _, b := range c.scopes {
		if b.typ == t {
			return &scope{store: b.store, key: b.key}
		}
	}
	return missingScope{typ: t}
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

// Scope is one addressable config store: read it whole, seed it, or set a field.
type Scope interface {
	// Read decodes the scope's config into target. A missing file is not an error.
	Read(target any) error
	// Write persists cfg as the whole scope document.
	Write(cfg any) error
	// Set updates a single dotted path (read-modify-write) within the scope.
	Set(path string, value any) error
	// Get returns the value at a dotted path, or nil when absent.
	Get(path string) (any, error)
}

type scope struct {
	store Store
	key   string
}

var _ Scope = (*scope)(nil)

func (s *scope) Read(target any) error {
	if !s.store.Exists(s.key) {
		return nil
	}
	if err := s.store.Load(s.key, target); err != nil {
		return fmt.Errorf("failed to read config %q: %w", s.key, err)
	}
	return nil
}

func (s *scope) Write(cfg any) error {
	if err := s.store.Save(s.key, cfg); err != nil {
		return fmt.Errorf("failed to write config %q: %w", s.key, err)
	}
	return nil
}

func (s *scope) Set(path string, value any) error {
	values := map[string]any{}
	if s.store.Exists(s.key) {
		if err := s.store.Load(s.key, &values); err != nil {
			return fmt.Errorf("failed to read config %q: %w", s.key, err)
		}
	}
	setPath(values, path, value)
	if err := s.store.Save(s.key, values); err != nil {
		return fmt.Errorf("failed to set %q in config %q: %w", path, s.key, err)
	}
	return nil
}

func (s *scope) Get(path string) (any, error) {
	if !s.store.Exists(s.key) {
		return nil, nil
	}
	values := map[string]any{}
	if err := s.store.Load(s.key, &values); err != nil {
		return nil, fmt.Errorf("failed to read config %q: %w", s.key, err)
	}
	v, _ := getPath(values, path)
	return v, nil
}

// missingScope is returned for an unregistered Type so every method reports the
// mistake instead of silently no-oping.
type missingScope struct{ typ Type }

var _ Scope = missingScope{}

func (m missingScope) err() error { return fmt.Errorf("config scope %q is not registered", m.typ) }

func (m missingScope) Read(any) error          { return m.err() }
func (m missingScope) Write(any) error         { return m.err() }
func (m missingScope) Set(string, any) error   { return m.err() }
func (m missingScope) Get(string) (any, error) { return nil, m.err() }

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
