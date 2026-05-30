package cli

import (
	"fmt"
	"strings"

	"github.com/toaweme/structs"

	"github.com/toaweme/cli/config"
)

// storage is the default Storage accessor over a set of ordered config stores.
// The primary store is embedded so its Save/Load/Delete/Exists are promoted onto
// storage directly (store.Save(...) instead of store.Store().Save(...)).
type storage struct {
	config.Store
	secrets config.Store
	dir     string
	stores  []config.Store
}

var _ Storage = (*storage)(nil)

// newStorage builds the accessor from config stores ordered lowest precedence
// first (the last, most-specific store is the primary one for direct access),
// the secrets store, and the base directory.
func newStorage(stores []config.Store, secrets config.Store, dir string) *storage {
	return &storage{Store: stores[len(stores)-1], secrets: secrets, dir: dir, stores: stores}
}

func (s *storage) Secrets() config.Store { return s.secrets }
func (s *storage) Dir() string           { return s.dir }

func (s *storage) Resolve(target any, opts LoadOptions) error {
	key := opts.Key
	if key == "" {
		key = "config"
	}

	return mergeConfig(target, s.stores, key, opts.Env, opts.Flags, opts.Mapping)
}

// mergeConfig layers configuration into target, lowest precedence first: struct
// `default:` tags, then each config store (in order) under key, then environment
// variables (when useEnv is set), then flags. Each layer is a separate Set so
// later layers override earlier ones field by field; a single merged map cannot
// express this because structs.Set short-circuits on the first matching tag.
//
// mapping, when set, remaps target fields onto arbitrary paths within each config
// store layer (see ConfigMapping); it is applied after the plain tag match for
// that layer so an explicit mapping wins. It does not touch env or flags.
//
// stores may be nil (no config files), which is the no-Store path: defaults then
// env then flags only. This is also why flags must beat env, the bug a single
// Set would reintroduce by checking the `env:` tag first.
func mergeConfig(target any, stores []config.Store, key string, useEnv bool, flags map[string]any, mapping ConfigMapping) error {
	manager := structs.New(target, structs.DefaultRules, structs.WithTags(defaultTags...))

	// seed struct `default:` tags as the lowest layer
	if err := manager.Set(map[string]any{}); err != nil {
		return fmt.Errorf("failed to apply config defaults: %w", err)
	}

	// config stores, low -> high (home, then project)
	for _, store := range stores {
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
		if overlay := remapConfig(target, values, mapping); len(overlay) > 0 {
			if err := manager.Set(overlay); err != nil {
				return fmt.Errorf("failed to apply config mapping for layer %q: %w", key, err)
			}
		}
	}

	// environment variables, matched via `env:` tags
	if useEnv {
		values := map[string]any{}
		env(values)
		if err := manager.Set(values); err != nil {
			return fmt.Errorf("failed to apply config environment: %w", err)
		}
	}

	// explicit flags win
	if len(flags) > 0 {
		if err := manager.Set(flags); err != nil {
			return fmt.Errorf("failed to apply config flags: %w", err)
		}
	}

	return nil
}

// TopLevelMapping is the identity mapping: every field is sourced from its own
// top-level config key. Equivalent to the default tag match, stated explicitly.
var TopLevelMapping = ConfigMapping{"*": "*"}

// Namespaced maps every field from under a single config namespace, e.g.
// Namespaced("http") sources each field from "http.<field>". It is shorthand for
// ConfigMapping{"*": ns + ".*"}.
func Namespaced(ns string) ConfigMapping {
	return ConfigMapping{"*": ns + ".*"}
}

// remapConfig builds an overlay from a ConfigMapping by resolving each entry
// against values (dotted paths walking nested maps) and keying the result under
// the field, so a later Set matches it the usual way.
//
// The "*" key applies a path template to every top-level field of target; a "*"
// segment in a value is replaced by the field's own config key (json, then yaml,
// then arg, then the Go field name). So {"*":"*"} is the identity mapping and
// {"*":"http.*"} sources every field from under "http.". Explicit field entries
// are applied after, overriding the wildcard for that field. Paths absent in
// values are skipped; returns nil when there is nothing to overlay.
func remapConfig(target any, values map[string]any, mapping ConfigMapping) map[string]any {
	if len(mapping) == 0 {
		return nil
	}

	overlay := map[string]any{}

	if template, ok := mapping["*"]; ok {
		fields, err := structs.GetStructFields(target, nil, structs.DefaultEncodingTags)
		if err == nil {
			for _, field := range fields {
				fk := fieldConfigKey(field)
				if v, found := resolveConfigPath(values, expandConfigPath(template, fk)); found {
					overlay[fk] = v
				}
			}
		}
	}

	for field, path := range mapping {
		if field == "*" {
			continue
		}
		if v, found := resolveConfigPath(values, expandConfigPath(path, field)); found {
			overlay[field] = v
		}
	}

	if len(overlay) == 0 {
		return nil
	}
	return overlay
}

// fieldConfigKey returns the key a field is addressed by in a config file: its
// json tag, then yaml, then arg, falling back to the Go field name.
func fieldConfigKey(field structs.Field) string {
	for _, tag := range []string{"json", "yaml", "arg"} {
		if v := field.Tags[tag]; v != "" {
			return v
		}
	}
	return field.Name
}

// expandConfigPath replaces each "*" segment of a dotted path template with the
// field's own key, so "*" -> fieldKey and "http.*" -> "http.<fieldKey>". An empty
// template resolves to the field key alone.
func expandConfigPath(template, fieldKey string) string {
	if template == "" {
		return fieldKey
	}
	parts := strings.Split(template, ".")
	for i, part := range parts {
		if part == "*" {
			parts[i] = fieldKey
		}
	}
	return strings.Join(parts, ".")
}

// resolveConfigPath walks a dotted path through nested map[string]any values,
// returning the value at the path and whether it was found.
func resolveConfigPath(values map[string]any, path string) (any, bool) {
	if path == "" {
		return nil, false
	}
	var current any = values
	for _, part := range strings.Split(path, ".") {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := m[part]
		if !ok {
			return nil, false
		}
		current = v
	}
	return current, true
}
