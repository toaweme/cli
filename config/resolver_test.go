package config

import (
	"testing"
)

func Test_Resolvers_LayerStoresLowToHigh(t *testing.T) {
	global := NewFileStore(t.TempDir(), "config")
	project := NewFileStore(t.TempDir(), "config")
	if err := global.Write(map[string]any{"host": "global", "region": "eu"}); err != nil {
		t.Fatalf("seed global: %v", err)
	}
	// project overrides host, leaves region to the global layer
	if err := project.Write(map[string]any{"host": "project"}); err != nil {
		t.Fatalf("seed project: %v", err)
	}

	// run global then project, threading the accumulated values like the App does.
	values, err := NewResolver(global, nil).Resolve("serve", nil)
	if err != nil {
		t.Fatalf("resolve global: %v", err)
	}
	values, err = NewResolver(project, nil).Resolve("serve", values)
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if values["host"] != "project" {
		t.Fatalf("project store should override global, got %v", values["host"])
	}
	if values["region"] != "eu" {
		t.Fatalf("global store should remain where project does not set it, got %v", values["region"])
	}
}

func Test_Resolver_MapPathAndFunc(t *testing.T) {
	store := NewFileStore(t.TempDir(), "config")
	if err := store.Write(map[string]any{
		"http": map[string]any{"location": "tokyo"},
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	r := NewResolver(store, map[string]map[string]Source{
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
