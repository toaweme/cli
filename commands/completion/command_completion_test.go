package completion

import (
	"bytes"
	"io"
	"os"
	"strings"
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

func Test_CompletionCommand_Run(t *testing.T) {
	tests := []struct {
		name  string
		shell string
	}{
		{name: "bash", shell: "bash"},
		{name: "zsh", shell: "zsh"},
		{name: "fish", shell: "fish"},
		{name: "uppercase is normalized", shell: "BASH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCompletionCommand("myapp")
			cmd.Inputs = &CompletionConfig{Shell: tt.shell}

			out := captureStdout(t, func() {
				if err := cmd.Run(cli.GlobalOptions{}, cli.Unknowns{}); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})

			if out == "" {
				t.Fatalf("expected non-empty completion script")
			}
			if strings.Contains(out, "{{.AppName}}") {
				t.Fatalf("template placeholder was not replaced")
			}
			if !strings.Contains(out, "myapp") {
				t.Fatalf("expected app name %q in output", "myapp")
			}
		})
	}
}

func Test_CompletionCommand_Run_UnsupportedShell(t *testing.T) {
	cmd := NewCompletionCommand("myapp")
	cmd.Inputs = &CompletionConfig{Shell: "powershell"}

	err := cmd.Run(cli.GlobalOptions{}, cli.Unknowns{})
	if err == nil {
		t.Fatalf("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Fatalf("expected unsupported shell error, got %v", err)
	}
}

func Test_CompletionCommand_Run_NoShell(t *testing.T) {
	cmd := NewCompletionCommand("myapp")

	err := cmd.Run(cli.GlobalOptions{}, cli.Unknowns{})
	if err == nil {
		t.Fatalf("expected error when no shell provided")
	}
}

func Test_CompletionCommand_Examples(t *testing.T) {
	cmd := NewCompletionCommand("myapp")
	examples := cmd.Examples()

	if len(examples) != 3 {
		t.Fatalf("expected 3 examples, got %d", len(examples))
	}
	for _, ex := range examples {
		if !strings.Contains(ex, "myapp") {
			t.Fatalf("expected example to reference app name, got %q", ex)
		}
	}
}

func Test_CompletionCommand_Help(t *testing.T) {
	cmd := NewCompletionCommand("myapp")
	if got := cmd.Help(); got != "Generate shell completion scripts" {
		t.Fatalf("want %q, got %q", "Generate shell completion scripts", got)
	}
}
