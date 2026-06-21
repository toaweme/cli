package cli

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func mockHelpCommand(app App) chan any {
	cmdChan := make(chan any)
	run := func() error {
		go func() {
			cmdChan <- 1
		}()
		return nil
	}
	app.Add(helpCommand, NewMockCommand(run))

	return cmdChan
}

func mockMultipleCommands(app App) chan any {
	run := func() error { return nil }
	app.Add(helpCommand, NewMockCommand(run))
	app.Add(helpCommand+"2", NewMockCommand(run))

	return nil
}

func mockSubCommands(app App) chan any {
	run := func() error { return nil }
	runSub := func() error { return nil }
	app.Add(helpCommand, NewMockCommand(run)).Add("sub", NewMockCommand(runSub))

	return nil
}

func Test_App(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		settings  Config
		opts      GlobalFlags
		bootstrap func(app App) chan any

		err    error
		errors []error
		check  func(t *testing.T, app *app)
	}{
		{
			name:      "no commands",
			bootstrap: nil,
			err:       ErrNoCommands,
		},
		{
			name:      "no args",
			bootstrap: mockHelpCommand,
			err:       ErrShowingHelp,
		},
		{
			name:      "help by command Name",
			settings:  Config{},
			bootstrap: mockHelpCommand,
			args:      []string{"help"},
		},
		{
			name:      "help by option",
			settings:  Config{},
			bootstrap: mockHelpCommand,
			args:      []string{"--help"},
			err:       ErrShowingHelp,
		},
		{
			name:      "help by short option",
			settings:  Config{},
			bootstrap: mockHelpCommand,
			args:      []string{"-h"},
			err:       ErrShowingHelp,
		},
		{
			name:      "version by short option",
			settings:  Config{Name: "testapp", Version: "1.0.0"},
			bootstrap: mockHelpCommand,
			args:      []string{"-V"},
			err:       ErrShowingVersion,
		},
		{
			name:      "version by long option",
			settings:  Config{Name: "testapp", Version: "1.0.0"},
			bootstrap: mockHelpCommand,
			args:      []string{"--version"},
			err:       ErrShowingVersion,
		},
		{
			name:      "unknown command",
			settings:  Config{},
			bootstrap: mockMultipleCommands,
			args:      []string{"beep"},
			errors:    []error{ErrCommandNotFound, ErrShowingHelp},
		},
		{
			name:      "unknown command and options",
			settings:  Config{},
			bootstrap: mockMultipleCommands,
			args:      []string{"beep", "--boop"},
			errors:    []error{ErrCommandNotFound, ErrShowingHelp},
		},
		{
			name:      "global options with long flags",
			settings:  Config{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "--cwd", "/temp/dir"},
			check: func(t *testing.T, app *app) {
				t.Helper()
				assertEqual(t, "/temp/dir", app.globalFlags.Cwd)
			},
		},
		{
			// -c is no longer bound to cwd (freed for the author's own use); it falls
			// through as an unknown flag and must not populate the global Cwd.
			name:      "freed -c short is not cwd",
			settings:  Config{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "-c", "/temp/dir"},
			check: func(t *testing.T, app *app) {
				t.Helper()
				assertEqual(t, "", app.globalFlags.Cwd)
			},
		},
		{
			name:      "global options with key=value syntax",
			settings:  Config{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "--cwd=/temp/dir"},
			check: func(t *testing.T, app *app) {
				t.Helper()
				assertEqual(t, "/temp/dir", app.globalFlags.Cwd)
			},
		},
		{
			name:      "sub command",
			settings:  Config{},
			bootstrap: mockSubCommands,
			args:      []string{"help", "sub", "--cwd", "/temp/dir"},
			check: func(t *testing.T, app *app) {
				t.Helper()
				assertEqual(t, "/temp/dir", app.globalFlags.Cwd)
			},
		},
		{
			// a command's own bool flag is unknown to the global pre-scan, which would
			// otherwise swallow the following --help as that flag's value.
			name:      "help after a command bool flag",
			settings:  Config{},
			bootstrap: mockSubCommands,
			args:      []string{"help", "sub", "--beep", "--help"},
			err:       ErrShowingHelp,
		},
		{
			name:      "short help after a command bool flag",
			settings:  Config{},
			bootstrap: mockSubCommands,
			args:      []string{"help", "sub", "--beep", "-h"},
			err:       ErrShowingHelp,
		},
		{
			name:      "version after a command bool flag",
			settings:  Config{Name: "testapp", Version: "1.0.0"},
			bootstrap: mockSubCommands,
			args:      []string{"help", "sub", "--beep", "-V"},
			err:       ErrShowingVersion,
		},
		{
			// explicit --help=false must not trigger help: the command runs normally.
			name:      "explicit help=false runs the command",
			settings:  Config{},
			bootstrap: mockSubCommands,
			args:      []string{"help", "sub", "--beep", "--help=false"},
		},
		{
			// --help-values is a help mode, so it implies --help even on its own.
			name:      "help-values implies help",
			settings:  Config{},
			bootstrap: mockSubCommands,
			args:      []string{"help", "sub", "--help-values"},
			err:       ErrShowingHelp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &app{
				config:      tt.settings,
				globalFlags: &tt.opts,
				commands:    make([]Command[any], 0),
			}
			var cmdChan chan any
			if tt.bootstrap != nil {
				cmdChan = tt.bootstrap(app)
			}

			err := app.Run(tt.args)
			if tt.err != nil {
				// t.Log("error", "err", err)
				assertErrorIs(t, err, tt.err)
				return
			}
			if tt.errors != nil {
				for _, e := range tt.errors {
					assertErrorIs(t, err, e)
				}
				return
			}
			assertNoError(t, err)
			if tt.check != nil {
				tt.check(t, app)
			}

			if cmdChan == nil {
				return
			}

			expected := 1
			given := <-cmdChan

			t.Log("result", "expected", expected, "given", given)

			assertNotNil(t, given)
			assertEqual(t, expected, given)
		})
	}
}

