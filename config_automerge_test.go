package cli

import (
	"testing"

	"github.com/toaweme/cli/config"
)

type autoMergeDB struct {
	Host string `json:"host" env:"HOST" default:"db.local"`
	Port int    `json:"port" env:"PORT" default:"5432"`
}

type autoMergeConfig struct {
	Database autoMergeDB `arg:"database" json:"database" env:"DATABASE"`
	Region   string      `arg:"region" json:"region" env:"REGION" default:"us"`
}

type autoMergeCommand struct {
	BaseCommand[autoMergeConfig]
	got      *autoMergeConfig
	strategy MergeStrategy
	mapping  ConfigMapping
}

var _ Command[autoMergeConfig] = (*autoMergeCommand)(nil)

func (c *autoMergeCommand) Help() string { return "auto-merge" }
func (c *autoMergeCommand) Run(_ GlobalOptions, _ Unknowns) error {
	*c.got = *c.Inputs
	return nil
}

// ConfigStrategy returns the command's override; the zero value MergeInherit
// means "use the app default", matching the BaseCommand behavior.
func (c *autoMergeCommand) ConfigStrategy() (MergeStrategy, ConfigMapping) {
	return c.strategy, c.mapping
}

// storeWith builds a single-store file-backed Storage seeded with values under
// the "config" key, so app.Run merges them into the command's Options().
func storeWith(t *testing.T, values map[string]any) Storage {
	t.Helper()
	dir := t.TempDir()
	st := config.NewFileStore(dir)
	if err := st.Save("config", values); err != nil {
		t.Fatalf("failed to seed config store: %v", err)
	}
	return &storage{store: st, dir: dir, stores: []config.Store{st}}
}

