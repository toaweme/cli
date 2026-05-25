package cli

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	tlog "github.com/toaweme/log"
)

func mockHelpCommand(app App) chan any {
	var cmdChan = make(chan any)
	run := func() error {
		slog.Info("running single mock command with channel")
		go func() {
			cmdChan <- 1
		}()
		return nil
	}
	app.Add(helpCommand, NewMockCommand(run))

	return cmdChan
}

func mockMultipleCommands(app App) chan any {
	run := func() error {
		slog.Info("running multiple mock command")
		return nil
	}
	app.Add(helpCommand, NewMockCommand(run))
	app.Add(helpCommand+"2", NewMockCommand(run))

	return nil
}

func mockSubCommands(app App) chan any {
	run := func() error {
		slog.Info("running parent command")
		return nil
	}
	runSub := func() error {
		slog.Info("running sub command")
		return nil
	}
	app.Add(helpCommand, NewMockCommand(run)).Add("sub", NewMockCommand(runSub))

	return nil
}

func Test_App(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		settings  Settings
		opts      GlobalOptions
		bootstrap func(app App) chan any

		err    error
		errors []error
		assert func(t *testing.T, app *CLI)
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
			settings:  Settings{},
			bootstrap: mockHelpCommand,
			args:      []string{"help"},
		},
		{
			name:      "help by option",
			settings:  Settings{},
			bootstrap: mockHelpCommand,
			args:      []string{"--help"},
			err:       ErrShowingHelp,
		},
		{
			name:      "help by short option",
			settings:  Settings{},
			bootstrap: mockHelpCommand,
			args:      []string{"-h"},
			err:       ErrShowingHelp,
		},
		{
			name:      "version by short option",
			settings:  Settings{Name: "testapp", Version: "1.0.0"},
			bootstrap: mockHelpCommand,
			args:      []string{"-v"},
			err:       ErrShowingVersion,
		},
		{
			name:      "version by long option",
			settings:  Settings{Name: "testapp", Version: "1.0.0"},
			bootstrap: mockHelpCommand,
			args:      []string{"--version"},
			err:       ErrShowingVersion,
		},
		{
			name:      "unknown command",
			settings:  Settings{},
			bootstrap: mockMultipleCommands,
			args:      []string{"beep"},
			errors:    []error{ErrCommandNotFound, ErrShowingHelp},
		},
		{
			name:      "unknown command and options",
			settings:  Settings{},
			bootstrap: mockMultipleCommands,
			args:      []string{"beep", "--boop"},
			errors:    []error{ErrCommandNotFound, ErrShowingHelp},
		},
		{
			name:      "global options with long flags",
			settings:  Settings{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "--cwd", "/temp/dir", "--verbosity", "2"},
			assert: func(t *testing.T, app *CLI) {
				assert.Equal(t, "/temp/dir", app.globalOptions.Cwd)
				assert.Equal(t, 2, app.globalOptions.Verbosity)
			},
		},
		{
			name:      "global options with short flags",
			settings:  Settings{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "-c", "/temp/dir", "--verbosity", "2"},
			assert: func(t *testing.T, app *CLI) {
				assert.Equal(t, "/temp/dir", app.globalOptions.Cwd)
				assert.Equal(t, 2, app.globalOptions.Verbosity)
			},
		},
		{
			name:      "global options with key=value syntax",
			settings:  Settings{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "--cwd=/temp/dir", "--verbosity=2"},
			assert: func(t *testing.T, app *CLI) {
				assert.Equal(t, "/temp/dir", app.globalOptions.Cwd)
				assert.Equal(t, 2, app.globalOptions.Verbosity)
			},
		},
		{
			name:      "sub command",
			settings:  Settings{},
			bootstrap: mockSubCommands,
			args:      []string{"help", "sub", "-c", "/temp/dir", "--verbosity", "2"},
			assert: func(t *testing.T, app *CLI) {
				assert.Equal(t, "/temp/dir", app.globalOptions.Cwd)
				assert.Equal(t, 2, app.globalOptions.Verbosity)
			},
		},
	}

	// os.Args[1:]
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.settings, tt.opts, tlog.NewExtendedLogger(slog.Default()))
			var cmdChan chan any
			if tt.bootstrap != nil {
				cmdChan = tt.bootstrap(app)
			}

			err := app.Run(tt.args)
			if tt.err != nil {
				slog.Info("error", "err", err)
				assert.ErrorIs(t, err, tt.err)
				return
			}
			if tt.errors != nil {
				for _, e := range tt.errors {
					assert.ErrorIs(t, err, e)
				}
				return
			}
			assert.NoError(t, err)
			if tt.assert != nil {
				tt.assert(t, app)
			}

			// if bootstrap returns nil, we don't need to assert the result
			// just the error above
			if cmdChan == nil {
				return
			}

			expected := 1
			given := <-cmdChan

			slog.Info("result", "expected", expected, "given", given)

			assert.NotNil(t, given)
			assert.Equal(t, expected, given)
		})
	}
}