type MockCommandConfig struct {
	Beep   bool `arg:"beep" help:"Beep"`
	Number int  `arg:"number" help:"Number"`
}

type MockCommand struct {
	BaseCommand[MockCommandConfig]
	HelpText string

	run func() error
}

var _ Command[MockCommandConfig] = (*MockCommand)(nil)

func (m MockCommand) Help() string {
	s := "Mock command"
	if m.HelpText != "" {
		s = m.HelpText
	}
	return s
}

func (m MockCommand) Validate(options map[string]any) error {
	return nil
}

func (m MockCommand) Run(options GlobalFlags, unknowns Unknowns) error {
	return m.run()
}

func NewMockCommand(run func() error) *MockCommand {
	return &MockCommand{run: run, BaseCommand: NewBaseCommand[MockCommandConfig]()}
}

func Test_boolFlagRequested(t *testing.T) {
	names := []string{"help", "h"}
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"bare long", []string{"--help"}, true},
		{"bare short", []string{"-h"}, true},
		{"after a value-taking flag", []string{"--target", "--help"}, true},
		{"after a bool-like unknown flag", []string{"--force", "-h"}, true},
		{"in the middle of arguments", []string{"cmd", "-h", "rest"}, true},
		{"explicit true", []string{"--help=true"}, true},
		{"absent", []string{"cmd", "--force", "tag"}, false},
		{"explicit false", []string{"--help=false"}, false},
		{"explicit zero", []string{"--help=0"}, false},
		{"after the -- terminator", []string{"--", "--help"}, false},
		{"name only as a flag value", []string{"--message=--help"}, false},
		{"name as a positional", []string{"help"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqual(t, tt.want, boolFlagRequested(tt.args, names))
		})
	}
}

func Test_globalBoolFlagNames(t *testing.T) {
	assertEqual(t, "help,h", strings.Join(globalBoolFlagNames("help"), ","))
	assertEqual(t, "version,V", strings.Join(globalBoolFlagNames("version"), ","))
}

func newTestApp(settings Config, opts GlobalFlags) *app {
	return &app{
		config:      settings,
		globalFlags: &opts,
		commands:    make([]Command[any], 0),
	}
}

type recordingHelp struct {
	BaseCommand[MockCommandConfig]
	gotArgs []string
}

var _ Command[MockCommandConfig] = (*recordingHelp)(nil)

func (m *recordingHelp) Help() string { return "help" }
func (m *recordingHelp) Run(_ GlobalFlags, unknowns Unknowns) error {
	m.gotArgs = unknowns.Args
	return nil
}

