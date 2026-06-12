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
			cmd.Inputs = &Config{Shell: tt.shell}

			out := captureStdout(t, func() {
				if err := cmd.Run(cli.GlobalFlags{}, cli.Unknowns{}); err != nil {
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
	cmd.Inputs = &Config{Shell: "powershell"}

	err := cmd.Run(cli.GlobalFlags{}, cli.Unknowns{})
	if err == nil {
		t.Fatalf("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Fatalf("expected unsupported shell error, got %v", err)
	}
}

func Test_CompletionCommand_Run_NoShell(t *testing.T) {
	cmd := NewCompletionCommand("myapp")

	err := cmd.Run(cli.GlobalFlags{}, cli.Unknowns{})
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
		if len(ex) == 0 {
			t.Fatalf("expected non-empty example lines")
		}
		if !strings.Contains(ex[0], "myapp") {
			t.Fatalf("expected example to reference app name, got %q", ex[0])
		}
	}
}

func Test_CompletionCommand_Args(t *testing.T) {
	cmd := NewCompletionCommand("myapp")
	args := cmd.Args()

	lines, ok := args[0]
	if !ok {
		t.Fatalf("expected a description for positional arg 0")
	}
	if len(lines) < 2 {
		t.Fatalf("expected a multi-line description, got %d line(s)", len(lines))
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"bash", "zsh", "fish"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected arg description to mention %q, got %q", want, joined)
		}
	}
}

func Test_CompletionCommand_Help(t *testing.T) {
	cmd := NewCompletionCommand("myapp")
	if got := cmd.Help(); got != "Generate shell completion scripts" {
		t.Fatalf("want %q, got %q", "Generate shell completion scripts", got)
	}
}

func Test_CompletionCommand_Description(t *testing.T) {
	cmd := NewCompletionCommand("myapp")
	desc := cmd.Description()

	if !strings.Contains(desc, "\n") {
		t.Fatalf("expected a multi-line description, got %q", desc)
	}
	for _, want := range []string{"Install", "myapp completion bash", "myapp completion zsh", "myapp completion fish"} {
		if !strings.Contains(desc, want) {
			t.Fatalf("description missing %q in:\n%s", want, desc)
		}
	}
}
