package cli

import (
	"testing"

	"github.com/toaweme/cli/config"
)

type mergeDB struct {
	Host string `json:"host" env:"DB_HOST" default:"db.local"`
	Port int    `json:"port" env:"DB_PORT" default:"5432"`
}

type mergeConfig struct {
	Database mergeDB `arg:"database" json:"database"`
	Region   string  `arg:"region" json:"region" env:"REGION" default:"us"`
}

type mergeCommand struct {
	BaseCommand[mergeConfig]
	got *mergeConfig
}

var _ Command[mergeConfig] = (*mergeCommand)(nil)

func (c *mergeCommand) Help() string { return "merge" }
func (c *mergeCommand) Run(_ GlobalFlags, _ Unknowns) error {
	*c.got = *c.Inputs
	return nil
}

// fileStore builds a single config store seeded with values, so a resolver over it merges them into Options().
func fileStore(t *testing.T, values map[string]any) config.Store {
	t.Helper()
	st := config.NewFileStore(t.TempDir(), "config", true)
	if err := st.Write(values); err != nil {
		t.Fatalf("failed to seed config store: %v", err)
	}
	return st
}

func runMerge(t *testing.T, resolver Resolver, args []string, env map[string]string) *mergeConfig {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
	got := &mergeConfig{}
	cmd := &mergeCommand{BaseCommand: NewBaseCommand[mergeConfig](), got: got}
	app := NewApp(Config{Name: "app"}, GlobalFlags{})
	if resolver != nil {
		app.Resolve(resolver)
	}
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)
	if err := app.Run(append([]string{"serve"}, args...)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return got
}

func Test_Resolve_NoResolver_DefaultsEnvFlags(t *testing.T) {
	// no resolver set -> only env and flags apply; flags overlaid on top.
	got := runMerge(t, nil, []string{"--region", "apac"}, nil)
	assertEqual(t, "db.local", got.Database.Host, "default applies with no files")
	assertEqual(t, 5432, got.Database.Port, "default applies with no files")
	assertEqual(t, "apac", got.Region, "flag overrides default")
}

func Test_Resolve_NoResolver_EnvBeatsDefault(t *testing.T) {
	got := runMerge(t, nil, nil, map[string]string{"REGION": "eu"})
	assertEqual(t, "eu", got.Region, "env overrides the struct default")
}

func Test_Resolve_FileResolver_Precedence(t *testing.T) {
	store := fileStore(t, map[string]any{
		"database": map[string]any{"host": "10.0.0.1", "port": 6543},
		"region":   "frankfurt",
	})
	got := runMerge(t, config.NewResolver(store, nil),
		[]string{"--database.host", "0.0.0.0"},
		map[string]string{"DB_PORT": "7000"},
	)
	assertEqual(t, "0.0.0.0", got.Database.Host, "flag beats config file")
	assertEqual(t, 7000, got.Database.Port, "env beats config file")
	assertEqual(t, "frankfurt", got.Region, "config file beats the struct default")
}

func Test_Resolve_FileResolver_PerCommandMapping(t *testing.T) {
	store := fileStore(t, map[string]any{
		"http": map[string]any{
			"location": "tokyo",
			"db":       map[string]any{"host": "10.0.0.5", "port": 5500},
		},
	})
	resolver := config.NewResolver(store, map[string]map[string]config.Source{
		"serve": {
			"region":   "http.location",
			"database": "http.db",
		},
	})
	got := runMerge(t, resolver, nil, nil)
	assertEqual(t, "tokyo", got.Region, "scalar field mapped from http.location")
	assertEqual(t, "10.0.0.5", got.Database.Host, "struct field mapped from http.db subtree")
	assertEqual(t, 5500, got.Database.Port, "struct field mapped from http.db subtree")
}

func Test_Resolve_FileResolver_FuncSource(t *testing.T) {
	store := fileStore(t, map[string]any{})
	resolver := config.NewResolver(store, map[string]map[string]config.Source{
		"serve": {"region": func() (any, error) { return "computed", nil }},
	})

	got := runMerge(t, resolver, nil, nil)
	assertEqual(t, "computed", got.Region, "func source computes the value")

	// a flag still wins over a mapped func value.
	got = runMerge(t, resolver, []string{"--region", "flagged"}, nil)
	assertEqual(t, "flagged", got.Region, "flag beats a mapped func value")
}
