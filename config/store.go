package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	jsoncodec "github.com/toaweme/cli/config/addons/json"
)

// FileStore is a single config file at dir/name, encoded by one codec.
// A name with an explicit extension is used verbatim ("config.yaml");
// a name without one gets the codec's extension appended ("config" -> "config.json").
type FileStore struct {
	dir   string
	name  string
	perm  os.FileMode
	codec Codec
}

var _ Store = (*FileStore)(nil)

// NewFileStore creates a file-based store for the single file name within dir, at 0o644.
// An empty name defaults to "config". Pass the codec the file is encoded with (at most one);
// with none it defaults to JSON (config/addons/json). Use FileSecrets for a 0o600 store.
func NewFileStore(dir, name string, codec ...Codec) *FileStore {
	if name == "" {
		name = "config"
	}
	c := Codec(jsoncodec.New())
	if len(codec) > 0 {
		c = codec[0]
	}
	return &FileStore{dir: ExpandHome(dir), name: name, perm: 0o644, codec: c}
}

// Read decodes the whole file into target. A missing file is not an error.
func (s *FileStore) Read(target any) error {
	if !s.Exists() {
		return nil
	}
	path, codec := s.resolve()
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config %q: %w", s.name, err)
	}
	if err := codec.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to decode config %q: %w", s.name, err)
	}
	return nil
}

// Write persists value as the whole file, creating parent directories as needed.
func (s *FileStore) Write(value any) error {
	path, codec := s.resolve()
	data, err := codec.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to encode config %q: %w", s.name, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory for %q: %w", s.name, err)
	}

	if err := atomicWrite(path, data, s.perm); err != nil {
		return fmt.Errorf("failed to write config %q: %w", s.name, err)
	}

	return nil
}

// Exists reports whether the file exists.
func (s *FileStore) Exists() bool {
	path, _ := s.resolve()
	_, err := os.Stat(path)
	return err == nil
}

// KeyRead returns the value at a dotted path within the file, or nil when absent.
func (s *FileStore) KeyRead(key string) (any, error) {
	if !s.Exists() {
		return nil, nil
	}
	values := map[string]any{}
	if err := s.Read(&values); err != nil {
		return nil, err
	}
	v, _ := getPath(values, key)
	return v, nil
}

// KeyWrite sets a single dotted path within the file (read-modify-write).
func (s *FileStore) KeyWrite(key string, value any) error {
	values := map[string]any{}
	if err := s.Read(&values); err != nil {
		return err
	}
	setPath(values, key, value)
	if err := s.Write(values); err != nil {
		return fmt.Errorf("failed to set %q in config %q: %w", key, s.name, err)
	}
	return nil
}

// KeyExists reports whether a dotted path is present in the file.
func (s *FileStore) KeyExists(key string) bool {
	if !s.Exists() {
		return false
	}
	values := map[string]any{}
	if err := s.Read(&values); err != nil {
		return false
	}
	_, ok := getPath(values, key)
	return ok
}

// KeyDelete removes a single dotted path within the file (read-modify-write).
// A missing file or absent path is not an error.
func (s *FileStore) KeyDelete(key string) error {
	if !s.Exists() {
		return nil
	}
	values := map[string]any{}
	if err := s.Read(&values); err != nil {
		return err
	}
	if !deletePath(values, key) {
		return nil
	}
	if err := s.Write(values); err != nil {
		return fmt.Errorf("failed to delete %q in config %q: %w", key, s.name, err)
	}
	return nil
}

// Delete removes the file. Returns nil if it does not exist.
func (s *FileStore) Delete() error {
	path, _ := s.resolve()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config %q: %w", s.name, err)
	}
	return nil
}

// Dir returns the base directory of the store.
func (s *FileStore) Dir() string {
	return s.dir
}

// resolve returns the full file path and the store's codec.
// A name with an explicit extension is used verbatim; otherwise the codec's extension is appended.
func (s *FileStore) resolve() (string, Codec) {
	name := s.name
	if filepath.Ext(name) == "" {
		name += s.codec.Extension()
	}
	return filepath.Join(s.dir, name), s.codec
}

// atomicWrite writes data to path via a temporary file and rename.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// setPath writes value at a dotted path within m, creating nested maps as needed.
func setPath(m map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	for _, p := range parts[:len(parts)-1] {
		next, ok := m[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[p] = next
		}
		m = next
	}
	m[parts[len(parts)-1]] = value
}

// deletePath removes the leaf at a dotted path within m, reporting whether anything was removed.
// Intermediate maps are left in place even if they become empty.
func deletePath(m map[string]any, path string) bool {
	parts := strings.Split(path, ".")
	for _, p := range parts[:len(parts)-1] {
		next, ok := m[p].(map[string]any)
		if !ok {
			return false
		}
		m = next
	}
	leaf := parts[len(parts)-1]
	if _, ok := m[leaf]; !ok {
		return false
	}
	delete(m, leaf)
	return true
}

// getPath walks a dotted path through nested map[string]any values.
func getPath(m map[string]any, path string) (any, bool) {
	var current any = m
	for _, p := range strings.Split(path, ".") {
		mm, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = mm[p]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
