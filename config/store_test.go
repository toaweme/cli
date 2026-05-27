package config

import (
	"os"
	"path/filepath"
	"testing"
)

type testCfg struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func Test_FileStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	cfg := testCfg{Name: "app", Port: 8080}
	if err := store.Save("test", cfg); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	var loaded testCfg
	if err := store.Load("test", &loaded); err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if loaded != cfg {
		t.Fatalf("want %+v, got %+v", cfg, loaded)
	}
}

func Test_FileStore_DefaultJSON(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	store.Save("cfg", testCfg{Name: "test"})

	// should create .json by default
	if _, err := os.Stat(filepath.Join(dir, "cfg.json")); err != nil {
		t.Fatal("expected cfg.json to exist")
	}
}

func Test_FileStore_ExplicitExtension(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	store.Save("app.json", testCfg{Name: "explicit"})

	// should not double the extension
	if _, err := os.Stat(filepath.Join(dir, "app.json")); err != nil {
		t.Fatal("expected app.json to exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "app.json.json")); !os.IsNotExist(err) {
		t.Fatal("should not double extension")
	}
}

func Test_FileStore_Exists(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	if store.Exists("nope") {
		t.Fatal("expected missing key to not exist")
	}

	store.Save("yep", testCfg{Name: "x"})
	if !store.Exists("yep") {
		t.Fatal("expected saved key to exist")
	}
}

func Test_FileStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	store.Save("del", testCfg{Name: "gone"})
	if err := store.Delete("del"); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
	if store.Exists("del") {
		t.Fatal("expected deleted key to not exist")
	}
}

func Test_FileStore_DeleteNonExistent(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	if err := store.Delete("missing"); err != nil {
		t.Fatalf("deleting non-existent should not error: %v", err)
	}
}

func Test_FileStore_ConfigPermissions(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	store.Save("cfg", testCfg{Name: "public"})

	info, err := os.Stat(filepath.Join(dir, "cfg.json"))
	if err != nil {
		t.Fatalf("failed to stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Fatalf("want 0644, got %04o", perm)
	}
}

func Test_FileStore_SecretPermissions(t *testing.T) {
	dir := t.TempDir()
	store := NewSecretFileStore(dir)

	store.Save("secret", testCfg{Name: "token"})

	info, err := os.Stat(filepath.Join(dir, "secret.json"))
	if err != nil {
		t.Fatalf("failed to stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("want 0600, got %04o", perm)
	}
}

func Test_FileStore_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	store.Save("atomic", testCfg{Name: "v1"})
	store.Save("atomic", testCfg{Name: "v2"})

	tmp := filepath.Join(dir, "atomic.json.tmp")
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Fatal("tmp file should not remain after write")
	}

	var loaded testCfg
	store.Load("atomic", &loaded)
	if loaded.Name != "v2" {
		t.Fatalf("want v2, got %s", loaded.Name)
	}
}

func Test_FileStore_NestedKey(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	cfg := testCfg{Name: "deep"}
	if err := store.Save("a/b/c", cfg); err != nil {
		t.Fatalf("failed to save nested: %v", err)
	}

	if !store.Exists("a/b/c") {
		t.Fatal("expected nested key to exist")
	}

	var loaded testCfg
	store.Load("a/b/c", &loaded)
	if loaded.Name != "deep" {
		t.Fatalf("want deep, got %s", loaded.Name)
	}
}

func Test_FileStore_Dir(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	if store.Dir() != dir {
		t.Fatalf("want %s, got %s", dir, store.Dir())
	}
}

// mockCodec pretends to be a YAML codec for testing multi-codec support.
type mockCodec struct {
	jsonCodec
}

func (c *mockCodec) Extension() string { return ".yaml" }

func Test_FileStore_AddCodec(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir).AddCodec(&mockCodec{})

	// default should now be the last added codec
	store.Save("cfg", testCfg{Name: "yaml"})
	if _, err := os.Stat(filepath.Join(dir, "cfg.yaml")); err != nil {
		t.Fatal("expected cfg.yaml (new default)")
	}

	// explicit .json still uses json codec
	store.Save("other.json", testCfg{Name: "json"})
	if _, err := os.Stat(filepath.Join(dir, "other.json")); err != nil {
		t.Fatal("expected other.json")
	}
}

func Test_FileStore_SetDefault(t *testing.T) {
	dir := t.TempDir()
	mock := &mockCodec{}
	store := NewFileStore(dir).AddCodec(mock).SetDefault(&jsonCodec{})

	// default is json again after SetDefault
	store.Save("cfg", testCfg{Name: "test"})
	if _, err := os.Stat(filepath.Join(dir, "cfg.json")); err != nil {
		t.Fatal("expected cfg.json after SetDefault to json")
	}

	// explicit .yaml still works
	store.Save("other.yaml", testCfg{Name: "test"})
	if _, err := os.Stat(filepath.Join(dir, "other.yaml")); err != nil {
		t.Fatal("expected other.yaml via explicit extension")
	}
}

func Test_FileStore_MultiCodecRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir).AddCodec(&mockCodec{})

	cfg := testCfg{Name: "multi", Port: 9090}

	// save as yaml (default after AddCodec)
	store.Save("app", cfg)

	// save as json (explicit)
	store.Save("app.json", cfg)

	// both should load correctly
	var fromYaml, fromJSON testCfg
	if err := store.Load("app", &fromYaml); err != nil {
		t.Fatalf("failed to load yaml: %v", err)
	}
	if err := store.Load("app.json", &fromJSON); err != nil {
		t.Fatalf("failed to load json: %v", err)
	}

	if fromYaml != cfg {
		t.Fatalf("yaml roundtrip: want %+v, got %+v", cfg, fromYaml)
	}
	if fromJSON != cfg {
		t.Fatalf("json roundtrip: want %+v, got %+v", cfg, fromJSON)
	}
}