func Test_App_AutoMerge_ConfigEnvFlags(t *testing.T) {
	store := storeWith(t, map[string]any{
		"database": map[string]any{"host": "127.0.0.1", "port": 6543},
		"region":   "eu",
	})

	got := &autoMergeConfig{}
	cmd := &autoMergeCommand{BaseCommand: NewBaseCommand[autoMergeConfig](), got: got}

	// app-wide default opts every command into layered merge
	app := NewApp(Config{Name: "app", Store: store, Merge: MergeLayered}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	// env overrides the config file for the port; a flag overrides the host.
	t.Setenv("DATABASE_PORT", "7000")

	err := app.Run([]string{"serve", "--database.host", "0.0.0.0"})
	assertNoError(t, err)

	assertEqual(t, "0.0.0.0", got.Database.Host, "flag beats config file")
	assertEqual(t, 7000, got.Database.Port, "env beats config file")
	assertEqual(t, "eu", got.Region, "config file beats the struct default")
}

func Test_App_Merge_PerCommandOptOut(t *testing.T) {
	store := storeWith(t, map[string]any{"region": "eu"})

	got := &autoMergeConfig{}
	// command overrides the app-wide MergeLayered default, opting out
	cmd := &autoMergeCommand{
		BaseCommand: NewBaseCommand[autoMergeConfig](),
		got:         got,
		strategy:    MergeEnvFlags,
	}

	app := NewApp(Config{Name: "app", Store: store, Merge: MergeLayered}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	err := app.Run([]string{"serve"})
	assertNoError(t, err)

	// the config store is ignored for this command; region keeps its default
	assertEqual(t, "us", got.Region, "opted-out command ignores the config store")
}

func Test_App_Merge_FieldMapping(t *testing.T) {
	// global config nests the values under http: differently than the command
	// struct, which has flat fields tagged region and a database block.
	store := storeWith(t, map[string]any{
		"http": map[string]any{
			"location": "frankfurt",
			"db":       map[string]any{"host": "10.0.0.5", "port": 5500},
		},
	})

	got := &autoMergeConfig{}
	cmd := &autoMergeCommand{
		BaseCommand: NewBaseCommand[autoMergeConfig](),
		got:         got,
		strategy:    MergeLayered,
		mapping: ConfigMapping{
			"region":   "http.location",
			"database": "http.db",
		},
	}

	app := NewApp(Config{Name: "app", Store: store}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	err := app.Run([]string{"serve"})
	assertNoError(t, err)

	assertEqual(t, "frankfurt", got.Region, "scalar field mapped from http.location")
	assertEqual(t, "10.0.0.5", got.Database.Host, "struct field mapped from http.db subtree")
	assertEqual(t, 5500, got.Database.Port, "struct field mapped from http.db subtree")
}

func Test_App_Merge_NamespacedWildcard(t *testing.T) {
	// the whole command config lives under an "http:" namespace in global config
	store := storeWith(t, map[string]any{
		"http": map[string]any{
			"region":   "frankfurt",
			"database": map[string]any{"host": "10.0.0.9", "port": 5400},
		},
	})

	got := &autoMergeConfig{}
	cmd := &autoMergeCommand{
		BaseCommand: NewBaseCommand[autoMergeConfig](),
		got:         got,
		strategy:    MergeLayered,
		mapping:     Namespaced("http"),
	}

	app := NewApp(Config{Name: "app", Store: store}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	err := app.Run([]string{"serve"})
	assertNoError(t, err)

	assertEqual(t, "frankfurt", got.Region, "every field sourced from under http.")
	assertEqual(t, "10.0.0.9", got.Database.Host, "nested struct field sourced from http.database")
	assertEqual(t, 5400, got.Database.Port, "nested struct field sourced from http.database")
}

func Test_App_Merge_WildcardWithExplicitOverride(t *testing.T) {
	store := storeWith(t, map[string]any{
		"http": map[string]any{
			"database": map[string]any{"host": "10.0.0.9", "port": 5400},
		},
		"location": "tokyo",
	})

	got := &autoMergeConfig{}
	cmd := &autoMergeCommand{
		BaseCommand: NewBaseCommand[autoMergeConfig](),
		got:         got,
		strategy:    MergeLayered,
		// default everything under http., but pull region from a top-level key
		mapping: ConfigMapping{"*": "http.*", "region": "location"},
	}

	app := NewApp(Config{Name: "app", Store: store}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	err := app.Run([]string{"serve"})
	assertNoError(t, err)

	assertEqual(t, "tokyo", got.Region, "explicit entry overrides the wildcard")
	assertEqual(t, "10.0.0.9", got.Database.Host, "wildcard still applies to other fields")
}

func Test_App_Merge_CommandNameNamespaceDefault(t *testing.T) {
	// shared top-level region, plus a per-command "serve:" section that overrides
	store := storeWith(t, map[string]any{
		"region": "shared-eu",
		"serve":  map[string]any{"region": "serve-us"},
	})

	got := &autoMergeConfig{}
	cmd := &autoMergeCommand{
		BaseCommand: NewBaseCommand[autoMergeConfig](),
		got:         got,
		strategy:    MergeLayered,
		// no explicit mapping: the command name becomes the override namespace
	}

	app := NewApp(Config{Name: "app", Store: store}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	err := app.Run([]string{"serve"})
	assertNoError(t, err)

	assertEqual(t, "serve-us", got.Region, "the command's own section overrides shared top-level")
}

func Test_App_Merge_CommandNameNamespaceFallsBackToShared(t *testing.T) {
	// only shared top-level config; no "serve:" section
	store := storeWith(t, map[string]any{"region": "shared-eu"})

	got := &autoMergeConfig{}
	cmd := &autoMergeCommand{
		BaseCommand: NewBaseCommand[autoMergeConfig](),
		got:         got,
		strategy:    MergeLayered,
	}

	app := NewApp(Config{Name: "app", Store: store}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	err := app.Run([]string{"serve"})
	assertNoError(t, err)

	assertEqual(t, "shared-eu", got.Region, "falls back to shared top-level when no command section exists")
}

func Test_App_BindsConfigToCommands(t *testing.T) {
	store := storeWith(t, map[string]any{})
	cmd := &autoMergeCommand{BaseCommand: NewBaseCommand[autoMergeConfig](), got: &autoMergeConfig{}}

	app := NewApp(Config{Name: "myapp", Version: "1.2.3", Store: store}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	err := app.Run([]string{"serve"})
	assertNoError(t, err)

	// the command can read the global config without any constructor wiring
	assertEqual(t, "myapp", cmd.Config().Name, "command sees the app config name")
	assertEqual(t, "1.2.3", cmd.Config().Version, "command sees the app config version")
	assertTrue(t, cmd.Store() != nil, "command sees the configured store")
}

func Test_App_Merge_DefaultIsEnvFlagsWhenUnset(t *testing.T) {
	store := storeWith(t, map[string]any{"region": "eu"})

	got := &autoMergeConfig{}
	cmd := &autoMergeCommand{BaseCommand: NewBaseCommand[autoMergeConfig](), got: got}

	// Store is set but Merge is left unset: config store must NOT be read
	app := NewApp(Config{Name: "app", Store: store}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	err := app.Run([]string{"serve"})
	assertNoError(t, err)

	assertEqual(t, "us", got.Region, "unset Merge means env+flags only, even with a Store")
}

func Test_App_AutoMerge_NoStore_DefaultsAndFlags(t *testing.T) {
	got := &autoMergeConfig{}
	cmd := &autoMergeCommand{BaseCommand: NewBaseCommand[autoMergeConfig](), got: got}

	// no Store: defaults apply, flags override, no config file involved.
	app := NewApp(Config{Name: "app"}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Add("serve", cmd)

	err := app.Run([]string{"serve", "--region", "apac"})
	assertNoError(t, err)

	assertEqual(t, "db.local", got.Database.Host, "default applies with no store")
	assertEqual(t, 5432, got.Database.Port, "default applies with no store")
	assertEqual(t, "apac", got.Region, "flag overrides the default")
}
