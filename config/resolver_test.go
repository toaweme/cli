package config

import (
	"testing"
)

func Test_NewFileResolver_LayersScopesLowToHigh(t *testing.T) {
	global := NewFileStore(t.TempDir())
	project := NewFileStore(t.TempDir())
	if err := global.Save("config", map[string]any{"host": "global", "region": "eu"}); err != nil {
		t.Fatalf("seed global: %v", err)
	}
	// project overrides host, leaves region to the global layer
	if err := project.Save("config", map[string]any{"host": "project"}); err != nil {
		t.Fatalf("seed project: %v", err)
	}

	r := NewFileResolver(New().
		Add(Global, global, "config").
		Add(Project, project, "config"), nil)

	values, err := r.Resolve("serve", nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if values["host"] != "project" {
		t.Fatalf("project scope should override global, got %v", values["host"])
	}
	if values["region"] != "eu" {
		t.Fatalf("global scope should remain where project does not set it, got %v", values["region"])
	}
}

func Test_NewFileResolver_MapPathAndFunc(t *testing.T) {
	store := NewFileStore(t.TempDir())
	if err := store.Save("config", map[string]any{
		"http": map[string]any{"location": "tokyo"},
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	r := NewFileResolver(New().Add(Global, store, "config"), map[string]map[string]Source{
		"serve": {
			"region":   "http.location",
			"computed": func() (any, error) { return 42, nil },
		},
	})

	values, err := r.Resolve("serve", nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if values["region"] != "tokyo" {
		t.Fatalf("path source should resolve http.location, got %v", values["region"])
	}
	if values["computed"] != 42 {
		t.Fatalf("func source should compute the value, got %v", values["computed"])
	}

	// rules only apply to the matching command
	other, err := r.Resolve("build", nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if _, ok := other["region"]; ok {
		t.Fatal("mapping for serve must not leak into build")
	}
}
