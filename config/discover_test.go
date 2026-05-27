package config

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_Discover(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(root string)
		start   func(root string) string
		names   []string
		wantHit bool
		wantDir string
	}{
		{
			name: "finds in current dir",
			setup: func(root string) {
				writeFile(t, filepath.Join(root, "config.json"), "{}")
			},
			start:   func(root string) string { return root },
			names:   []string{"config.json"},
			wantHit: true,
			wantDir: "",
		},
		{
			name: "finds walking up",
			setup: func(root string) {
				writeFile(t, filepath.Join(root, "app.yml"), "{}")
				os.MkdirAll(filepath.Join(root, "a", "b", "c"), 0o755)
			},
			start:   func(root string) string { return filepath.Join(root, "a", "b", "c") },
			names:   []string{"app.yml"},
			wantHit: true,
			wantDir: "",
		},
		{
			name:  "returns empty when not found",
			setup: func(root string) {},
			start: func(root string) string { return root },
			names: []string{"nope.json"},
		},
		{
			name: "respects name priority",
			setup: func(root string) {
				writeFile(t, filepath.Join(root, ".app.yml"), "hidden")
				writeFile(t, filepath.Join(root, "app.yml"), "visible")
			},
			start:   func(root string) string { return root },
			names:   []string{"app.yml", ".app.yml"},
			wantHit: true,
			wantDir: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(root)
			start := tt.start(root)

			result := Discover(start, tt.names)

			if !tt.wantHit {
				if result != "" {
					t.Fatalf("expected empty, got %s", result)
				}
				return
			}

			if result == "" {
				t.Fatal("expected a match, got empty")
			}

			if _, err := os.Stat(result); err != nil {
				t.Fatalf("discovered path does not exist: %s", result)
			}
		})
	}
}

func Test_HomePath(t *testing.T) {
	path := HomePath("myapp")
	if path == "" {
		t.Fatal("expected non-empty home path")
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".myapp")
	if path != expected {
		t.Fatalf("want %s, got %s", expected, path)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
