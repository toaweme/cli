package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name:         "global options with short and long flags",
			args:         []string{"init", "arg1", "arg2", "-v", "--help", "beep"},
			structure:    &GlobalOptions{},
			expectedArgs: []string{},
			unknownArgs:  []string{"init", "arg1", "arg2", "beep"},
			expectedOptions: map[string]any{
				"v":    true,
				"help": true,
			},
			unknownOptions: map[string]any{},
		},
		{
			name:            "key=value syntax with known option",
			args:            []string{"--verbosity=2"},
			structure:       &GlobalOptions{},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{"verbosity": "2"},
			unknownOptions:  map[string]any{},
		},
		{
			name:            "key=value syntax with unknown option",
			args:            []string{"--foo=bar"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{"foo": "bar"},
		},
		{
			name:            "key=value with equals in value",
			args:            []string{"--filter=key=val"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{"filter": "key=val"},
		},
		{
			name:         "mixed key=value and space-separated options",
			args:         []string{"--verbosity=2", "--help", "-c", "/tmp"},
			structure:    &GlobalOptions{},
			expectedArgs: []string{},
			unknownArgs:  []string{},
			expectedOptions: map[string]any{
				"verbosity": "2",
				"help":      true,
				"c":         "/tmp",
			},
			unknownOptions: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("test:", tt.name, "args: ", strings.Join(tt.args, " "))
			fields := make([]structs.Field, 0)
			if tt.structure != nil {
				var err error
				fields, err = structs.GetStructFields(tt.structure, nil)
				require.NoError(t, err)
			}
			args, unknownArgs, options, unknownOptions := getCommandArgs(tt.args, fields)

			assert.Equal(t, tt.expectedArgs, args, "expected args")
			assert.Equal(t, tt.unknownArgs, unknownArgs, "unknown args")
			assert.Equal(t, tt.expectedOptions, options, "expected options")
			assert.Equal(t, tt.unknownOptions, unknownOptions, "unknown options")
		})
	}
}

func Test_splitKeyValue(t *testing.T) {
	tests := []struct {
		name  string
		arg   string
		key   string
		value string
	}{
		{
			name:  "simple pair",
			arg:   "cwd=/tmp",
			key:   "cwd",
			value: "/tmp",
		},
		{
			name:  "equals in value",
			arg:   "filter=a=b",
			key:   "filter",
			value: "a=b",
		},
		{
			name:  "empty value",
			arg:   "flag=",
			key:   "flag",
			value: "",
		},
		{
			name:  "no equals",
			arg:   "flag",
			key:   "flag",
			value: "",
		},
		{
			name:  "value with spaces",
			arg:   "msg=hello world",
			key:   "msg",
			value: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value := splitKeyValue(tt.arg)
			assert.Equal(t, tt.key, key)
			assert.Equal(t, tt.value, value)
		})
	}
}

func Test_matchField(t *testing.T) {
	fields := []structs.Field{
		{Tags: map[string]string{"arg": "cwd", "short": "c"}},
		{Tags: map[string]string{"arg": "help", "short": "h"}},
		{Tags: map[string]string{"arg": "verbosity", "short": ""}},
	}

	tests := []struct {
		name     string
		search   string
		expected string
	}{
		{
			name:     "match by arg tag",
			search:   "cwd",
			expected: "cwd",
		},
		{
			name:     "match by short tag",
			search:   "c",
			expected: "cwd",
		},
		{
			name:     "match arg without short",
			search:   "verbosity",
			expected: "verbosity",
		},
		{
			name:     "no match",
			search:   "unknown",
			expected: "",
		},
		{
			name:     "empty search matches empty short tag",
			search:   "",
			expected: "verbosity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchField(fields, tt.search)
			if tt.expected == "" {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.Tags["arg"])
			}
		})
	}
}

func Test_getCommandArgs_EdgeCases(t *testing.T) {
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
			name:            "empty args",
			args:            []string{},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{},
		},
		{
			name:            "only positional args without struct",
			args:            []string{"one", "two", "three"},
			expectedArgs:    []string{},
			unknownArgs:     []string{"one", "two", "three"},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{},
		},
		{
			name:            "trailing option without value",
			args:            []string{"--orphan"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{"orphan": true},
		},
		{
			name:            "double dash prefix stripped",
			args:            []string{"---triple"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{"triple": true},
		},
		{
			name:         "bool flag between positional args",
			args:         []string{"cmd", "--help", "arg"},
			structure:    &GlobalOptions{},
			expectedArgs: []string{},
			unknownArgs:  []string{"cmd", "arg"},
			expectedOptions: map[string]any{
				"help": true,
			},
			unknownOptions: map[string]any{},
		},
		{
			name:            "key=value with single dash",
			args:            []string{"-x=42"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{"x": "42"},
		},
		{
			name:         "all global options at once",
			args:         []string{"--cwd=/app", "--help", "--version", "--verbosity=2"},
			structure:    &GlobalOptions{},
			expectedArgs: []string{},
			unknownArgs:  []string{},
			expectedOptions: map[string]any{
				"cwd":       "/app",
				"help":      true,
				"version":   true,
				"verbosity": "2",
			},
			unknownOptions: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := make([]structs.Field, 0)
			if tt.structure != nil {
				var err error
				fields, err = structs.GetStructFields(tt.structure, nil)
				require.NoError(t, err)
			}
			args, unknownArgs, options, unknownOptions := getCommandArgs(tt.args, fields)

			assert.Equal(t, tt.expectedArgs, args, "args")
			assert.Equal(t, tt.unknownArgs, unknownArgs, "unknown args")
			assert.Equal(t, tt.expectedOptions, options, "options")
			assert.Equal(t, tt.unknownOptions, unknownOptions, "unknown options")
		})
	}
}