type MockCommandOptions struct {
	Beep   bool `arg:"beep" help:"Beep"`
	Number int  `arg:"number" help:"Number"`
}

type MockCommand struct {
	BaseCommand[MockCommandOptions]
	HelpText string

	run func() error
}

var _ Command[MockCommandOptions] = (*MockCommand)(nil)

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
	return &MockCommand{run: run, BaseCommand: NewBaseCommand[MockCommandOptions]()}
}

func newTestApp(settings Settings, opts GlobalOptions) *CLI {
	return NewApp(settings, opts, tlog.NewExtendedLogger(slog.Default()))
}

func Test_App_DefaultCommand(t *testing.T) {
	var ran bool
	app := newTestApp(Settings{}, GlobalOptions{})
	defaultCmd := NewMockCommand(func() error {
		ran = true
		return nil
	})
	app.Add("help", NewMockCommand(nil))
	app.Default(defaultCmd)

	err := app.Run([]string{})
	assert.NoError(t, err)
	assert.True(t, ran)
}

func Test_App_DefaultCommand_NotSet(t *testing.T) {
	app := newTestApp(Settings{}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))

	err := app.Run([]string{})
	assert.ErrorIs(t, err, ErrShowingHelp)
}

func Test_App_JSON_Flag(t *testing.T) {
	app := newTestApp(Settings{}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))

	err := app.Run([]string{"help", "--json"})
	assert.NoError(t, err)
	assert.True(t, app.globalOptions.JSON)
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
			app := newTestApp(Settings{}, GlobalOptions{})
			app.Add("help", NewMockCommand(func() error { return nil }))
			cmd := NewMockCommand(func() error { return nil })
			captured = cmd
			app.Add("test", cmd)

			err := app.Run(tt.args)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBeep, captured.Inputs.Beep)
			assert.Equal(t, tt.expectedNum, captured.Inputs.Number)
		})
	}
}

func Test_App_DisplaySubCommands(t *testing.T) {
	app := newTestApp(Settings{}, GlobalOptions{})
	app.Add("help", NewMockCommand(func() error { return nil }))
	parent := NewMockCommand(func() error {
		return ErrDisplaySubCommands
	})
	app.Add("parent", parent)

	err := app.Run([]string{"parent"})
	assert.NoError(t, err)
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
			assert.Equal(t, tt.expected, exists(tt.slice, tt.val))
		})
	}
}

func Test_App_MatchCommandByName(t *testing.T) {
	app := newTestApp(Settings{}, GlobalOptions{})
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
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.Name(""))
			}
		})
	}
}

func Test_App_Logger(t *testing.T) {
	logger := tlog.NewExtendedLogger(slog.Default())
	app := NewApp(Settings{}, GlobalOptions{}, logger)
	assert.Equal(t, logger, app.Logger())
}

func Test_App_Commands(t *testing.T) {
	app := newTestApp(Settings{}, GlobalOptions{})
	assert.Empty(t, app.Commands())

	app.Add("one", NewMockCommand(nil))
	app.Add("two", NewMockCommand(nil))
	assert.Len(t, app.Commands(), 2)
}

func Test_App_MatchCommandByArgs(t *testing.T) {
	app := newTestApp(Settings{}, GlobalOptions{})
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
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCmd, cmd.Name(""))
		})
	}
}
