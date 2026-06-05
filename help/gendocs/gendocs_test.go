package gendocs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/toaweme/cli"
)

type serveConfig struct {
	Port int `arg:"port" short:"p" env:"PORT" help:"Port to listen on" default:"8080"`
}

type migrateConfig struct{}

func newCmd[T any](name string, subs ...cli.Command[any]) cli.Command[any] {
	cmd := &stubCommand[T]{BaseCommand: cli.NewBaseCommand[T]()}
	cmd.Name(name)
	for _, sub := range subs {
		cmd.Add(sub.Name(""), sub)
	}
	return cmd
}

type stubCommand[T any] struct {
	cli.BaseCommand[T]
}

func (c *stubCommand[T]) Run(_ cli.GlobalFlags, _ cli.Unknowns) error { return nil }
func (c *stubCommand[T]) Help() string                                { return "stub command" }

func sampleTree() []cli.Command[any] {
	return []cli.Command[any]{
		newCmd[serveConfig]("serve"),
		newCmd[migrateConfig]("db", newCmd[migrateConfig]("migrate")),
	}
}

func Test_Generate_WritesTreeFiles(t *testing.T) {
	dir := t.TempDir()

	written, err := Generate(Options{
		AppName:  "myapp",
		Commands: sampleTree(),
		Dir:      dir,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	want := []string{
		filepath.Join("myapp", "help.md"),
		filepath.Join("myapp", "help.txt"),
		filepath.Join("myapp", "help.json"),
		filepath.Join("myapp", "schema.json"),
	}
	for _, rel := range want {
		path := filepath.Join(dir, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected %q to be written: %v", rel, err)
		}
		if len(data) == 0 {
			t.Fatalf("expected %q to be non-empty", rel)
		}
	}

	// returned paths are relative to Dir
	for _, p := range written {
		if filepath.IsAbs(p) {
			t.Fatalf("expected relative path, got %q", p)
		}
	}

	md, _ := os.ReadFile(filepath.Join(dir, "myapp", "help.md"))
	if !strings.Contains(string(md), "## serve") {
		t.Fatalf("expected serve command in markdown docs, got:\n%s", md)
	}
	if !strings.Contains(string(md), "--port") {
		t.Fatalf("expected port flag in markdown docs, got:\n%s", md)
	}
}

func Test_Generate_PerCommandFiles(t *testing.T) {
	dir := t.TempDir()

	_, err := Generate(Options{
		AppName:    "myapp",
		Commands:   sampleTree(),
		Dir:        dir,
		PerCommand: true,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	for _, rel := range []string{
		filepath.Join("myapp", "commands", "serve", "help.md"),
		filepath.Join("myapp", "commands", "db", "help.md"),
		filepath.Join("myapp", "commands", "db", "migrate", "help.md"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected per-command file %q: %v", rel, err)
		}
	}
}

func Test_Generate_RequiresAppName(t *testing.T) {
	if _, err := Generate(Options{Dir: t.TempDir()}); err == nil {
		t.Fatal("expected error when app name is empty")
	}
}

func Test_commandPaths(t *testing.T) {
	paths := commandPaths(sampleTree(), nil)

	var got []string
	for _, p := range paths {
		got = append(got, strings.Join(p, " "))
	}
	sort.Strings(got)

	want := []string{"db", "db migrate", "serve"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("want %v, got %v", want, got)
	}
}
