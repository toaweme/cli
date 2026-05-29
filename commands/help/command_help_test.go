package help

import (
	"bytes"
	"errors"
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

type stubCommand struct {
	cli.BaseCommand[any]
	help string
}

var _ cli.Command[any] = (*stubCommand)(nil)

func (s *stubCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error { return nil }
func (s *stubCommand) Help() string                                  { return s.help }

func newStub(name, help string) cli.Command[any] {
	cmd := &stubCommand{help: help}
	cmd.Name(name)
	return cmd
}

func newHelpCommand() *HelpCommand {
	settings := func() cli.Config { return cli.Config{Name: "myapp", Version: "1.0.0"} }
	commands := func() []cli.Command[any] {
		return []cli.Command[any]{
			newStub("build", "Build the project"),
			newStub("serve", "Start the server"),
		}
	}
	return NewHelpCommand(settings, commands)
}

func Test_HelpCommand_Run_Formats(t *testing.T) {
	formats := []string{"", "plain", "md", "pretty", "plain-flags", "json", "jsonschema"}

	for _, format := range formats {
		name := format
		if name == "" {
			name = "default"
		}
		t.Run(name, func(t *testing.T) {
			cmd := newHelpCommand()
			out := captureStdout(t, func() {
				err := cmd.Run(cli.GlobalOptions{Format: format}, cli.Unknowns{})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
			if out == "" {
				t.Fatalf("expected non-empty help output for format %q", format)
			}
		})
	}
}

func Test_HelpCommand_Run_FilteredByArgs(t *testing.T) {
	cmd := newHelpCommand()
	out := captureStdout(t, func() {
		err := cmd.Run(cli.GlobalOptions{}, cli.Unknowns{Args: []string{"build"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if out == "" {
		t.Fatalf("expected non-empty filtered help output")
	}
}

func Test_HelpCommand_Validate(t *testing.T) {
	cmd := newHelpCommand()
	if err := cmd.Validate(map[string]any{"anything": "goes"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_HelpCommand_Help(t *testing.T) {
	cmd := newHelpCommand()
	if got := cmd.Help(); got != "Display help" {
		t.Fatalf("want %q, got %q", "Display help", got)
	}
}

func Test_ParentCommand_Run_DisplaysSubCommands(t *testing.T) {
	parent := NewParentPlaceholder()
	err := parent.Run(cli.GlobalOptions{}, cli.Unknowns{})
	if !errors.Is(err, cli.ErrDisplaySubCommands) {
		t.Fatalf("expected ErrDisplaySubCommands, got %v", err)
	}
}

func Test_ParentCommand_Help_IsEmpty(t *testing.T) {
	parent := NewParentPlaceholder()
	if got := parent.Help(); got != "" {
		t.Fatalf("expected empty help, got %q", got)
	}
}

func Test_ParentCommand_Add_RegistersSubCommand(t *testing.T) {
	parent := NewParentPlaceholder()
	parent.Add("child", newStub("child", "a child"))

	subs := parent.Commands()
	if len(subs) != 1 {
		t.Fatalf("expected 1 subcommand, got %d", len(subs))
	}
	if subs[0].Name("") != "child" {
		t.Fatalf("expected subcommand name %q, got %q", "child", subs[0].Name(""))
	}
}
