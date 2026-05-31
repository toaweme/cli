package config

import (
	"os"
	"path/filepath"
	"testing"
)

type scopeConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func Test_Scope_WriteRead(t *testing.T) {
	dir := t.TempDir()
	cfg := New().Add(Global, NewFileStore(dir), "config")

	want := scopeConfig{Host: "localhost", Port: 8080}
	if err := cfg.Scope(Global).Write(want); err != nil {
		t.Fatalf("failed to write scope: %v", err)
	}

	var got scopeConfig
	if err := cfg.Scope(Global).Read(&got); err != nil {
		t.Fatalf("failed to read scope: %v", err)
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func Test_Scope_ReadMissingIsNoError(t *testing.T) {
	cfg := New().Add(Global, NewFileStore(t.TempDir()), "config")
	var got scopeConfig
	if err := cfg.Scope(Global).Read(&got); err != nil {
		t.Fatalf("reading a missing scope should not error: %v", err)
	}
	if got != (scopeConfig{}) {
		t.Fatalf("expected zero value, got %+v", got)
	}
}

func Test_Scope_SetGet(t *testing.T) {
	cfg := New().Add(Project, NewFileStore(t.TempDir()), "config")

	if err := cfg.Scope(Project).Set("server.host", "0.0.0.0"); err != nil {
		t.Fatalf("failed to set: %v", err)
	}
	if err := cfg.Scope(Project).Set("server.port", 3000); err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	host, err := cfg.Scope(Project).Get("server.host")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if host != "0.0.0.0" {
		t.Fatalf("want 0.0.0.0, got %v", host)
	}

	missing, err := cfg.Scope(Project).Get("server.missing")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if missing != nil {
		t.Fatalf("want nil for missing path, got %v", missing)
	}
}

func Test_Scope_UnregisteredErrors(t *testing.T) {
	cfg := New().Add(Global, NewFileStore(t.TempDir()), "config")
	if err := cfg.Scope(Project).Write(scopeConfig{}); err == nil {
		t.Fatal("expected an error addressing an unregistered scope")
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
