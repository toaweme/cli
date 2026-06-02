package config

import (
	"os"
	"path/filepath"
	"strings"
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
	return filepath.Join(home, appName)
}

// ExpandHome replaces a leading "~" in dir with the user's home directory. A bare
// "~" or "~/..." expands; "~" anywhere else in the path is left untouched.
func ExpandHome(dir string) string {
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(dir, "~"))
		}
	}
	return dir
}
