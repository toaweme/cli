package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/toaweme/structs"
)

// parseFields returns the fields of structure, none when structure is nil.
func parseFields(t *testing.T, structure any) []structs.Field {
	t.Helper()
	if structure == nil {
		return []structs.Field{}
	}
	fields, err := structs.GetStructFields(structure, nil, structs.DefaultEncodingTags)
	assertNoError(t, err)

	return fields
}

func Test_Args(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		structure       any
		expectedArgs    []string
		unknownArgs     []string
		expectedOptions map[string]any
		expectedGlobals map[string]any
		err             error
	}{
		{
			name:            "no struct fields given leaves global flags only",
			args:            []string{"init", "arg1", "arg2", "-v", "2", "--help", "beep", "--boop"},
			expectedArgs:    []string{},
			unknownArgs:     []string{"init", "arg1", "arg2", "2", "beep"},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{
				"help": true,
			},
			err: ErrUnknownFlag,
		},
		{
			// -v is not a global short (version is -V), so it is an unknown flag while --help is matched.
			name:         "global long flag plus freed -v short",
			args:         []string{"init", "arg1", "arg2", "-v", "--help", "beep"},
			structure:    &GlobalFlags{},
			expectedArgs: []string{},
			unknownArgs:  []string{"init", "arg1", "arg2", "beep"},
			expectedOptions: map[string]any{
				"help": true,
			},
			expectedGlobals: map[string]any{},
			err:             ErrUnknownFlag,
		},
		{
			name:            "key=value syntax with known option",
			args:            []string{"--cwd=/tmp/dir"},
			structure:       &GlobalFlags{},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{"cwd": "/tmp/dir"},
			expectedGlobals: map[string]any{},
		},
		{
			name:            "key=value syntax with unknown option",
			args:            []string{"--foo=bar"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{},
			err:             ErrUnknownFlag,
		},
		{
			name:            "key=value with equals in value",
			args:            []string{"--cwd=key=val"},
			structure:       &GlobalFlags{},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{"cwd": "key=val"},
			expectedGlobals: map[string]any{},
		},
		{
			// -c is not a global short (cwd is long-only), so it is an unknown flag
			// and the following /tmp is a positional.
			name:         "mixed key=value and space-separated options",
			args:         []string{"--cwd=/app", "--help", "-c", "/tmp"},
			structure:    &GlobalFlags{},
			expectedArgs: []string{},
			unknownArgs:  []string{"/tmp"},
			expectedOptions: map[string]any{
				"cwd":  "/app",
				"help": true,
			},
			expectedGlobals: map[string]any{},
			err:             ErrUnknownFlag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("test:", tt.name, "args: ", strings.Join(tt.args, " "))
			args, unknownArgs, options, globals, err := getCommandArgs(tt.args, parseFields(t, tt.structure), parseFields(t, &GlobalFlags{}))

			assertEqual(t, tt.expectedArgs, args, "expected args")
			assertEqual(t, tt.unknownArgs, unknownArgs, "unknown args")
			assertEqual(t, tt.expectedOptions, options, "expected options")
			assertEqual(t, tt.expectedGlobals, globals, "expected globals")
			assertErrorIs(t, err, tt.err)
		})
	}
}

// repeatFlags exercises repeated flags: Tags is slice-typed (accumulates), Name
// is scalar (last-wins), Ports carries a custom sep so repeats compose with it,
// and Vars is a map filled from key=value entries.
type repeatFlags struct {
	Tags  []string          `arg:"tags" short:"t"`
	Ports []string          `arg:"ports" short:"p" sep:"|"`
	Name  string            `arg:"name" short:"n"`
	Vars  map[string]string `arg:"var" short:"e"`
}

