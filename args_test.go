package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toaweme/structs"
)

func Test_Args(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		structure       any
		expectedArgs    []string
		unknownArgs     []string
		expectedOptions map[string]any
		unknownOptions  map[string]any
	}{
		{
			name:            "no struct fields given produces --help with value",
			args:            []string{"init", "arg1", "arg2", "-v", "2", "--help", "beep", "--boop"},
			expectedArgs:    []string{},
			unknownArgs:     []string{"init", "arg1", "arg2"},
			expectedOptions: map[string]any{},
			unknownOptions: map[string]any{
				"v":    "2",
				"help": "beep",
				"boop": true,
			},
		},
		{
			name:         "global options",
			args:         []string{"init", "arg1", "arg2", "-v", "2", "--help", "beep"},
			structure:    &GlobalOptions{},
			expectedArgs: []string{},
			unknownArgs:  []string{"init", "arg1", "arg2", "beep"},
			expectedOptions: map[string]any{
				"v":    "2",
				"help": true,
			},
			unknownOptions: map[string]any{},
		},
		{
			name:         "global options",
			args:         []string{"init", "arg1", "arg2", "-v", "2", "--help", "beep"},
			structure:    &GlobalOptions{},
			expectedArgs: []string{},
			unknownArgs:  []string{"init", "arg1", "arg2", "beep"},
			expectedOptions: map[string]any{
				"v":    "2",
				"help": true,
			},
			unknownOptions: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("test:", tt.name, "args: ", strings.Join(tt.args, " "))
			var fields = make([]structs.Field, 0)
			if tt.structure != nil {
				var err error
				fields, err = structs.GetStructFields(tt.structure, nil)
				assert.NoError(t, err)
			}
			args, unknownArgs, options, unknownOptions := getCommandArgs(tt.args, fields)

			// spew.Dump(args, unknownArgs, options, unknownOptions)
			assert.Equal(t, tt.expectedArgs, args, "expected args")
			assert.Equal(t, tt.unknownArgs, unknownArgs, "unknown args")
			assert.Equal(t, tt.expectedOptions, options, "expected options")
			assert.Equal(t, tt.unknownOptions, unknownOptions, "unknown options")
		})
	}
}
