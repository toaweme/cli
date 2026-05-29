package dev

import (
	"strings"
	"testing"

	"github.com/toaweme/cli"
)

func Test_DevCommand_Help(t *testing.T) {
	cmd := NewDevCommand(func() cli.Settings { return cli.Settings{} })
	if got := cmd.Help(); got != "Generate example outputs for all commands" {
		t.Fatalf("want %q, got %q", "Generate example outputs for all commands", got)
	}
}

func Test_cmdDir(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{name: "single", parts: []string{"db"}, want: "db"},
		{name: "nested", parts: []string{"db", "migrate"}, want: "db/migrate"},
		{name: "empty", parts: []string{}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmdDir(tt.parts); got != tt.want {
				t.Fatalf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func Test_buildRuns_BaseRuns(t *testing.T) {
	runs := buildRuns("full", nil)

	// the base runs always include help, the format variants, version,
	// completion, and the aggregated complete output
	wantNames := map[string]bool{
		"help":              false,
		"help-flags":        false,
		"format-md":         false,
		"format-plain":      false,
		"format-pretty":     false,
		"format-json":       false,
		"format-jsonschema": false,
		"version":           false,
		"completion":        false,
		"complete":          false,
	}

	for _, r := range runs {
		if _, ok := wantNames[r.name]; ok {
			wantNames[r.name] = true
		}
	}

	for name, seen := range wantNames {
		if !seen {
			t.Fatalf("expected base run %q to be present", name)
		}
	}
}

func Test_buildRuns_PerCommandRuns(t *testing.T) {
	runs := buildRuns("full", []string{"db migrate"})

	var found bool
	for _, r := range runs {
		if r.dir == "db/migrate" && r.name == "format-md" {
			found = true
			if len(r.args) == 0 || r.args[0] != "help" {
				t.Fatalf("expected help args, got %v", r.args)
			}
			if r.args[len(r.args)-1] != "--format=md" {
				t.Fatalf("expected --format=md suffix, got %v", r.args)
			}
		}
	}
	if !found {
		t.Fatalf("expected per-command run for %q", "db migrate")
	}
}

func Test_buildRuns_CompletionShells(t *testing.T) {
	runs := buildRuns("full", nil)

	for _, r := range runs {
		if r.name == "completion" {
			if len(r.completeArgs) != 3 {
				t.Fatalf("expected 3 completion shells, got %d", len(r.completeArgs))
			}
			return
		}
	}
	t.Fatalf("expected a completion run")
}

func Test_formatOutput(t *testing.T) {
	r := run{name: "help", args: []string{"--help"}, desc: "default help output"}
	out := formatOutput("greet", r, "hello world", 0)

	if !strings.Contains(out, "# greet: default help output") {
		t.Fatalf("expected heading in output, got %q", out)
	}
	if !strings.Contains(out, "❯ greet --help") {
		t.Fatalf("expected command line in output, got %q", out)
	}
	if !strings.Contains(out, "hello world") {
		t.Fatalf("expected captured output, got %q", out)
	}
	if strings.Contains(out, "exit code") {
		t.Fatalf("did not expect exit code for success, got %q", out)
	}
}

func Test_formatOutput_NonZeroExit(t *testing.T) {
	r := run{name: "help", args: []string{"bogus"}, desc: "bad command"}
	out := formatOutput("greet", r, "error: not found", 1)

	if !strings.Contains(out, "exit code: 1") {
		t.Fatalf("expected exit code in output, got %q", out)
	}
}
