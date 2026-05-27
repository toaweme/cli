package config

import (
	"os"
	"path/filepath"
)

// Discover walks up from start looking for a config file matching any of the
// given names. Returns the full path of the first match, or empty string.
func Discover(start string, names []string) string {
	dir, err := filepath.Abs(start)
	if err != nil {
		return ""
	}

	for {
		for _, name := range names {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// HomePath returns the config directory under the user's home directory.
// Returns ~/.appName/ or empty string if home cannot be determined.
func HomePath(appName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "."+appName)
}
