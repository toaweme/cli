package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/toaweme/cli/config"
)

// Config is the application configuration accessor available to commands.
// It exposes a primary store for regular configuration plus a separate secrets
// store that writes with restricted (0600) file permissions.
type Config interface {
	config.Store
	// Secrets returns the secrets store, backed by 0600-permission files under
	// a "secrets" subdirectory of the config directory.
	Secrets() config.Store
	// Dir returns the resolved base directory backing this config.
	Dir() string
}

// ConfigSource builds the stores backing a Config. Implement it to provide a
// custom backend (in-memory, remote, ...); NewFileConfig provides the
// file-backed source used by most applications.
type ConfigSource interface {
	// Stores returns the primary config store, the secrets store, and the
	// resolved base directory.
	Stores() (cfg config.Store, secrets config.Store, dir string)
}

// FileConfig describes a file-backed configuration source.
type FileConfig struct {
	// Dir is the base directory. A leading "~" expands to the user's home dir.
	// When empty, it defaults to "~/.<Name>".
	Dir string
	// Name is the application name. It derives the default directory and, when
	// PerProject is set, the ".<Name>" directory searched for upward.
	Name string
	// PerProject, when set, walks up from the current working directory looking
	// for a ".<Name>" directory and uses it when found, falling back to Dir.
	PerProject bool
	// Codecs registers additional codecs (e.g. yaml.Codec, toml.Codec). JSON is
	// always available; the first registered codec becomes the default for keys
	// without a file extension.
	Codecs []config.Codec
}

type fileSource struct {
	cfg     config.Store
	secrets config.Store
	dir     string
}

var _ ConfigSource = (*fileSource)(nil)

func (s *fileSource) Stores() (config.Store, config.Store, string) {
	return s.cfg, s.secrets, s.dir
}

// NewFileConfig builds a file-backed ConfigSource from opts. Pass the result to
// NewConfig: cli.NewConfig(cli.NewFileConfig(cli.FileConfig{Name: "blink"})).
func NewFileConfig(opts FileConfig) ConfigSource {
	dir := resolveConfigDir(opts)

	cfg := config.NewFileStore(dir)
	secrets := config.NewSecretFileStore(filepath.Join(dir, "secrets"))
	for _, codec := range opts.Codecs {
		cfg.AddCodec(codec)
		secrets.AddCodec(codec)
	}

	return &fileSource{cfg: cfg, secrets: secrets, dir: dir}
}

type appConfig struct {
	config.Store
	secrets config.Store
	dir     string
}

var _ Config = (*appConfig)(nil)

func (c *appConfig) Secrets() config.Store { return c.secrets }
func (c *appConfig) Dir() string           { return c.dir }

// NewConfig wraps a ConfigSource into the Config accessor. Hand it to the app
// via Settings.Config and pass it to the commands that need it.
func NewConfig(src ConfigSource) Config {
	cfg, secrets, dir := src.Stores()
	return &appConfig{Store: cfg, secrets: secrets, dir: dir}
}

func resolveConfigDir(opts FileConfig) string {
	if opts.PerProject && opts.Name != "" {
		if dir := discoverProjectDir(opts.Name); dir != "" {
			return dir
		}
	}

	dir := opts.Dir
	if dir == "" && opts.Name != "" {
		dir = config.HomePath(opts.Name)
	}

	return expandHome(dir)
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
