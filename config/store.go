package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Store reads and writes configuration values by key.
// Keys map to files within the store's base directory.
type Store interface {
	// Load reads the value at key into target.
	Load(key string, target any) error
	// Save writes value to key. Creates parent directories as needed.
	Save(key string, value any) error
	// Delete removes the value at key. Returns nil if key does not exist.
	Delete(key string) error
	// Exists reports whether a value exists at key.
	Exists(key string) bool
}

// SecretStore is a Store with restricted file permissions (0o600).
// Implementations must ensure secrets are never written world-readable.
type SecretStore interface {
	Store
}

// Codec serializes and deserializes config values.
type Codec interface {
	// Marshal encodes v into bytes.
	Marshal(v any) ([]byte, error)
	// Unmarshal decodes data into v.
	Unmarshal(data []byte, v any) error
	// Extension returns the file extension for this codec (e.g. ".json", ".yaml").
	Extension() string
}

// jsonCodec is the built-in JSON codec using stdlib encoding/json.
type jsonCodec struct{}

func (c *jsonCodec) Marshal(v any) ([]byte, error)      { return json.MarshalIndent(v, "", "  ") }
func (c *jsonCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
func (c *jsonCodec) Extension() string                  { return ".json" }

// FileStore handles file-based config storage with support for multiple codecs.
// JSON is built-in. Additional codecs (YAML, TOML) can be registered via AddCodec.
// The codec is selected by file extension: "key.yaml" uses the YAML codec,
// "key" without an extension uses the default codec.
type FileStore struct {
	dir      string
	perm     os.FileMode
	codecs   map[string]Codec
	fallback Codec
}

// NewFileStore creates a file-based store at dir.
// JSON is the default codec. Files are created with 0o644 permissions.
func NewFileStore(dir string) *FileStore {
	jc := &jsonCodec{}
	return &FileStore{
		dir:      dir,
		perm:     0o644,
		codecs:   map[string]Codec{".json": jc},
		fallback: jc,
	}
}

// NewSecretFileStore creates a file-based store at dir.
// JSON is the default codec. Files are created with 0o600 permissions.
func NewSecretFileStore(dir string) *FileStore {
	jc := &jsonCodec{}
	return &FileStore{
		dir:      dir,
		perm:     0o600,
		codecs:   map[string]Codec{".json": jc},
		fallback: jc,
	}
}

// AddCodec registers a codec for its file extension.
// If this is the first non-JSON codec added, it becomes the default
// for keys without an explicit extension.
func (s *FileStore) AddCodec(codec Codec) *FileStore {
	ext := codec.Extension()
	s.codecs[ext] = codec
	s.fallback = codec
	return s
}

// SetDefault sets the default codec used for keys without a file extension.
func (s *FileStore) SetDefault(codec Codec) *FileStore {
	s.codecs[codec.Extension()] = codec
	s.fallback = codec
	return s
}

func (s *FileStore) Load(key string, target any) error {
	path, codec := s.resolve(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to load config %q: %w", key, err)
	}
	if err := codec.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to decode config %q: %w", key, err)
	}
	return nil
}

func (s *FileStore) Save(key string, value any) error {
	path, codec := s.resolve(key)
	data, err := codec.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to encode config %q: %w", key, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory for %q: %w", key, err)
	}

	if err := atomicWrite(path, data, s.perm); err != nil {
		return fmt.Errorf("failed to write config %q: %w", key, err)
	}

	return nil
}

func (s *FileStore) Delete(key string) error {
	path, _ := s.resolve(key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config %q: %w", key, err)
	}
	return nil
}

func (s *FileStore) Exists(key string) bool {
	path, _ := s.resolve(key)
	_, err := os.Stat(path)
	return err == nil
}

// Dir returns the base directory of the store.
func (s *FileStore) Dir() string {
	return s.dir
}

// resolve returns the full file path and the codec to use for the given key.
// If the key has a known extension, that codec is used.
// Otherwise the fallback codec is used and its extension is appended.
func (s *FileStore) resolve(key string) (string, Codec) {
	ext := filepath.Ext(key)
	if codec, ok := s.codecs[ext]; ok {
		return filepath.Join(s.dir, key), codec
	}
	return filepath.Join(s.dir, key+s.fallback.Extension()), s.fallback
}

// atomicWrite writes data to path via a temporary file and rename.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