func Test_App_Help_RegistersUnderReservedName(t *testing.T) {
	app := NewApp(Config{Name: "myapp"}, GlobalFlags{})
	rec := &recordingHelp{BaseCommand: NewBaseCommand[MockCommandConfig]()}

	// the caller never types the reserved name
	returned := app.Help(rec)
	if returned != rec {
		t.Fatalf("Help should return the registered command")
	}
	if rec.Name("") != helpCommand {
		t.Fatalf("expected command registered as %q, got %q", helpCommand, rec.Name(""))
	}

	// it is wired so that --help dispatches to it with the command path as args
	app.Add("build", NewMockCommand(func() error { return nil }))
	_ = app.Run([]string{"--help", "build"})
	if len(rec.gotArgs) != 1 || rec.gotArgs[0] != "build" {
		t.Fatalf("expected help to receive [build], got %v", rec.gotArgs)
	}
}

func Test_App_DefaultCommand(t *testing.T) {
	tests := []struct {
		name       string
		defaultCmd func() (*MockCommand, *bool)
		wantErr    error
		wantRan    bool
	}{
		{
			name: "runs default command with no args",
			defaultCmd: func() (*MockCommand, *bool) {
				ran := false
				cmd := NewMockCommand(func() error {
					ran = true
					return nil
				})
				return cmd, &ran
			},
			wantRan: true,
		},
		{
			name:    "shows help when no default set",
			wantErr: ErrShowingHelp,
		},
		{
			name: "returns error when default command fails",
			defaultCmd: func() (*MockCommand, *bool) {
				ran := false
				cmd := NewMockCommand(func() error {
					ran = true
					return errors.New("boom")
				})
				return cmd, &ran
			},
			wantRan: true,
			wantErr: errors.New("failed to run command"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(Config{}, GlobalFlags{})
			app.Add("help", NewMockCommand(func() error { return nil }))

			var ran *bool
			if tt.defaultCmd != nil {
				cmd, r := tt.defaultCmd()
				ran = r
				app.Default(cmd)
			}

			err := app.Run([]string{})
			if tt.wantErr != nil {
				assertError(t, err)
				if errors.Is(tt.wantErr, ErrShowingHelp) {
					assertErrorIs(t, err, ErrShowingHelp)
				} else {
					assertContains(t, err.Error(), tt.wantErr.Error())
				}
			} else {
				assertNoError(t, err)
			}

			if ran != nil {
				assertEqual(t, tt.wantRan, *ran)
			}
		})
	}
}

// when Default(cmd) is set, invoking the app with flags but no command must run the
// default command with those flags parsed into its Options - the same as if the
// default command's name had been typed. --help still wins, and bare invocation
// (no args) runs the default with no flags.
func Test_App_DefaultCommand_WithFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantRan      bool
		expectedBeep bool
		expectedNum  int
		wantErr      error
	}{
		{
			name:         "flags without a command run the default command with them parsed",
			args:         []string{"--beep", "--number", "42"},
			wantRan:      true,
			expectedBeep: true,
			expectedNum:  42,
		},
		{
			name:        "key=value flags without a command",
			args:        []string{"--number=7"},
			wantRan:     true,
			expectedNum: 7,
		},
		{
			name:    "bare invocation runs the default command",
			args:    []string{},
			wantRan: true,
		},
		{
			name:    "--help wins over the default command",
			args:    []string{"--help"},
			wantRan: false,
			wantErr: ErrShowingHelp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ran := false
			cmd := NewMockCommand(func() error {
				ran = true
				return nil
			})
			app := newTestApp(Config{}, GlobalFlags{})
			app.Add("help", NewMockCommand(func() error { return nil }))
			app.Default(cmd)

			err := app.Run(tt.args)
			if tt.wantErr != nil {
				assertErrorIs(t, err, tt.wantErr)
			} else {
				assertNoError(t, err)
			}
			assertEqual(t, tt.wantRan, ran)
			if tt.wantRan {
				assertEqual(t, tt.expectedBeep, cmd.Inputs.Beep)
				assertEqual(t, tt.expectedNum, cmd.Inputs.Number)
			}
		})
	}
}

