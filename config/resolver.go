package config

import (
	"fmt"
)

// Source is a per-command field mapping value: either a string (a dotted path into
// the merged config) or a func() (any, error) computing the value.
type Source = any

// storeResolver resolves one Store's config file into a command's option values,
// with optional per-command field mapping. It is a middleware: Resolve receives the
// values accumulated by earlier resolvers and overlays its own layer on top, so an
// App that registers several (App.Resolve takes many) layers them lowest-precedence
// first. It satisfies the cli.Resolver shape structurally and does not import cli.
type storeResolver struct {
	store Store
	rules map[string]map[string]Source
}

// NewResolver builds a resolver over a single store. rules optionally maps a command
// path (e.g. "db migrate") to per-field Sources, each a dotted config path or a
// func() (any, error); pass nil for none. A mapped field overrides the value sourced
// directly from the file. Compose resolvers by registering several on the App,
// low-to-high precedence: app.Resolve(global, project, secrets).
func NewResolver(store Store, rules map[string]map[string]Source) *storeResolver {
	return &storeResolver{store: store, rules: rules}
}

// Resolve overlays this store's file (and any per-command mapping) onto values, the
// map accumulated by earlier resolvers, and returns it. The framework folds env and
// then flags on top afterwards, so the effective order is earlier-resolvers < this
// store < mapping < env < flags.
func (r *storeResolver) Resolve(cmd string, values map[string]any) (map[string]any, error) {
	if values == nil {
		values = map[string]any{}
	}

	layer := map[string]any{}
	if err := r.store.Read(&layer); err != nil {
		return nil, fmt.Errorf("failed to read config for command %q: %w", cmd, err)
	}
	deepMerge(values, layer)

	for field, src := range r.rules[cmd] {
		value, ok, err := resolveSource(src, values)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve mapping for command %q field %q: %w", cmd, field, err)
		}
		if ok {
			values[field] = value
		}
	}

	return values, nil
}

// resolveSource evaluates a mapping Source against the merged config.
func resolveSource(src Source, merged map[string]any) (any, bool, error) {
	switch s := src.(type) {
	case string:
		value, ok := getPath(merged, s)
		return value, ok, nil
	case func() (any, error):
		value, err := s()
		if err != nil {
			return nil, false, err
		}
		return value, true, nil
	default:
		return nil, false, fmt.Errorf("unsupported mapping source type %T", src)
	}
}

// deepMerge overlays src onto dst, recursing into nested map[string]any values so a
// higher layer overrides leaf-by-leaf rather than replacing whole subtrees.
func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if sv, ok := v.(map[string]any); ok {
			if dv, ok := dst[k].(map[string]any); ok {
				deepMerge(dv, sv)
				continue
			}
		}
		dst[k] = v
	}
}
