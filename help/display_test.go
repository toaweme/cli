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

func (s *flagStub) Run(_ cli.GlobalFlags, _ cli.Unknowns) error { return nil }
func (s *flagStub) Help() string                                { return s.help }

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

func (s *descStub) Run(_ cli.GlobalFlags, _ cli.Unknowns) error { return nil }
func (s *descStub) Help() string                                { return s.help }
func (s *descStub) Description() string                         { return s.desc }

func newDescStub(name, help, desc string) cli.Command[any] {
	cmd := &descStub{BaseCommand: cli.NewBaseCommand[testFlags](), help: help, desc: desc}
	cmd.Name(name)
	return cmd
}

// providerStub overrides the BaseCommand Examples/Args/Flags defaults with
// multi-line content.
type providerStub struct {
	cli.BaseCommand[testFlags]
}

var _ cli.Command[testFlags] = (*providerStub)(nil)

func (s *providerStub) Run(_ cli.GlobalFlags, _ cli.Unknowns) error { return nil }
func (s *providerStub) Help() string                                { return "Query things" }
func (s *providerStub) Examples() [][]string {
	return [][]string{
		{"myapp query --name=foo"},
		{"myapp query --name=bar", "result: 42 rows"},
	}
}
func (s *providerStub) Args() map[int][]string {
	return map[int][]string{0: {"the target to query", "accepts a glob pattern"}}
}
func (s *providerStub) Flags() map[string][]string {
	return map[string][]string{"--name, -n": {"name of the thing", "must be unique"}}
}

func newProviderStub(name string) cli.Command[any] {
	cmd := &providerStub{BaseCommand: cli.NewBaseCommand[testFlags]()}
	cmd.Name(name)
	return cmd
}

func Test_DisplayHelp_RendersArgAndFlagDocs(t *testing.T) {
	tree := []cli.Command[any]{newProviderStub("query")}

	out := captureStdout(t, func() {
		DisplayHelp(os.Stdout, "myapp", tree, []string{"query"})
	})

	for _, want := range []string{"Arguments:", "[0]", "the target to query", "accepts a glob pattern", "Flag details:", "--name, -n", "name of the thing", "must be unique"} {
		if !strings.Contains(out, want) {
			t.Errorf("single-command help missing %q in:\n%s", want, out)
		}
	}
}

func Test_DisplayHelpAgent_RendersMultilineExampleAndDocs(t *testing.T) {
	tree := []cli.Command[any]{newProviderStub("query")}

	out := captureStdout(t, func() {
		DisplayHelpAgent(os.Stdout, AgentOptions{AppName: "myapp", Format: "plain", Commands: tree})
	})

	for _, want := range []string{"❯ myapp query --name=bar", "result: 42 rows", "Arguments:", "Flag details:", "must be unique"} {
		if !strings.Contains(out, want) {
			t.Errorf("agent help missing %q in:\n%s", want, out)
		}
	}
}

func Test_DisplayHelpJSON_IncludesExamplesAndDocs(t *testing.T) {
	tree := []cli.Command[any]{newProviderStub("query")}

	out := captureStdout(t, func() {
		DisplayHelpJSON(os.Stdout, tree)
	})

	var infos []CommandInfo
	if err := json.Unmarshal([]byte(out), &infos); err != nil {
		t.Fatalf("failed to parse help JSON: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 command, got %d", len(infos))
	}
	info := infos[0]
	if len(info.Examples) != 2 || len(info.Examples[1]) != 2 || info.Examples[1][1] != "result: 42 rows" {
		t.Errorf("expected multi-line examples in JSON, got: %+v", info.Examples)
	}
	if got := info.ArgDocs["0"]; len(got) != 2 || got[0] != "the target to query" {
		t.Errorf("expected arg docs in JSON, got: %+v", info.ArgDocs)
	}
	if got := info.FlagDocs["--name, -n"]; len(got) != 2 || got[1] != "must be unique" {
		t.Errorf("expected flag docs in JSON, got: %+v", info.FlagDocs)
	}
}

type enumFlags struct {
	Format string `arg:"format" help:"output format" rules:"oneof:json,yaml,toml" default:"json"`
}

type enumStub struct {
	cli.BaseCommand[enumFlags]
}

