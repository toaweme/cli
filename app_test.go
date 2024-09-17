package cli

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
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
			err:       ErrNoArguments,
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
		},
		{
			name:      "help by short option",
			settings:  Settings{},
			bootstrap: mockHelpCommand,
			args:      []string{"-h"},
		},
		{
			name:      "unknown command",
			settings:  Settings{},
			bootstrap: mockMultipleCommands,
			args:      []string{"beep"},
			err:       ErrCommandNotFound,
		},
		{
			name:      "unknown command and options",
			settings:  Settings{},
			bootstrap: mockMultipleCommands,
			args:      []string{"beep", "--boop"},
			err:       ErrCommandNotFound,
		},
		{
			name:      "global options",
			settings:  Settings{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "--cwd", "/temp/dir", "-v", "2"},
			assert: func(t *testing.T, app *CLI) {
				assert.Equal(t, "/temp/dir", app.globalOptions.Cwd)
				assert.Equal(t, 2, app.globalOptions.Verbosity)
			},
		},
		{
			name:      "global options",
			settings:  Settings{},
			bootstrap: mockMultipleCommands,
			args:      []string{"help", "--c", "/temp/dir", "-v", "2"},
			assert: func(t *testing.T, app *CLI) {
				assert.Equal(t, "/temp/dir", app.globalOptions.Cwd)
				assert.Equal(t, 2, app.globalOptions.Verbosity)
			},
		},
		{
			name:      "sub command",
			settings:  Settings{},
			bootstrap: mockSubCommands,
			args:      []string{"help", "sub", "--c", "/temp/dir", "-v", "2"},
			assert: func(t *testing.T, app *CLI) {
				assert.Equal(t, "/temp/dir", app.globalOptions.Cwd)
				assert.Equal(t, 2, app.globalOptions.Verbosity)
			},
		},
	}

	// os.Args[1:]
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.settings, tt.opts)
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

	run func() error
}

func (m MockCommand) Help() string {
	return "Mock command"
}

func (m MockCommand) Validate(options map[string]any) error {
	return nil
}

func (m MockCommand) Run(options GlobalOptions) error {
	return m.run()
}

func NewMockCommand(run func() error) *MockCommand {
	return &MockCommand{run: run, BaseCommand: NewBaseCommand[MockCommandOptions]()}
}

var _ Command[MockCommandOptions] = (*MockCommand)(nil)
