package cli

import (
	"github.com/toaweme/cli/config"
)

// Storage is the configuration accessor available to commands. It wraps a
// primary key-value store, a separate secrets store (0600-permission files),
// and a layered Load that merges a typed config from several sources.
//
// NewFileStorage returns the file-backed implementation; implement Storage
// directly to back it with something else (in-memory, remote, ...).
type Storage interface {
	// Store returns the primary key-value store for direct Load/Save by key.
	Store() config.Store
	// Secrets returns the secrets store, backed by 0600-permission files under
	// a "secrets" subdirectory of the config directory.
	Secrets() config.Store
	// Dir returns the resolved base directory backing this storage.
	Dir() string
	// Load merges configuration into target from layered sources, lowest
	// precedence first: struct `default:` tags, then each config store (home,
	// then any per-project store discovered by walking up), then environment
	// variables (when opts.Env is set, matched via `env:` tags), then opts.Flags.
	// Later layers override earlier ones field by field; absent layers are skipped.
	Load(target any, opts LoadOptions) error
}

// LoadOptions configures a layered load. See Storage.Load.
type LoadOptions struct {
	// Key is the store key (file name) read from each config store layer.
	// Defaults to "config" when empty.
	Key string
	// Env, when set, folds environment variables in as a layer above the config
	// stores. Fields are matched by their `env:` tag.
	Env bool
	// Flags is the highest-precedence layer, typically the parsed CLI options.
	// Keys are matched the same way struct inputs are (arg/short/json/yaml tags).
	Flags map[string]any
}
