package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func Test_NewConfig_FileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := NewConfig(NewFileConfig(FileConfig{Dir: dir, Name: "round"}))

	assertEqual(t, dir, cfg.Dir())

	want := roundTripConfig{Host: "localhost", Port: 8080}
	assertNoError(t, cfg.Save("config", want))
	assertTrue(t, cfg.Exists("config"))

	var got roundTripConfig
	assertNoError(t, cfg.Load("config", &got))
	assertEqual(t, want.Host, got.Host)
	assertEqual(t, want.Port, got.Port)
}

func Test_NewConfig_SecretsUseSeparateDirAndPerms(t *testing.T) {
	dir := t.TempDir()
	cfg := NewConfig(NewFileConfig(FileConfig{Dir: dir}))

	assertNoError(t, cfg.Secrets().Save("token", map[string]string{"value": "s3cr3t"}))

	secretPath := filepath.Join(dir, "secrets", "token.json")
	info, err := os.Stat(secretPath)
	assertNoError(t, err, "secret should be written under the secrets subdir")
	assertEqual(t, os.FileMode(0o600), info.Mode().Perm())

	// regular config must not leak into the secrets store
	assertTrue(t, !cfg.Exists("token"))
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
		opts  FileConfig
		check func(t *testing.T, dir string)
	}{
		{
			name: "explicit dir wins",
			opts: FileConfig{Dir: "/var/lib/app", Name: "app"},
			check: func(t *testing.T, dir string) {
				assertEqual(t, "/var/lib/app", dir)
			},
		},
		{
			name: "name derives home dir",
			opts: FileConfig{Name: "derived"},
			check: func(t *testing.T, dir string) {
				assertTrue(t, strings.HasSuffix(dir, string(os.PathSeparator)+".derived"))
			},
		},
		{
			name: "tilde dir expands",
			opts: FileConfig{Dir: "~/.tilde"},
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
