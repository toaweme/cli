package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toaweme/cli/config"
)

type roundTripConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func Test_NewStorage_FileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := NewFileStorage(FileStorage{Dir: dir, Name: "round"})

	assertEqual(t, dir, cfg.Dir())

	want := roundTripConfig{Host: "localhost", Port: 8080}
	assertNoError(t, cfg.Store().Save("config", want))
	assertTrue(t, cfg.Store().Exists("config"))

	var got roundTripConfig
	assertNoError(t, cfg.Store().Load("config", &got))
	assertEqual(t, want.Host, got.Host)
	assertEqual(t, want.Port, got.Port)
}

func Test_NewStorage_SecretsUseSeparateDirAndPerms(t *testing.T) {
	dir := t.TempDir()
	cfg := NewFileStorage(FileStorage{Dir: dir})

	assertNoError(t, cfg.Secrets().Save("token", map[string]string{"value": "s3cr3t"}))

	secretPath := filepath.Join(dir, "secrets", "token.json")
	info, err := os.Stat(secretPath)
	assertNoError(t, err, "secret should be written under the secrets subdir")
	assertEqual(t, os.FileMode(0o600), info.Mode().Perm())

	// regular config must not leak into the secrets store
	assertTrue(t, !cfg.Store().Exists("token"))
}

func Test_Load_Precedence(t *testing.T) {
	type layeredConfig struct {
		Host    string `arg:"host" env:"APP_HOST" json:"host" default:"default-host"`
		Port    int    `arg:"port" env:"APP_PORT" json:"port" default:"1"`
		Region  string `json:"region" default:"us"`
		Verbose bool   `arg:"verbose" json:"verbose"`
	}

	homeDir := t.TempDir()
	projDir := t.TempDir()
	home := config.NewFileStore(homeDir)
	proj := config.NewFileStore(projDir)

	assertNoError(t, home.Save("config", map[string]any{"host": "home-host", "port": 10, "region": "eu"}))
	// project layer overrides host and port, leaves region to the home layer
	assertNoError(t, proj.Save("config", map[string]any{"host": "proj-host", "port": 20}))

	c := &storage{store: proj, dir: projDir, stores: []config.Store{home, proj}}

	t.Setenv("APP_PORT", "30")

	var got layeredConfig
	err := c.Load(&got, LoadOptions{Env: true, Flags: map[string]any{"host": "flag-host"}})
	assertNoError(t, err)

	assertEqual(t, "flag-host", got.Host, "flags are the highest layer")
	assertEqual(t, 30, got.Port, "env overrides the project store")
	assertEqual(t, "eu", got.Region, "home layer wins where no higher layer sets it")
	assertEqual(t, false, got.Verbose, "unset field keeps its zero value")
}

func Test_Load_DefaultsWhenNoFiles(t *testing.T) {
	type layeredConfig struct {
		Host string `arg:"host" json:"host" default:"localhost"`
		Port int    `arg:"port" json:"port" default:"8080"`
	}

	dir := t.TempDir()
	c := NewFileStorage(FileStorage{Dir: dir, Name: "layered"})

	var got layeredConfig
	assertNoError(t, c.Load(&got, LoadOptions{}))

	assertEqual(t, "localhost", got.Host)
	assertEqual(t, 8080, got.Port)
}

func Test_Load_FlagsOverrideDefaults(t *testing.T) {
	type layeredConfig struct {
		Host string `arg:"host" json:"host" default:"localhost"`
	}

	dir := t.TempDir()
	c := NewFileStorage(FileStorage{Dir: dir, Name: "layered"})

	var got layeredConfig
	assertNoError(t, c.Load(&got, LoadOptions{Flags: map[string]any{"host": "0.0.0.0"}}))
	assertEqual(t, "0.0.0.0", got.Host)
}

func Test_ExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	assertNoError(t, err)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"tilde slash", "~/.blink", filepath.Join(home, ".blink")},
		{"bare tilde", "~", home},
		{"absolute untouched", "/etc/blink", "/etc/blink"},
		{"relative untouched", "blink/cfg", "blink/cfg"},
		{"empty untouched", "", ""},
		{"tilde mid string untouched", "/x/~/y", "/x/~/y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqual(t, tt.want, expandHome(tt.in))
		})
	}
}

func Test_ResolveConfigDir(t *testing.T) {
	tests := []struct {
		name  string
		opts  FileStorage
		check func(t *testing.T, dir string)
	}{
		{
			name: "explicit dir wins",
			opts: FileStorage{Dir: "/var/lib/app", Name: "app"},
			check: func(t *testing.T, dir string) {
				assertEqual(t, "/var/lib/app", dir)
			},
		},
		{
			name: "name derives home dir",
			opts: FileStorage{Name: "derived"},
			check: func(t *testing.T, dir string) {
				assertTrue(t, strings.HasSuffix(dir, string(os.PathSeparator)+".derived"))
			},
		},
		{
			name: "tilde dir expands",
			opts: FileStorage{Dir: "~/.tilde"},
			check: func(t *testing.T, dir string) {
				assertTrue(t, !strings.HasPrefix(dir, "~"))
				assertTrue(t, strings.HasSuffix(dir, string(os.PathSeparator)+".tilde"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, resolveConfigDir(tt.opts))
		})
	}
}
