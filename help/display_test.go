package help

import (
	"bytes"
	"encoding/json"
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

type testFlags struct {
	Name    string `arg:"name" short:"n" help:"the name to use" env:"NAME" rules:"required"`
	Verbose bool   `arg:"verbose" help:"enable verbose output"`
	Pos     string `arg:"0" help:"a positional argument"`
}

type flagStub struct {
	cli.BaseCommand[testFlags]
	help string
}

var _ cli.Command[testFlags] = (*flagStub)(nil)

func (s *flagStub) Run(_ cli.GlobalOptions, _ cli.Unknowns) error { return nil }
func (s *flagStub) Help() string                                  { return s.help }

func newFlagStub(name, help string, subs ...cli.Command[any]) cli.Command[any] {
	cmd := &flagStub{BaseCommand: cli.NewBaseCommand[testFlags](), help: help}
	cmd.Name(name)
	for _, sub := range subs {
		cmd.Add(sub.Name(""), sub)
	}
	return cmd
}

func commandTree() []cli.Command[any] {
	migrate := newFlagStub("migrate", "Run database migrations")
	db := newFlagStub("db", "Database commands", migrate)
	build := newFlagStub("build", "Build the project")
	return []cli.Command[any]{build, db}
}

// descStub is a command that provides a multi-line Description.
type descStub struct {
	cli.BaseCommand[testFlags]
	help string
	desc string
}

var _ cli.Command[testFlags] = (*descStub)(nil)
var _ cli.DescriptionProvider = (*descStub)(nil)

func (s *descStub) Run(_ cli.GlobalOptions, _ cli.Unknowns) error { return nil }
func (s *descStub) Help() string                                  { return s.help }
func (s *descStub) Description() string                           { return s.desc }

func newDescStub(name, help, desc string) cli.Command[any] {
	cmd := &descStub{BaseCommand: cli.NewBaseCommand[testFlags](), help: help, desc: desc}
	cmd.Name(name)
	return cmd
}

func Test_DisplayHelp_RendersMultilineDescription(t *testing.T) {
	desc := "First line of detail.\n\nSecond paragraph with install steps:\n  do this thing"
	tree := []cli.Command[any]{newDescStub("setup", "Set things up", desc)}

	out := captureStdout(t, func() {
		DisplayHelp("myapp", tree, []string{"setup"})
	})

	for _, want := range []string{"Set things up", "First line of detail.", "Second paragraph with install steps:", "  do this thing"} {
		if !strings.Contains(out, want) {
			t.Errorf("single-command help missing %q in:\n%s", want, out)
		}
	}
}

func Test_DisplayHelp_ListingUsesFirstLineOnly(t *testing.T) {
	desc := "summary\nhidden detail line"
	// a command whose Help summary accidentally spans lines must not break listings
	tree := []cli.Command[any]{newDescStub("setup", "one-liner\nleaked second line", desc)}

	out := captureStdout(t, func() {
		DisplayHelp("myapp", tree, nil)
	})

	if strings.Contains(out, "leaked second line") {
		t.Errorf("listing should only show the first line of Help, got:\n%s", out)
	}
}

func Test_DisplayHelpJSON_IncludesDescription(t *testing.T) {
	tree := []cli.Command[any]{newDescStub("setup", "Set things up", "long form description")}

	out := captureStdout(t, func() {
		DisplayHelpJSON(tree)
	})

	var infos []CommandInfo
	if err := json.Unmarshal([]byte(out), &infos); err != nil {
		t.Fatalf("failed to parse help JSON: %v", err)
	}
	if len(infos) != 1 || infos[0].Description != "long form description" {
		t.Errorf("expected description in JSON output, got: %s", out)
	}
}

func Test_DisplayHelp_AllCommands(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelp("myapp", commandTree(), nil)
	})

	assertions := []string{"Usage: myapp", "Commands:", "build", "db", "Global Options:"}
	for _, want := range assertions {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func Test_DisplayHelp_SingleCommand(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelp("myapp", commandTree(), []string{"db"})
	})

	if !strings.Contains(out, "Database commands") {
		t.Fatalf("expected command help text, got:\n%s", out)
	}
	if !strings.Contains(out, "migrate") {
		t.Fatalf("expected subcommand listed, got:\n%s", out)
	}
}

func Test_DisplayHelp_WithFlagsAndEnv(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelp("myapp", commandTree(), nil, HelpDisplayOptions{ShowFlags: true, ShowEnv: true})
	})

	if !strings.Contains(out, "--name") {
		t.Fatalf("expected flag in output, got:\n%s", out)
	}
	if !strings.Contains(out, "[env: NAME]") {
		t.Fatalf("expected env annotation in output, got:\n%s", out)
	}
}

func Test_DisplayHelp_UnknownCommand(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelp("myapp", commandTree(), []string{"nope"})
	})

	if !strings.Contains(out, "Command not found") {
		t.Fatalf("expected not-found message, got:\n%s", out)
	}
}

func Test_DisplayHelpJSON_IsValidAndStructured(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelpJSON(commandTree())
	})

	var infos []CommandInfo
	if err := json.Unmarshal([]byte(out), &infos); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}

	if len(infos) != 2 {
		t.Fatalf("expected 2 top-level commands, got %d", len(infos))
	}

	var db *CommandInfo
	for i := range infos {
		if infos[i].Name == "db" {
			db = &infos[i]
		}
	}
	if db == nil {
		t.Fatalf("expected db command in JSON")
	}
	if len(db.SubCommands) != 1 || db.SubCommands[0].Name != "migrate" {
		t.Fatalf("expected db to have a migrate subcommand, got %+v", db.SubCommands)
	}

	// the --name flag is non-positional and required, so it should be present
	var hasName bool
	for _, f := range db.Flags {
		if f.Name == "name" {
			hasName = true
			if !f.Required {
				t.Fatalf("expected name flag to be required")
			}
		}
	}
	if !hasName {
		t.Fatalf("expected name flag in JSON, got %+v", db.Flags)
	}
}

func Test_DisplayHelpJSONSchema_IsValidAndStructured(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelpJSONSchema(commandTree())
	})

	var schemas []CommandSchema
	if err := json.Unmarshal([]byte(out), &schemas); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}

	var found bool
	for _, s := range schemas {
		if s.Name == "build" {
			found = true
			if _, ok := s.Properties["name"]; !ok {
				t.Fatalf("expected name property in schema, got %+v", s.Properties)
			}
			if s.Properties["verbose"].Type != "boolean" {
				t.Fatalf("expected verbose to be boolean, got %q", s.Properties["verbose"].Type)
			}
		}
	}
	if !found {
		t.Fatalf("expected build schema in output")
	}
}

func Test_DisplayHelpAgent_Formats(t *testing.T) {
	for _, format := range []string{"plain", "md", "pretty"} {
		t.Run(format, func(t *testing.T) {
			out := captureStdout(t, func() {
				DisplayHelpAgent(AgentOptions{
					AppName:  "myapp",
					Format:   format,
					Commands: commandTree(),
				})
			})
			if out == "" {
				t.Fatalf("expected non-empty agent help for format %q", format)
			}
		})
	}
}
