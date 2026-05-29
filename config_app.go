package cli

import (
	"fmt"

	"github.com/toaweme/structs"

	"github.com/toaweme/cli/config"
)

// storage is the default Storage accessor over a set of ordered config stores.
type storage struct {
	store   config.Store
	secrets config.Store
	dir     string
	stores  []config.Store
}

var _ Storage = (*storage)(nil)

// newStorage builds the accessor from config stores ordered lowest precedence
// first (the last, most-specific store is the primary one for direct access),
// the secrets store, and the base directory.
func newStorage(stores []config.Store, secrets config.Store, dir string) *storage {
	return &storage{store: stores[len(stores)-1], secrets: secrets, dir: dir, stores: stores}
}

func (s *storage) Store() config.Store   { return s.store }
func (s *storage) Secrets() config.Store { return s.secrets }
func (s *storage) Dir() string           { return s.dir }

func (s *storage) Load(target any, opts LoadOptions) error {
	key := opts.Key
	if key == "" {
		key = "config"
	}

	manager := structs.New(target, structs.DefaultRules, structs.WithTags(defaultTags...))

	// seed struct `default:` tags as the lowest layer
	if err := manager.Set(map[string]any{}); err != nil {
		return fmt.Errorf("failed to apply config defaults: %w", err)
	}

	// config stores, low -> high (home, then project)
	for _, store := range s.stores {
		if !store.Exists(key) {
			continue
		}
		values := map[string]any{}
		if err := store.Load(key, &values); err != nil {
			return fmt.Errorf("failed to load config layer %q: %w", key, err)
		}
		if err := manager.Set(values); err != nil {
			return fmt.Errorf("failed to apply config layer %q: %w", key, err)
		}
	}

	// environment variables, matched via `env:` tags
	if opts.Env {
		values := map[string]any{}
		env(values)
		if err := manager.Set(values); err != nil {
			return fmt.Errorf("failed to apply config environment: %w", err)
		}
	}

	// explicit flags win
	if len(opts.Flags) > 0 {
		if err := manager.Set(opts.Flags); err != nil {
			return fmt.Errorf("failed to apply config flags: %w", err)
		}
	}

	return nil
}
