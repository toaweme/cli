package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrDotenvNotFound is returned when a requested .env file does not exist.
var ErrDotenvNotFound = errors.New("dotenv file not found")

// LoadDotEnv reads .env files and sets environment variables that are not already set.
// With no arguments, it loads ".env" from the current working directory.
// Silently skips files that do not exist.
func LoadDotEnv(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{".env"}
	}

	for _, path := range paths {
		values, err := readDotEnv(path)
		if err != nil {
			if errors.Is(err, ErrDotenvNotFound) {
				continue
			}
			return fmt.Errorf("failed to load env file %q: %w", path, err)
		}

		for key, value := range values {
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, value)
			}
		}
	}

	return nil
}

// GetDotEnv parses a single .env file and returns its key/value pairs without touching the process environment.
// Returns ErrDotenvNotFound if the file does not exist.
func GetDotEnv(file string) (map[string]string, error) {
	return readDotEnv(file)
}

// GetDotEnvs parses several .env files and merges them into a single map without
// touching the process environment. Earlier files take precedence: a key set by an
// earlier file is not overwritten by a later one. Returns the first error encountered,
// including ErrDotenvNotFound for a missing file.
func GetDotEnvs(files ...string) (map[string]string, error) {
	merged := make(map[string]string)

	for _, file := range files {
		values, err := readDotEnv(file)
		if err != nil {
			return nil, err
		}

		for key, value := range values {
			if _, exists := merged[key]; !exists {
				merged[key] = value
			}
		}
	}

	return merged, nil
}

// readDotEnv opens and parses a .env file into a map. It returns ErrDotenvNotFound when the file does not exist.
func readDotEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%q: %w", path, ErrDotenvNotFound)
		}
		return nil, fmt.Errorf("failed to open env file %q: %w", path, err)
	}
	defer f.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := parseLine(line)
		if !ok {
			continue
		}

		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read env file %q: %w", path, err)
	}

	return values, nil
}

func parseLine(line string) (string, string, bool) {
	idx := strings.IndexByte(line, '=')
	if idx < 1 {
		return "", "", false
	}

	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])

	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}

	return key, value, true
}
