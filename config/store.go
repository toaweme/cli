package config

import (
	"fmt"
	"os"
	"path/filepath"

	jsoncodec "github.com/toaweme/cli/config/addons/json"
)

// multiCodec is implemented by codecs that recognize more than one file extension
// when reading (e.g. YAML: ".yml" and ".yaml"). AddCodec registers every extension
// it reports; the primary Extension() is still the one used for writing.
type multiCodec interface {
	Extensions() []string
}

// codecExtensions returns every extension a codec should be registered under: all of
// Extensions() when it implements multiCodec, otherwise just its primary Extension().
func codecExtensions(codec Codec) []string {
	if mc, ok := codec.(multiCodec); ok {
		if exts := mc.Extensions(); len(exts) > 0 {
			return exts
		}
	}
	return []string{codec.Extension()}
}

// FileStore handles file-based config storage with one or more codecs. The codec is
// selected by file extension ("key.yml" uses the YAML codec); a key without an
// extension uses the default codec (the first one registered). More codecs can be
// added later with AddCodec.
type FileStore struct {
	dir      string
	perm     os.FileMode
	codecs   map[string]Codec
	fallback Codec
}

var _ Store = (*FileStore)(nil)
var _ SecretStore = (*FileStore)(nil)
var _ SecretBackend = (*FileStore)(nil) // FileSecrets returns one as a SecretBackend

// NewFileStore creates a file-based store at dir with files at 0o644. Pass the
// codecs the store should use; the first is the default for extension-less keys.
// With no codecs it defaults to JSON (config/addons/json). Pass only a YAML codec,
// for example, and the store never touches JSON. Use FileSecrets for a 0o600 store.
func NewFileStore(dir string, codecs ...Codec) *FileStore {
	if len(codecs) == 0 {
		codecs = []Codec{jsoncodec.New()}
	}
	s := &FileStore{dir: ExpandHome(dir), perm: 0o644, codecs: map[string]Codec{}}
	for _, codec := range codecs {
		for _, ext := range codecExtensions(codec) {
			s.codecs[ext] = codec
		}
	}
	// the first codec is the default for keys without an extension.
	s.fallback = codecs[0]
	return s
}

// AddCodec registers a codec under every extension it recognizes (see multiCodec).
// The most recently added codec becomes the default for keys without an explicit
// extension.
func (s *FileStore) AddCodec(codec Codec) *FileStore {
	for _, ext := range codecExtensions(codec) {
		s.codecs[ext] = codec
	}
	s.fallback = codec
	return s
}

// SetDefault registers a codec (under every extension it recognizes) and makes it
// the default used for keys without a file extension.
func (s *FileStore) SetDefault(codec Codec) *FileStore {
	for _, ext := range codecExtensions(codec) {
		s.codecs[ext] = codec
	}
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
