package cli

import (
	"os"
	"strings"
)

// ResolverDefault is the built-in resolver used when an app sets none. It returns
// the process environment only; the framework overlays parsed flags on top, so the
// effective behavior is `default:` tags, then env, then flags. It reads no config
// files - file-backed resolution lives in the config package (config.NewFileResolver).
var ResolverDefault Resolver = envResolver{}

type envResolver struct{}

var _ Resolver = envResolver{}

func (envResolver) Resolve(_ string, _ map[string]any) (map[string]any, error) {
	values := map[string]any{}
	env(values)
	return values, nil
}

// env folds the process environment into commandOptions, keyed by variable name,
// so fields are matched by their `env:` tag during the merge.
func env(commandOptions map[string]any) {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			commandOptions[pair[0]] = pair[1]
		}
	}
}