// a token in command position that matches no command is an unknown command, not a
// silent fall-through to the default. The default command declares no positional args,
// so the bare token must surface ErrCommandNotFound and show help instead of running.
func Test_App_DefaultCommand_UnknownCommand(t *testing.T) {
	ran := false
	cmd := NewMockCommand(func() error {
		ran = true
		return nil
	})
	app := newTestApp(Config{}, GlobalFlags{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Default(cmd)

	err := app.Run([]string{"foo"})
	assertErrorIs(t, err, ErrCommandNotFound)
	assertErrorIs(t, err, ErrShowingHelp)
	assertEqual(t, false, ran)
}

// an unknown command reports not-found even when --help is also passed: the unknown
// command is detected before the --help short-circuit, so --help does not mask it.
func Test_App_UnknownCommand_WithHelpFlag(t *testing.T) {
	ran := false
	cmd := NewMockCommand(func() error {
		ran = true
		return nil
	})
	app := newTestApp(Config{}, GlobalFlags{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Default(cmd)

	err := app.Run([]string{"foo", "--help"})
	assertErrorIs(t, err, ErrCommandNotFound)
	assertErrorIs(t, err, ErrShowingHelp)
	assertEqual(t, false, ran)
}

// an unrecognized trailing token after a valid command path shows help for the deepest
// matched command (here "db migrate") rather than running it with the token dropped.
func Test_App_UnknownSubcommand_ShowsDeepestHelp(t *testing.T) {
	app := NewApp(Config{Name: "beep"}, GlobalFlags{})
	rec := &recordingHelp{BaseCommand: NewBaseCommand[MockCommandConfig]()}
	app.Help(rec)

	migrateRan := false
	app.Add("db", NewMockCommand(func() error { return nil })).
		Add("migrate", NewMockCommand(func() error {
			migrateRan = true
			return nil
		}))

	err := app.Run([]string{"db", "migrate", "back"})
	assertErrorIs(t, err, ErrCommandNotFound)
	assertErrorIs(t, err, ErrShowingHelp)
	assertEqual(t, false, migrateRan)
	if len(rec.gotArgs) != 2 || rec.gotArgs[0] != "db" || rec.gotArgs[1] != "migrate" {
		t.Fatalf("expected help for [db migrate], got %v", rec.gotArgs)
	}
}

type PositionalCommandConfig struct {
	Target string `arg:"0" help:"Target"`
}

type PositionalCommand struct {
	BaseCommand[PositionalCommandConfig]
	run func() error
}

var _ Command[PositionalCommandConfig] = (*PositionalCommand)(nil)

func (m *PositionalCommand) Help() string                        { return "positional" }
func (m *PositionalCommand) Run(_ GlobalFlags, _ Unknowns) error { return m.run() }

func newPositionalCommand(run func() error) *PositionalCommand {
	return &PositionalCommand{run: run, BaseCommand: NewBaseCommand[PositionalCommandConfig]()}
}

// when the default command declares a positional arg, a leading bare token is that
// positional and runs the default; a second, undeclared bare token is unknown.
func Test_App_DefaultCommand_WithPositional(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantRan bool
		wantErr bool
	}{
		{name: "single token fills the positional", args: []string{"foo"}, wantRan: true},
		{name: "extra token is unknown", args: []string{"foo", "bar"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ran := false
			cmd := newPositionalCommand(func() error {
				ran = true
				return nil
			})
			app := newTestApp(Config{}, GlobalFlags{})
			app.Add("help", NewMockCommand(func() error { return nil }))
			app.Default(cmd)

			err := app.Run(tt.args)
			if tt.wantErr {
				assertErrorIs(t, err, ErrCommandNotFound)
				assertErrorIs(t, err, ErrShowingHelp)
			} else {
				assertNoError(t, err)
				assertEqual(t, "foo", cmd.Inputs.Target)
			}
			assertEqual(t, tt.wantRan, ran)
		})
	}
}

// an unknown flag is tolerated and the default command still runs; only unknown bare
// positional tokens trigger the unknown-command path.
func Test_App_UnknownFlagTolerated(t *testing.T) {
	ran := false
	cmd := NewMockCommand(func() error {
		ran = true
		return nil
	})
	app := newTestApp(Config{}, GlobalFlags{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	app.Default(cmd)

	err := app.Run([]string{"--unknown-flag"})
	assertNoError(t, err)
	assertEqual(t, true, ran)
}

func Test_App_CommandWithOptions(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedBeep bool
		expectedNum  int
	}{
		{
			name:         "bool option",
			args:         []string{"test", "--beep"},
			expectedBeep: true,
		},
		{
			name:        "int option",
			args:        []string{"test", "--number", "42"},
			expectedNum: 42,
		},
		{
			name:         "both options",
			args:         []string{"test", "--beep", "--number", "7"},
			expectedBeep: true,
			expectedNum:  7,
		},
		{
			name:         "key=value syntax",
			args:         []string{"test", "--beep", "--number=99"},
			expectedBeep: true,
			expectedNum:  99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured *MockCommand
			app := newTestApp(Config{}, GlobalFlags{})
			app.Add("help", NewMockCommand(func() error { return nil }))
			cmd := NewMockCommand(func() error { return nil })
			captured = cmd
			app.Add("test", cmd)

			err := app.Run(tt.args)
			assertNoError(t, err)
			assertEqual(t, tt.expectedBeep, captured.Inputs.Beep)
			assertEqual(t, tt.expectedNum, captured.Inputs.Number)
		})
	}
}

func Test_App_DisplaySubCommands(t *testing.T) {
	app := newTestApp(Config{}, GlobalFlags{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	parent := NewMockCommand(func() error {
		return ErrDisplaySubCommands
	})
	app.Add("parent", parent)

	err := app.Run([]string{"parent"})
	assertNoError(t, err)
}

func Test_exists(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		val      int
		expected bool
	}{
		{
			name:     "found",
			slice:    []int{1, 2, 3},
			val:      2,
			expected: true,
		},
		{
			name:     "not found",
			slice:    []int{1, 2, 3},
			val:      5,
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []int{},
			val:      1,
			expected: false,
		},
		{
			name:     "first element",
			slice:    []int{0, 1, 2},
			val:      0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqual(t, tt.expected, exists(tt.slice, tt.val))
		})
	}
}

func Test_App_MatchCommandByName(t *testing.T) {
	app := newTestApp(Config{}, GlobalFlags{})
	cmd1 := NewMockCommand(nil)
	cmd2 := NewMockCommand(nil)
	app.Add("alpha", cmd1)
	app.Add("beta", cmd2)

	tests := []struct {
		name     string
		arg      string
		expected string
	}{
		{
			name:     "matches first",
			arg:      "alpha",
			expected: "alpha",
		},
		{
			name:     "matches second",
			arg:      "beta",
			expected: "beta",
		},
		{
			name:     "no match",
			arg:      "gamma",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.matchCommandByName(tt.arg, app.commands)
			if tt.expected == "" {
				assertNil(t, result)
			} else {
				assertNotNil(t, result)
				assertEqual(t, tt.expected, result.Name(""))
			}
		})
	}
}

func Test_App_Commands(t *testing.T) {
	app := newTestApp(Config{}, GlobalFlags{})
	assertEmpty(t, app.Commands())

	app.Add("one", NewMockCommand(nil))
	app.Add("two", NewMockCommand(nil))
	assertLen(t, app.Commands(), 2)
}

func Test_App_MatchCommandByArgs(t *testing.T) {
	app := newTestApp(Config{}, GlobalFlags{})
	app.Add("help", NewMockCommand(nil))
	parent := NewMockCommand(nil)
	app.Add("deploy", parent)

	sub := NewMockCommand(nil)
	parent.Add("staging", sub)

	tests := []struct {
		name        string
		args        []string
		expectedCmd string
		expectedErr error
	}{
		{
			name:        "top level command",
			args:        []string{"help"},
			expectedCmd: "help",
		},
		{
			name:        "sub command",
			args:        []string{"deploy", "staging"},
			expectedCmd: "staging",
		},
		{
			name:        "command with options after",
			args:        []string{"help", "--verbose"},
			expectedCmd: "help",
		},
		{
			name:        "no match",
			args:        []string{"unknown"},
			expectedErr: ErrCommandNotFound,
		},
		{
			name:        "options before command",
			args:        []string{"--cwd", "/tmp", "help"},
			expectedCmd: "help",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, _, err := app.matchCommandByArgs(tt.args)
			if tt.expectedErr != nil {
				assertErrorIs(t, err, tt.expectedErr)
				return
			}
			assertNoError(t, err)
			assertEqual(t, tt.expectedCmd, cmd.Name(""))
		})
	}
}

type fakeFormatCodec struct{ ext string }

var _ OutputCodec = (*fakeFormatCodec)(nil)

func (f *fakeFormatCodec) Marshal(v any) ([]byte, error) { return nil, nil }
func (f *fakeFormatCodec) Extension() string             { return f.ext }

type fakeMultiFormatCodec struct{ exts []string }

var _ OutputCodec = (*fakeMultiFormatCodec)(nil)

func (f *fakeMultiFormatCodec) Marshal(v any) ([]byte, error) { return nil, nil }
func (f *fakeMultiFormatCodec) Extension() string             { return f.exts[0] }
func (f *fakeMultiFormatCodec) Extensions() []string          { return f.exts }

func Test_FormatAliases(t *testing.T) {
	single := FormatAliases(&fakeFormatCodec{ext: ".fake"})
	assertEqual(t, []string{"fake"}, single, "single-extension codec yields its trimmed name")

	multi := FormatAliases(&fakeMultiFormatCodec{exts: []string{".yml", ".yaml"}})
	assertEqual(t, []string{"yml", "yaml"}, multi, "multi-extension codec yields every trimmed name, primary first")
}

type formatRecorder struct {
	BaseCommand[MockCommandConfig]
	got string
}

var _ Command[MockCommandConfig] = (*formatRecorder)(nil)

func (m *formatRecorder) Help() string { return "rec" }
func (m *formatRecorder) Run(o GlobalFlags, _ Unknowns) error {
	m.got = o.HelpFormat
	return nil
}

func Test_App_Run_AcceptsRegisteredFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{name: "registered codec format passes", format: "fake", wantErr: false},
		{name: "built-in format passes", format: "json", wantErr: false},
		{name: "unregistered format rejected", format: "bogus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(Config{Name: "myapp"}, GlobalFlags{}).HelpOutputs(&fakeFormatCodec{ext: ".fake"})
			rec := &formatRecorder{BaseCommand: NewBaseCommand[MockCommandConfig]()}
			app.Add("run", rec)

			err := app.Run([]string{"run", "--help-format", tt.format})
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for format %q, got nil", tt.format)
				}
				// the rejection lists every accepted format: built-ins and the registered codec
				msg := err.Error()
				for _, want := range []string{"json", "jsonschema", "fake"} {
					if !strings.Contains(msg, want) {
						t.Fatalf("expected error to list accepted format %q, got: %s", want, msg)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for format %q: %v", tt.format, err)
			}
			if rec.got != tt.format {
				t.Fatalf("expected command to see format %q, got %q", tt.format, rec.got)
			}
		})
	}
}