func Test_getCommandArgs_RepeatedFlag(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		structure       any
		expectedOptions map[string]any
		err             error
	}{
		{
			name:            "repeated scalar flag keeps last value",
			args:            []string{"--name", "a", "--name", "b"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"name": "b"},
		},
		{
			name:            "repeated slice flag accumulates every value",
			args:            []string{"-t", "a", "-t", "b"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"tags": structs.MultiValue{"a", "b"}},
		},
		{
			// each occurrence is left as-written here; the sep-split-and-concat
			// happens downstream in the struct setter (see the end-to-end test).
			name:            "repeated slice flag with sep keeps each occurrence",
			args:            []string{"-t", "a,b", "-t", "c"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"tags": structs.MultiValue{"a,b", "c"}},
		},
		{
			name:            "mixed --k=v and -k v repetition on same slice flag",
			args:            []string{"--tags=a", "-tags", "b", "--tags=c"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"tags": structs.MultiValue{"a", "b", "c"}},
		},
		{
			name:            "single slice flag stays a one-element MultiValue",
			args:            []string{"--tags=a,b,c"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"tags": structs.MultiValue{"a,b,c"}},
		},
		{
			name:      "map flag entries split on the first equals",
			args:      []string{"--var", "url=a=b", "-e", "k=", "--var=list=a,b", "--var", "x=-5"},
			structure: &repeatFlags{},
			expectedOptions: map[string]any{
				"var": map[string]string{
					"url":  "a=b",
					"k":    "",
					"list": "a,b",
					"x":    "-5",
				},
			},
		},
		{
			name:      "repeated map key keeps last value",
			args:      []string{"--var", "a=1", "-e", "b=2", "--var=a=3"},
			structure: &repeatFlags{},
			expectedOptions: map[string]any{
				"var": map[string]string{
					"a": "3",
					"b": "2",
				},
			},
		},
		{
			name:            "map flag entry without a key=value is invalid",
			args:            []string{"--var", "k", "-e", "=v", "--var"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{},
			err:             ErrInvalidFlagValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, options, _, err := getCommandArgs(tt.args, parseFields(t, tt.structure), nil)

			assertEqual(t, tt.expectedOptions, options, "options")
			assertErrorIs(t, err, tt.err)
		})
	}
}

// Test_RepeatedSliceFlag_EndToEnd proves repeats survive all the way onto the
// struct field, and that the sep tag composes with the repeats.
func Test_RepeatedSliceFlag_EndToEnd(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		assert func(t *testing.T, cfg *repeatFlags)
	}{
		{
			name: "plain repeats accumulate onto the slice",
			args: []string{"-t", "a", "-t", "b"},
			assert: func(t *testing.T, cfg *repeatFlags) {
				t.Helper()
				assertEqual(t, []string{"a", "b"}, cfg.Tags)
			},
		},
		{
			name: "repeats compose with the default comma sep",
			args: []string{"-t", "a,b", "-t", "c"},
			assert: func(t *testing.T, cfg *repeatFlags) {
				t.Helper()
				assertEqual(t, []string{"a", "b", "c"}, cfg.Tags)
			},
		},
		{
			name: "repeats compose with a custom sep tag",
			args: []string{"-p", "8080|9090", "-p", "3000"},
			assert: func(t *testing.T, cfg *repeatFlags) {
				t.Helper()
				assertEqual(t, []string{"8080", "9090", "3000"}, cfg.Ports)
			},
		},
		{
			name: "repeated scalar keeps last value onto the field",
			args: []string{"-n", "a", "-n", "b"},
			assert: func(t *testing.T, cfg *repeatFlags) {
				t.Helper()
				assertEqual(t, "b", cfg.Name)
			},
		},
		{
			name: "map entries land on the map field",
			args: []string{"--var", "a=1", "-e", "b=x,y"},
			assert: func(t *testing.T, cfg *repeatFlags) {
				t.Helper()
				assertEqual(t, map[string]string{"a": "1", "b": "x,y"}, cfg.Vars)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &repeatFlags{}
			_, _, options, _, err := getCommandArgs(tt.args, parseFields(t, cfg), nil)
			assertNoError(t, err)
			assertNoError(t, mapStructToOptions(cfg, options))

			tt.assert(t, cfg)
		})
	}
}

