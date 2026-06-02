package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type testCfg struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func Test_FileStore_WriteAndRead(t *testing.T) {
	store := NewFileStore(t.TempDir(), "test")

	cfg := testCfg{Name: "app", Port: 8080}
	if err := store.Write(cfg); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	var loaded testCfg
	if err := store.Read(&loaded); err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if loaded != cfg {
		t.Fatalf("want %+v, got %+v", cfg, loaded)
	}
}

func Test_FileStore_DefaultJSON(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir, "cfg")

	store.Write(testCfg{Name: "test"})

	// should create .json by default
	if _, err := os.Stat(filepath.Join(dir, "cfg.json")); err != nil {
		t.Fatal("expected cfg.json to exist")
	}
}

func Test_FileStore_ExplicitExtension(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir, "app.json")

	store.Write(testCfg{Name: "explicit"})

	// should not double the extension
	if _, err := os.Stat(filepath.Join(dir, "app.json")); err != nil {
		t.Fatal("expected app.json to exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "app.json.json")); !os.IsNotExist(err) {
		t.Fatal("should not double extension")
	}
}

func Test_FileStore_Exists(t *testing.T) {
	store := NewFileStore(t.TempDir(), "yep")

	if store.Exists() {
		t.Fatal("expected missing file to not exist")
	}

	store.Write(testCfg{Name: "x"})
	if !store.Exists() {
		t.Fatal("expected written file to exist")
	}
}

func Test_FileStore_Delete(t *testing.T) {
	store := NewFileStore(t.TempDir(), "del")

	store.Write(testCfg{Name: "gone"})
	if err := store.Delete(); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
	if store.Exists() {
		t.Fatal("expected deleted file to not exist")
	}
}

func Test_FileStore_DeleteNonExistent(t *testing.T) {
	store := NewFileStore(t.TempDir(), "missing")

	if err := store.Delete(); err != nil {
		t.Fatalf("deleting non-existent should not error: %v", err)
	}
}

func Test_FileStore_ConfigPermissions(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir, "cfg")

	store.Write(testCfg{Name: "public"})

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
	store := FileSecrets(dir)

	store.Write(testCfg{Name: "token"})

	info, err := os.Stat(filepath.Join(dir, "secrets.json"))
	if err != nil {
		t.Fatalf("failed to stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("want 0600, got %04o", perm)
	}
}

func Test_FileStore_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir, "atomic")

	store.Write(testCfg{Name: "v1"})
	store.Write(testCfg{Name: "v2"})

	tmp := filepath.Join(dir, "atomic.json.tmp")
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Fatal("tmp file should not remain after write")
	}

	var loaded testCfg
	store.Read(&loaded)
	if loaded.Name != "v2" {
		t.Fatalf("want v2, got %s", loaded.Name)
	}
}

func Test_FileStore_NestedName(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir, "a/b/c")

	cfg := testCfg{Name: "deep"}
	if err := store.Write(cfg); err != nil {
		t.Fatalf("failed to write nested: %v", err)
	}

	if !store.Exists() {
		t.Fatal("expected nested file to exist")
	}

	var loaded testCfg
	store.Read(&loaded)
	if loaded.Name != "deep" {
		t.Fatalf("want deep, got %s", loaded.Name)
	}
}

func Test_FileStore_Dir(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir, "config")

	if store.Dir() != dir {
		t.Fatalf("want %s, got %s", dir, store.Dir())
	}
}

// mockCodec is a single-extension non-JSON codec (JSON-encoded payload) used to
// check that the store honors the codec it was given rather than the default.
type mockCodec struct{}

func (c *mockCodec) Marshal(v any) ([]byte, error)      { return json.Marshal(v) }
func (c *mockCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
func (c *mockCodec) Extension() string                  { return ".yaml" }

func Test_FileStore_CodecExtensionAppended(t *testing.T) {
	dir := t.TempDir()
	// extension-less name gets the codec's extension appended.
	store := NewFileStore(dir, "cfg", &mockCodec{})

	if err := store.Write(testCfg{Name: "yaml"}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cfg.yaml")); err != nil {
		t.Fatal("expected cfg.yaml (codec extension appended)")
	}

	var got testCfg
	if err := store.Read(&got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Name != "yaml" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func Test_FileStore_ExplicitExtensionName(t *testing.T) {
	dir := t.TempDir()
	// a name carrying its own extension is used verbatim, no doubling.
	store := NewFileStore(dir, "data.yaml", &mockCodec{})

	if err := store.Write(testCfg{Name: "x"}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "data.yaml")); err != nil {
		t.Fatal("expected data.yaml written verbatim")
	}
	if _, err := os.Stat(filepath.Join(dir, "data.yaml.yaml")); !os.IsNotExist(err) {
		t.Fatal("should not append a second extension")
	}
}
