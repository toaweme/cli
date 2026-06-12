package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type sampleConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func Test_Store_WriteRead(t *testing.T) {
	store := NewFileStore(t.TempDir(), "config", true)

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

func Test_Store_ReadMissingReturnsErrConfigNotFound(t *testing.T) {
	store := NewFileStore(t.TempDir(), "config", true)
	var got sampleConfig
	if err := store.Read(&got); !errors.Is(err, ErrConfigNotFound) {
		t.Fatalf("reading a missing file should return ErrConfigNotFound, got %v", err)
	}
}

func Test_Store_KeyWriteReadExists(t *testing.T) {
	store := NewFileStore(t.TempDir(), "config", true)

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

	// reading an absent key in an existing file is ErrKeyNotFound, distinct from a present null.
	if _, err := store.KeyRead("server.missing"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("want ErrKeyNotFound for absent key, got %v", err)
	}

	// reading any key from a non-existent file is ErrConfigNotFound.
	absent := NewFileStore(t.TempDir(), "nope", true)
	if _, err := absent.KeyRead("server.host"); !errors.Is(err, ErrConfigNotFound) {
		t.Fatalf("want ErrConfigNotFound for missing file, got %v", err)
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
