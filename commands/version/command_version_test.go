package version

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/toaweme/cli"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	os.Stdout = orig
	return <-done
}

func Test_VersionCommand_Run(t *testing.T) {
	cmd := NewVersionCommand(func() cli.Config {
		return cli.Config{Name: "myapp", Version: "1.2.3"}
	})

	out := captureStdout(t, func() {
		if err := cmd.Run(cli.GlobalOptions{}, cli.Unknowns{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	want := "myapp 1.2.3\n"
	if out != want {
		t.Fatalf("want %q, got %q", want, out)
	}
}

func Test_VersionCommand_Validate(t *testing.T) {
	cmd := NewVersionCommand(func() cli.Config { return cli.Config{} })
	if err := cmd.Validate(map[string]any{"anything": "goes"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_VersionCommand_Help(t *testing.T) {
	cmd := NewVersionCommand(func() cli.Config { return cli.Config{} })
	if got := cmd.Help(); got != "Display version" {
		t.Fatalf("want %q, got %q", "Display version", got)
	}
}