var _ cli.Command[enumFlags] = (*enumStub)(nil)

func (s *enumStub) Run(_ cli.GlobalFlags, _ cli.Unknowns) error { return nil }
func (s *enumStub) Help() string                                { return "Generate things" }

func newEnumStub(name string) cli.Command[any] {
	cmd := &enumStub{BaseCommand: cli.NewBaseCommand[enumFlags]()}
	cmd.Name(name)
	return cmd
}

func Test_DisplayHelp_ShowsOneOfValues(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelp(os.Stdout, "myapp", []cli.Command[any]{newEnumStub("gen")}, []string{"gen"})
	})

	if !strings.Contains(out, "one of: json, yaml, toml") {
		t.Errorf("expected allowed values in help, got:\n%s", out)
	}
}

func Test_DisplayHelpJSONSchema_IncludesEnum(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelpJSONSchema(os.Stdout, []cli.Command[any]{newEnumStub("gen")})
	})

	var schemas []CommandSchema
	if err := json.Unmarshal([]byte(out), &schemas); err != nil {
		t.Fatalf("failed to parse schema JSON: %v", err)
	}
	if len(schemas) == 0 {
		t.Fatalf("expected at least one schema")
	}
	got := schemas[0].Properties["format"].Enum
	want := []string{"json", "yaml", "toml"}
	if len(got) != len(want) {
		t.Fatalf("expected enum %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected enum %v, got %v", want, got)
		}
	}
}

type nestedEnumFlags struct {
	Database struct {
		Mode string `arg:"mode" help:"connection mode" rules:"oneof:tcp,unix,tls"`
	} `arg:"database"`
}

type nestedEnumStub struct {
	cli.BaseCommand[nestedEnumFlags]
}

var _ cli.Command[nestedEnumFlags] = (*nestedEnumStub)(nil)

func (s *nestedEnumStub) Run(_ cli.GlobalFlags, _ cli.Unknowns) error { return nil }
func (s *nestedEnumStub) Help() string                                { return "Connect to a database" }

func newNestedEnumStub(name string) cli.Command[any] {
	cmd := &nestedEnumStub{BaseCommand: cli.NewBaseCommand[nestedEnumFlags]()}
	cmd.Name(name)
	return cmd
}

func Test_DisplayHelp_ShowsOneOfValuesForNestedSubField(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelp(os.Stdout, "myapp", []cli.Command[any]{newNestedEnumStub("connect")}, []string{"connect"})
	})

	if !strings.Contains(out, "one of: tcp, unix, tls") {
		t.Errorf("expected nested sub-field allowed values in listing, got:\n%s", out)
	}
}

func Test_AgentDocs_ShowsOneOfValuesForNestedSubField(t *testing.T) {
	out := buildAgentOutput("myapp", []cli.Command[any]{newNestedEnumStub("connect")}, "md", nil, false, nil)

	if !strings.Contains(out, "one of: tcp, unix, tls") {
		t.Errorf("expected nested sub-field allowed values in flag table, got:\n%s", out)
	}
	if !strings.Contains(out, "database.mode") {
		t.Errorf("expected nested sub-field rendered with dotted FQN flag, got:\n%s", out)
	}
}

type RepoFlags struct {
	Groups string `arg:"groups" short:"g" env:"MEND_GROUPS" help:"Comma-separated repo groups"`
	Repos  string `arg:"repos" short:"r" env:"MEND_REPOS" help:"Comma-separated repo paths"`
}

type embeddedFlags struct {
	RepoFlags
	Force bool `arg:"force" short:"f" help:"Overwrite without merging"`
}

type embeddedStub struct {
	cli.BaseCommand[embeddedFlags]
}

var _ cli.Command[embeddedFlags] = (*embeddedStub)(nil)

func (s *embeddedStub) Run(_ cli.GlobalFlags, _ cli.Unknowns) error { return nil }
func (s *embeddedStub) Help() string                                { return "Write quality config" }

func newEmbeddedStub(name string) cli.Command[any] {
	cmd := &embeddedStub{BaseCommand: cli.NewBaseCommand[embeddedFlags]()}
	cmd.Name(name)
	return cmd
}

