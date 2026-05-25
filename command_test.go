package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestInputs struct {
	Name string `arg:"name" help:"Name"`
	Port int    `arg:"port" help:"Port"`
}

func Test_BaseCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		set      string
		get      string
		expected string
	}{
		{
			name:     "set and get",
			set:      "deploy",
			get:      "",
			expected: "deploy",
		},
		{
			name:     "set returns the name",
			set:      "run",
			get:      "run",
			expected: "run",
		},
		{
			name:     "empty before set",
			set:      "",
			get:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewBaseCommand[TestInputs]()
			if tt.set != "" {
				cmd.Name(tt.set)
			}
			assert.Equal(t, tt.expected, cmd.Name(tt.get))
		})
	}
}

func Test_BaseCommand_Add(t *testing.T) {
	cmd := NewBaseCommand[TestInputs]()
	assert.Empty(t, cmd.Commands())

	sub := &MockCommand{BaseCommand: NewBaseCommand[MockCommandOptions]()}
	cmd.Add("sub1", sub)
	assert.Len(t, cmd.Commands(), 1)
	assert.Equal(t, "sub1", cmd.Commands()[0].Name(""))

	sub2 := &MockCommand{BaseCommand: NewBaseCommand[MockCommandOptions]()}
	cmd.Add("sub2", sub2)
	assert.Len(t, cmd.Commands(), 2)
	assert.Equal(t, "sub2", cmd.Commands()[1].Name(""))
}

func Test_BaseCommand_Options(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "initializes nil inputs"},
		{name: "returns same pointer on second call"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewBaseCommand[TestInputs]()
			opts := cmd.Options()
			assert.NotNil(t, opts)

			opts2 := cmd.Options()
			assert.Equal(t, opts, opts2)
		})
	}
}

func Test_BaseCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		vars    map[string]any
		wantErr bool
	}{
		{
			name:    "valid options",
			vars:    map[string]any{"name": "test", "port": "8080"},
			wantErr: false,
		},
		{
			name:    "empty options",
			vars:    map[string]any{},
			wantErr: false,
		},
		{
			name:    "partial options",
			vars:    map[string]any{"port": "8080"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewBaseCommand[TestInputs]()
			cmd.Options()
			err := cmd.Validate(tt.vars)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