func Test_IsRealError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil is not a real error", err: nil, want: false},
		{name: "ErrShowingHelp is a clean exit", err: ErrShowingHelp, want: false},
		{name: "ErrShowingVersion is a clean exit", err: ErrShowingVersion, want: false},
		{name: "wrapped ErrShowingHelp is a clean exit", err: fmt.Errorf("%w: %w", ErrCommandNotFound, ErrShowingHelp), want: false},
		{name: "wrapped ErrShowingVersion is a clean exit", err: fmt.Errorf("printed version: %w", ErrShowingVersion), want: false},
		{name: "a plain error is real", err: errors.New("boom"), want: true},
		{name: "ErrCommandNotFound alone is real", err: ErrCommandNotFound, want: true},
		{name: "wrapped plain error is real", err: fmt.Errorf("failed to run: %w", errors.New("boom")), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqual(t, tt.want, IsRealError(tt.err))
		})
	}
}

// Test_GlobalFlags_ArgNames guards the one unavoidable duplication: a Go struct tag
// must be a literal, so each built-in flag name lives both in the GlobalFlags `arg:`
// tag and in the matching arg* const that dispatch/parsing/help reference. If the two
// drift, the framework would scan for a flag the struct no longer declares; this test
// turns that into a build failure instead of a silent no-op.
func Test_GlobalFlags_ArgNames(t *testing.T) {
	want := map[string]string{
		"Help":       argHelp,
		"HelpValues": argHelpValues,
		"HelpFormat": argHelpFormat,
		"Version":    argVersion,
	}
	typ := reflect.TypeOf(GlobalFlags{})
	for field, argName := range want {
		f, ok := typ.FieldByName(field)
		if !ok {
			t.Fatalf("GlobalFlags has no field %q", field)
		}
		assertEqual(t, argName, f.Tag.Get(tagArg), fmt.Sprintf("GlobalFlags.%s arg tag", field))
	}
}