// fields from an embedded (anonymous) struct are promoted to plain top-level
// flags, so they must show up in help just like inline fields.
func Test_DisplayHelp_ShowsEmbeddedStructFlags(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelp(os.Stdout, "myapp", []cli.Command[any]{newEmbeddedStub("write")}, []string{"write"})
	})

	for _, want := range []string{"groups", "repos", "force"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected embedded flag %q in help, got:\n%s", want, out)
		}
	}
}

type fakeCodec struct {
	ext      string
	gotValue any
}

var _ cli.OutputCodec = (*fakeCodec)(nil)

func (f *fakeCodec) Marshal(v any) ([]byte, error) {
	f.gotValue = v
	return []byte("FAKE-OUTPUT"), nil
}

func (f *fakeCodec) Extension() string { return f.ext }

func Test_DisplayHelpEncoded_WrapsCommandsAndPrints(t *testing.T) {
	codec := &fakeCodec{ext: ".fake"}
	out := captureStdout(t, func() {
		if err := DisplayHelpEncoded(os.Stdout, []cli.Command[any]{newEnumStub("gen")}, codec); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "FAKE-OUTPUT") {
		t.Errorf("expected codec output on stdout, got: %q", out)
	}
	doc, ok := codec.gotValue.(commandsDoc)
	if !ok {
		t.Fatalf("expected commandsDoc passed to codec, got %T", codec.gotValue)
	}
	if len(doc.Commands) != 1 || doc.Commands[0].Name != "gen" {
		t.Errorf("expected wrapped command 'gen', got %+v", doc.Commands)
	}
}

func Test_DisplayHelp_ListsExtraFormatsInHint(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelp(os.Stdout, "myapp", []cli.Command[any]{newEnumStub("gen")}, nil, DisplayOptions{Formats: []string{"yaml", "toml"}})
	})

	if !strings.Contains(out, "jsonschema, yaml, toml") {
		t.Errorf("expected --help-format hint to append dynamic formats after the built-ins, got:\n%s", out)
	}
}

func Test_AgentDocs_ListsExtraFormatsInHint(t *testing.T) {
	out := buildAgentOutput("myapp", []cli.Command[any]{newEnumStub("gen")}, "md", []string{"yaml", "toml"}, false, nil)

	if !strings.Contains(out, "jsonschema, yaml, toml") {
		t.Errorf("expected global --help-format hint to append dynamic formats, got:\n%s", out)
	}
}

func Test_DisplayHelp_RendersMultilineDescription(t *testing.T) {
	desc := "First line of detail.\n\nSecond paragraph with install steps:\n  do this thing"
	tree := []cli.Command[any]{newDescStub("setup", "Set things up", desc)}

	out := captureStdout(t, func() {
		DisplayHelp(os.Stdout, "myapp", tree, []string{"setup"})
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
		DisplayHelp(os.Stdout, "myapp", tree, nil)
	})

	if strings.Contains(out, "leaked second line") {
		t.Errorf("listing should only show the first line of Help, got:\n%s", out)
	}
}

func Test_DisplayHelpJSON_IncludesDescription(t *testing.T) {
	tree := []cli.Command[any]{newDescStub("setup", "Set things up", "long form description")}

	out := captureStdout(t, func() {
		DisplayHelpJSON(os.Stdout, tree)
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
		DisplayHelp(os.Stdout, "myapp", commandTree(), nil)
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
		DisplayHelp(os.Stdout, "myapp", commandTree(), []string{"db"})
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
		DisplayHelp(os.Stdout, "myapp", commandTree(), nil, DisplayOptions{ShowFlags: true, ShowEnv: true})
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
		DisplayHelp(os.Stdout, "myapp", commandTree(), []string{"nope"})
	})

	if !strings.Contains(out, "Command not found") {
		t.Fatalf("expected not-found message, got:\n%s", out)
	}
}

func Test_DisplayHelpJSON_IsValidAndStructured(t *testing.T) {
	out := captureStdout(t, func() {
		DisplayHelpJSON(os.Stdout, commandTree())
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
		DisplayHelpJSONSchema(os.Stdout, commandTree())
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
				DisplayHelpAgent(os.Stdout, AgentOptions{
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
