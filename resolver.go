package cli

import (
	"os"
	"strings"
)

// env folds the process environment into commandOptions, keyed by variable name,
// so fields are matched by their `env:` tag during the merge. The framework folds
// env in after the resolver chain and before flags, so env beats files but loses to a typed flag.
func env(commandOptions map[string]any) {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			commandOptions[pair[0]] = pair[1]
		}
	}
}
