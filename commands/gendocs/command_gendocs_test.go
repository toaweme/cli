package gendocs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toaweme/cli"
)

func Test_GenDocsCommand_Help(t *testing.T) {
	cmd := newTestCommand(t, nil)
	if got, want := cmd.Help(), "Generate documentation for all commands"; got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func Test_GenDocsCommand_Run_WritesFiles(t *testing.T) {
	dir := t.TempDir()
	cmd := newTestCommand(t, nil)
	cmd.Inputs = &Config{OutputDir: dir}

	if err := cmd.Run(cli.GlobalFlags{}, cli.Unknowns{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "testapp", "help.md")); err != nil {
		t.Fatalf("expected generated docs: %v", err)
	}
}

func newTestCommand(t *testing.T, commands []cli.Command[any]) *Command {
	t.Helper()
	return NewGenDocsCommand(
		func() cli.Config { return cli.Config{Name: "testapp"} },
		func() []cli.Command[any] { return commands },
		func() []cli.OutputCodec { return nil },
	)
}
