package config

import (
	"fmt"
	"os"
	"strings"
)

// Source is a per-command field mapping value: either a string (a dotted path into
// the merged config) or a func() (any, error) computing the value.
type Source = any

// fileResolver resolves config files (low to high precedence across a Config's
// sources) and environment variables, with optional per-command field mapping. It
// returns the values for a command's options; the cli framework overlays parsed
// flags on top, so a typed flag always wins. It satisfies the cli.Resolver shape
// structurally and does not import cli.
type fileResolver struct {
	cfg   *Config
	rules map[string]map[string]Source
}

// NewFileResolver builds a resolver over cfg's config sources. rules optionally maps
// a command path (e.g. "db migrate") to per-field Sources, each a dotted config path
// or a func() (any, error); pass nil for none. A mapped field overrides the value
// sourced directly from the config files.
func NewFileResolver(cfg *Config, rules map[string]map[string]Source) *fileResolver {
	return &fileResolver{cfg: cfg, rules: rules}
}

// Resolve merges config files lowest-precedence-first, applies any per-command
// mapping, then folds in environment variables. The framework applies flags after,
// so the effective order is files < mapping < env < flags.
func (r *fileResolver) Resolve(cmd string, flags map[string]any) (map[string]any, error) {
	merged := map[string]any{}

	for _, f := range r.cfg.handlers {
		if !f.store.Exists(f.name) {
			continue
		}
		layer := map[string]any{}
		if err := f.store.Load(f.name, &layer); err != nil {
			return nil, fmt.Errorf("failed to load %s config: %w", f.configType, err)
		}
		deepMerge(merged, layer)
	}

	for field, src := range r.rules[cmd] {
		value, ok, err := resolveSource(src, merged)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve mapping for command %q field %q: %w", cmd, field, err)
		}
		if ok {
			merged[field] = value
		}
	}

	// environment variables matched by `env:` tag during overlay; env beats files.
	readEnv(merged)

	return merged, nil
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

// readEnv folds the process environment into dst, keyed by variable name.
func readEnv(dst map[string]any) {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			dst[pair[0]] = pair[1]
		}
	}
}
