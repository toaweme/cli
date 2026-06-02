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

func Test_Store_WriteRead(t *testing.T) {
	store := NewFileStore(t.TempDir(), "config")

	want := sampleConfig{Host: "localhost", Port: 8080}
	if err := store.Write(want); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	var got sampleConfig
	if err := store.Read(&got); err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func Test_Store_ReadMissingIsNoError(t *testing.T) {
	store := NewFileStore(t.TempDir(), "config")
	var got sampleConfig
	if err := store.Read(&got); err != nil {
		t.Fatalf("reading a missing file should not error: %v", err)
	}
	if got != (sampleConfig{}) {
		t.Fatalf("expected zero value, got %+v", got)
	}
}

func Test_Store_KeyWriteReadExists(t *testing.T) {
	store := NewFileStore(t.TempDir(), "config")

	if store.KeyExists("server.host") {
		t.Fatal("missing key should not exist")
	}

	if err := store.KeyWrite("server.host", "0.0.0.0"); err != nil {
		t.Fatalf("failed to set: %v", err)
	}
	if err := store.KeyWrite("server.port", 3000); err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	if !store.KeyExists("server.host") {
		t.Fatal("written key should exist")
	}

	host, err := store.KeyRead("server.host")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if host != "0.0.0.0" {
		t.Fatalf("want 0.0.0.0, got %v", host)
	}

	// the second key write must not clobber the first.
	port, err := store.KeyRead("server.port")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if port != float64(3000) && port != 3000 {
		t.Fatalf("want 3000, got %v", port)
	}

	missing, err := store.KeyRead("server.missing")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if missing != nil {
		t.Fatalf("want nil for missing path, got %v", missing)
	}

	// deleting a key leaves its siblings intact.
	if err := store.KeyDelete("server.host"); err != nil {
		t.Fatalf("failed to delete key: %v", err)
	}
	if store.KeyExists("server.host") {
		t.Fatal("deleted key should not exist")
	}
	if !store.KeyExists("server.port") {
		t.Fatal("sibling key should survive the delete")
	}

	// deleting an absent key (or from a missing file) is not an error.
	if err := store.KeyDelete("server.nope"); err != nil {
		t.Fatalf("deleting an absent key should not error: %v", err)
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
