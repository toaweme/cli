package config

import (
	"os"
	"path/filepath"
	"testing"
)

type sampleConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func mustFrom(t *testing.T, cfg *Config, configType Type) *handler {
	t.Helper()
	f, err := cfg.From(configType)
	if err != nil {
		t.Fatalf("From(%q): %v", configType, err)
	}
	return f
}

func Test_File_WriteRead(t *testing.T) {
	dir := t.TempDir()
	cfg := New().Add(Global, NewFileStore(dir), "config")

	want := sampleConfig{Host: "localhost", Port: 8080}
	if err := mustFrom(t, cfg, Global).Write(want); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	var got sampleConfig
	if err := mustFrom(t, cfg, Global).Read(&got); err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func Test_File_ReadMissingIsNoError(t *testing.T) {
	cfg := New().Add(Global, NewFileStore(t.TempDir()), "config")
	var got sampleConfig
	if err := mustFrom(t, cfg, Global).Read(&got); err != nil {
		t.Fatalf("reading a missing file should not error: %v", err)
	}
	if got != (sampleConfig{}) {
		t.Fatalf("expected zero value, got %+v", got)
	}
}

func Test_File_SetGet(t *testing.T) {
	cfg := New().Add(Project, NewFileStore(t.TempDir()), "config")
	project := mustFrom(t, cfg, Project)

	if err := project.Set("server.host", "0.0.0.0"); err != nil {
		t.Fatalf("failed to set: %v", err)
	}
	if err := project.Set("server.port", 3000); err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	host, err := project.Get("server.host")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if host != "0.0.0.0" {
		t.Fatalf("want 0.0.0.0, got %v", host)
	}

	missing, err := project.Get("server.missing")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if missing != nil {
		t.Fatalf("want nil for missing path, got %v", missing)
	}
}

func Test_From_UnregisteredErrors(t *testing.T) {
	cfg := New().Add(Global, NewFileStore(t.TempDir()), "config")
	if _, err := cfg.From(Project); err == nil {
		t.Fatal("expected an error addressing an unregistered config")
	}
}

func Test_ExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

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
			if got := ExpandHome(tt.in); got != tt.want {
				t.Fatalf("want %q, got %q", tt.want, got)
			}
		})
	}
}