func Test_matchField(t *testing.T) {
	fields := []structs.Field{
		{Tags: map[string]string{"arg": "cwd", "short": "c"}},
		{Tags: map[string]string{"arg": "help", "short": "h"}},
		{Tags: map[string]string{"arg": "help-values", "short": ""}},
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
			search:   "help-values",
			expected: "help-values",
		},
		{
			name:     "no match",
			search:   "unknown",
			expected: "",
		},
		{
			name:     "empty search matches empty short tag",
			search:   "",
			expected: "help-values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchField(fields, tt.search)
			if tt.expected == "" {
				assertNil(t, result)
			} else {
				assertNotNil(t, result)
				assertEqual(t, tt.expected, result.Tags["arg"])
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
		expectedGlobals map[string]any
		err             error
	}{
		{
			name:            "empty args",
			args:            []string{},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{},
		},
		{
			name:            "only positional args without struct",
			args:            []string{"one", "two", "three"},
			expectedArgs:    []string{},
			unknownArgs:     []string{"one", "two", "three"},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{},
		},
		{
			name:            "triple dash prefix is unknown",
			args:            []string{"---cwd", "/tmp"},
			structure:       &GlobalFlags{},
			expectedArgs:    []string{},
			unknownArgs:     []string{"/tmp"},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{},
			err:             ErrUnknownFlag,
		},
		{
			name:         "bool flag between positional args",
			args:         []string{"cmd", "--help", "arg"},
			structure:    &GlobalFlags{},
			expectedArgs: []string{},
			unknownArgs:  []string{"cmd", "arg"},
			expectedOptions: map[string]any{
				"help": true,
			},
			expectedGlobals: map[string]any{},
		},
		{
			name:            "key=value with single dash",
			args:            []string{"-cwd=/tmp"},
			structure:       &GlobalFlags{},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{"cwd": "/tmp"},
			expectedGlobals: map[string]any{},
		},
		{
			name:         "value-taking flag does not swallow a following flag",
			args:         []string{"--cwd", "--help"},
			structure:    &GlobalFlags{},
			expectedArgs: []string{},
			unknownArgs:  []string{},
			expectedOptions: map[string]any{
				"help": true,
			},
			expectedGlobals: map[string]any{},
			err:             ErrInvalidFlagValue,
		},
		{
			name:            "first unknown flag is reported",
			args:            []string{"--foo", "--bar"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{},
			err:             ErrUnknownFlag,
		},
		{
			name:         "all global options at once",
			args:         []string{"--cwd=/app", "--help", "--version"},
			structure:    &GlobalFlags{},
			expectedArgs: []string{},
			unknownArgs:  []string{},
			expectedOptions: map[string]any{
				"cwd":     "/app",
				"help":    true,
				"version": true,
			},
			expectedGlobals: map[string]any{},
		},
		{
			name:            "global flag takes a negative number",
			args:            []string{"--cwd", "-5", "-h"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{
				"cwd":  "-5",
				"help": true,
			},
		},
		{
			name:            "dash word after a value flag is not its value",
			args:            []string{"--cwd", "-inf"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{},
			err:             ErrInvalidFlagValue,
		},
		{
			name:            "negative number in a positional slot is unknown",
			args:            []string{"-5"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{},
			err:             ErrUnknownFlag,
		},
		{
			name:            "short flags are not clustered",
			args:            []string{"-hV"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{},
			err:             ErrUnknownFlag,
		},
		{
			name:            "flags with an empty name are unknown",
			args:            []string{"--", "--=v", "-="},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{},
			err:             ErrUnknownFlag,
		},
		{
			name:            "spaced word after a bool flag is a positional",
			args:            []string{"--help", "false"},
			expectedArgs:    []string{},
			unknownArgs:     []string{"false"},
			expectedOptions: map[string]any{},
			expectedGlobals: map[string]any{
				"help": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, unknownArgs, options, globals, err := getCommandArgs(tt.args, parseFields(t, tt.structure), parseFields(t, &GlobalFlags{}))

			assertEqual(t, tt.expectedArgs, args, "args")
			assertEqual(t, tt.unknownArgs, unknownArgs, "unknown args")
			assertEqual(t, tt.expectedOptions, options, "options")
			assertEqual(t, tt.expectedGlobals, globals, "globals")
			assertErrorIs(t, err, tt.err)
		})
	}
}

// positionals are matched by their order among the real arguments, not their raw index,
// so a global flag in front of a positional must not knock it out of its declared slot.
func Test_getCommandArgs_Positional(t *testing.T) {
	type onePositional struct {
		Cwd    string `arg:"cwd"`
		Target string `arg:"0"`
	}

	tests := []struct {
		name            string
		args            []string
		structure       any
		expectedArgs    []string
		unknownArgs     []string
		expectedOptions map[string]any
		err             error
	}{
		{
			name:            "positional after a value-taking flag still fills slot 0",
			args:            []string{"--cwd", "/x", "foo"},
			structure:       &onePositional{},
			expectedArgs:    []string{"foo"},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{"cwd": "/x"},
		},
		{
			name:            "positional with key=value flag before it",
			args:            []string{"--cwd=/x", "foo"},
			structure:       &onePositional{},
			expectedArgs:    []string{"foo"},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{"cwd": "/x"},
		},
		{
			name:            "second positional beyond declared slots is unknown",
			args:            []string{"foo", "bar"},
			structure:       &onePositional{},
			expectedArgs:    []string{"foo"},
			unknownArgs:     []string{"bar"},
			expectedOptions: map[string]any{},
		},
		{
			name:            "bare dash is a positional",
			args:            []string{"-"},
			structure:       &onePositional{},
			expectedArgs:    []string{"-"},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
		},
		{
			name:            "empty string is a positional",
			args:            []string{""},
			structure:       &onePositional{},
			expectedArgs:    []string{""},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, unknownArgs, options, _, err := getCommandArgs(tt.args, parseFields(t, tt.structure), nil)

			assertEqual(t, tt.expectedArgs, args, "args")
			assertEqual(t, tt.unknownArgs, unknownArgs, "unknown args")
			assertEqual(t, tt.expectedOptions, options, "options")
			assertErrorIs(t, err, tt.err)
		})
	}
}
