package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/toaweme/cli/config"
)

// FileStorage describes a file-backed configuration source.
type FileStorage struct {
	// Dir is the base directory. A leading "~" expands to the user's home dir.
	// When empty, it defaults to "~/.<Name>".
	Dir string
	// Name is the application name. It derives the default directory and, when
	// PerProject is set, the ".<Name>" directory searched for upward.
	Name string
	// PerProject, when set, walks up from the current working directory looking
	// for a ".<Name>" directory and uses it when found, falling back to Dir.
	// For LoadLayered it adds the project directory as a layer above the home
	// directory rather than replacing it.
	PerProject bool
	// Codecs registers additional codecs (e.g. yaml.Codec, toml.Codec). JSON is
	// always available; the first registered codec becomes the default for keys
	// without a file extension.
	Codecs []config.Codec
}

// NewFileStorage builds a file-backed Storage from opts:
// cli.NewFileStorage(cli.FileStorage{Name: "blink"}).
func NewFileStorage(opts FileStorage) Storage {
	home := resolveHomeDir(opts)
	project := resolveProjectDir(opts, home)

	stores := []config.Store{newFileStore(home, opts.Codecs)}

	// the reported directory is the most-specific location (the project dir when
	// one was found, else the home dir); the project store layers above home.
	dir := home
	if project != "" {
		stores = append(stores, newFileStore(project, opts.Codecs))
		dir = project
	}

	secrets := config.NewSecretFileStore(filepath.Join(dir, "secrets"))
	for _, codec := range opts.Codecs {
		secrets.AddCodec(codec)
	}

	return newStorage(stores, secrets, dir)
}

func newFileStore(dir string, codecs []config.Codec) config.Store {
	store := config.NewFileStore(dir)
	for _, codec := range codecs {
		store.AddCodec(codec)
	}
	return store
}

// resolveConfigDir returns the most-specific config directory: the per-project
// ".<Name>" directory when one is found, otherwise the home directory.
func resolveConfigDir(opts FileStorage) string {
	home := resolveHomeDir(opts)
	if project := resolveProjectDir(opts, home); project != "" {
		return project
	}
	return home
}

// resolveHomeDir resolves the base (home) directory for opts, expanding "~".
func resolveHomeDir(opts FileStorage) string {
	dir := opts.Dir
	if dir == "" && opts.Name != "" {
		dir = config.HomePath(opts.Name)
	}
	return expandHome(dir)
}

// resolveProjectDir returns the discovered per-project directory, or "" when
// PerProject is off, no directory is found, or it resolves to the home dir.
func resolveProjectDir(opts FileStorage, home string) string {
	if !opts.PerProject || opts.Name == "" {
		return ""
	}
	dir := discoverProjectDir(opts.Name)
	if dir == "" || dir == home {
		return ""
	}
	return dir
}

// discoverProjectDir walks up from the current directory looking for a
// ".<name>" directory and returns its path, or "" if none is found.
func discoverProjectDir(name string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	found := config.Discover(cwd, []string{"." + name})
	if found == "" {
		return ""
	}

	info, err := os.Stat(found)
	if err != nil || !info.IsDir() {
		return ""
	}

	return found
}

// expandHome replaces a leading "~" in dir with the user's home directory.
func expandHome(dir string) string {
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(dir, "~"))
		}
	}

	return dir
}
