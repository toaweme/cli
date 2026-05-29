package cli

import (
	"errors"
	"fmt"
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
		opts      GlobalOptions
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
			args:      []string{"-v"},
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
			args:      []string{"help", "--cwd", "/temp/dir", "--verbosity", "2"},
			check: func(t *testing.T, app *app) {
				assertEqual(t, "/temp/dir", app.globalOptions.Cwd)
				assertEqual(t, 2, app.globalOptions.Verbosity)
			},
		},
		{
			name:      "global options with short flags",
			settings:  Config{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "-c", "/temp/dir", "--verbosity", "2"},
			check: func(t *testing.T, app *app) {
				assertEqual(t, "/temp/dir", app.globalOptions.Cwd)
				assertEqual(t, 2, app.globalOptions.Verbosity)
			},
		},
		{
			name:      "global options with key=value syntax",
			settings:  Config{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "--cwd=/temp/dir", "--verbosity=2"},
			check: func(t *testing.T, app *app) {
				assertEqual(t, "/temp/dir", app.globalOptions.Cwd)
				assertEqual(t, 2, app.globalOptions.Verbosity)
			},
		},
		{
			name:      "sub command",
			settings:  Config{},
			bootstrap: mockSubCommands,
			args:      []string{"help", "sub", "-c", "/temp/dir", "--verbosity", "2"},
			check: func(t *testing.T, app *app) {
				assertEqual(t, "/temp/dir", app.globalOptions.Cwd)
				assertEqual(t, 2, app.globalOptions.Verbosity)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.settings, tt.opts)
			var cmdChan chan any
			if tt.bootstrap != nil {
				cmdChan = tt.bootstrap(app)
			}

			err := app.Run(tt.args)
			if tt.err != nil {
				t.Log("error", "err", err)
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

func (m MockCommand) Run(options GlobalOptions, unknowns Unknowns) error {
	return m.run()
}

func NewMockCommand(run func() error) *MockCommand {
	return &MockCommand{run: run, BaseCommand: NewBaseCommand[MockCommandConfig]()}
}

func newTestApp(settings Config, opts GlobalOptions) *app {
	return NewApp(settings, opts)
}

type recordingHelp struct {
	BaseCommand[MockCommandConfig]
	gotArgs []string
}

var _ Command[MockCommandConfig] = (*recordingHelp)(nil)

func (m *recordingHelp) Help() string { return "help" }
func (m *recordingHelp) Run(_ GlobalOptions, unknowns Unknowns) error {
	m.gotArgs = unknowns.Args
	return nil
}

func Test_App_Help_RegistersUnderReservedName(t *testing.T) {
	app := NewApp(Config{Name: "myapp"}, GlobalOptions{})
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
					return fmt.Errorf("boom")
				})
				return cmd, &ran
			},
			wantRan: true,
			wantErr: fmt.Errorf("failed to run default command"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(Config{}, GlobalOptions{})
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
			app := newTestApp(Config{}, GlobalOptions{})
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
	app := newTestApp(Config{}, GlobalOptions{})
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
	app := newTestApp(Config{}, GlobalOptions{})
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
	app := newTestApp(Config{}, GlobalOptions{})
	assertEmpty(t, app.Commands())

	app.Add("one", NewMockCommand(nil))
	app.Add("two", NewMockCommand(nil))
	assertLen(t, app.Commands(), 2)
}

func Test_App_MatchCommandByArgs(t *testing.T) {
	app := newTestApp(Config{}, GlobalOptions{})
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
