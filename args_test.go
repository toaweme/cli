package cli

import (
	"fmt"
	"strings"
	"testing"

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
			// -v is no longer a global short (version moved to -V), so it falls through
			// as an unknown bool flag while --help is still matched.
			name:         "global long flag plus freed -v short",
			args:         []string{"init", "arg1", "arg2", "-v", "--help", "beep"},
			structure:    &GlobalFlags{},
			expectedArgs: []string{},
			unknownArgs:  []string{"init", "arg1", "arg2", "beep"},
			expectedOptions: map[string]any{
				"help": true,
			},
			unknownOptions: map[string]any{
				"v": true,
			},
		},
		{
			name:            "key=value syntax with known option",
			args:            []string{"--cwd=/tmp/dir"},
			structure:       &GlobalFlags{},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{"cwd": "/tmp/dir"},
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
			// -c is no longer a global short (cwd is long-only), so it falls through
			// as an unknown value flag and consumes the following /tmp.
			name:         "mixed key=value and space-separated options",
			args:         []string{"--cwd=/app", "--help", "-c", "/tmp"},
			structure:    &GlobalFlags{},
			expectedArgs: []string{},
			unknownArgs:  []string{},
			expectedOptions: map[string]any{
				"cwd":  "/app",
				"help": true,
			},
			unknownOptions: map[string]any{
				"c": "/tmp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("test:", tt.name, "args: ", strings.Join(tt.args, " "))
			fields := make([]structs.Field, 0)
			if tt.structure != nil {
				var err error
				fields, err = structs.GetStructFields(tt.structure, nil, structs.DefaultEncodingTags)
				assertNoError(t, err)
			}
			args, unknownArgs, options, unknownOptions := getCommandArgs(tt.args, fields)

			assertEqual(t, tt.expectedArgs, args, "expected args")
			assertEqual(t, tt.unknownArgs, unknownArgs, "unknown args")
			assertEqual(t, tt.expectedOptions, options, "expected options")
			assertEqual(t, tt.unknownOptions, unknownOptions, "unknown options")
		})
	}
}

// repeatFlags exercises repeated flags: Tags is slice-typed (accumulates), Name
// is scalar (last-wins), Ports carries a custom sep so repeats compose with it.
type repeatFlags struct {
	Tags  []string `arg:"tags" short:"t"`
	Ports []string `arg:"ports" short:"p" sep:"|"`
	Name  string   `arg:"name" short:"n"`
}

func Test_getCommandArgs_RepeatedSliceFlag(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		structure       any
		expectedOptions map[string]any
		unknownOptions  map[string]any
	}{
		{
			name:            "repeated scalar flag keeps last value",
			args:            []string{"--name", "a", "--name", "b"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"name": "b"},
			unknownOptions:  map[string]any{},
		},
		{
			name:            "repeated slice flag accumulates every value",
			args:            []string{"-t", "a", "-t", "b"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"t": structs.MultiValue{"a", "b"}},
			unknownOptions:  map[string]any{},
		},
		{
			// each occurrence is left as-written here; the sep-split-and-concat
			// happens downstream in the struct setter (see the end-to-end test).
			name:            "repeated slice flag with sep keeps each occurrence",
			args:            []string{"-t", "a,b", "-t", "c"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"t": structs.MultiValue{"a,b", "c"}},
			unknownOptions:  map[string]any{},
		},
		{
			name:            "mixed --k=v and -k v repetition on same slice flag",
			args:            []string{"--tags=a", "-tags", "b", "--tags=c"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"tags": structs.MultiValue{"a", "b", "c"}},
			unknownOptions:  map[string]any{},
		},
		{
			name:            "single slice flag stays a one-element MultiValue",
			args:            []string{"--tags=a,b,c"},
			structure:       &repeatFlags{},
			expectedOptions: map[string]any{"tags": structs.MultiValue{"a,b,c"}},
			unknownOptions:  map[string]any{},
		},
		{
			name:            "repeated unknown flag accumulates into a slice",
			args:            []string{"--x", "a", "--x", "b", "--x=c"},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{"x": []any{"a", "b", "c"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := make([]structs.Field, 0)
			if tt.structure != nil {
				var err error
				fields, err = structs.GetStructFields(tt.structure, nil, structs.DefaultEncodingTags)
				assertNoError(t, err)
			}
			_, _, options, unknownOptions := getCommandArgs(tt.args, fields)

			assertEqual(t, tt.expectedOptions, options, "options")
			assertEqual(t, tt.unknownOptions, unknownOptions, "unknown options")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &repeatFlags{}
			fields, err := structs.GetStructFields(cfg, nil, structs.DefaultEncodingTags)
			assertNoError(t, err)

			_, _, options, _ := getCommandArgs(tt.args, fields)
			assertNoError(t, mapStructToOptions(cfg, options))

			tt.assert(t, cfg)
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
			assertEqual(t, tt.key, key)
			assertEqual(t, tt.value, value)
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
			structure:    &GlobalFlags{},
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
			name:         "value-taking flag does not swallow a following flag",
			args:         []string{"--cwd", "--help"},
			structure:    &GlobalFlags{},
			expectedArgs: []string{},
			unknownArgs:  []string{},
			expectedOptions: map[string]any{
				"cwd":  "",
				"help": true,
			},
			unknownOptions: map[string]any{},
		},
		{
			name:            "unknown flag followed by another flag stays boolean",
			args:            []string{"--foo", "--bar"},
			expectedArgs:    []string{},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{"foo": true, "bar": true},
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
			unknownOptions: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := make([]structs.Field, 0)
			if tt.structure != nil {
				var err error
				fields, err = structs.GetStructFields(tt.structure, nil, structs.DefaultEncodingTags)
				assertNoError(t, err)
			}
			args, unknownArgs, options, unknownOptions := getCommandArgs(tt.args, fields)

			assertEqual(t, tt.expectedArgs, args, "args")
			assertEqual(t, tt.unknownArgs, unknownArgs, "unknown args")
			assertEqual(t, tt.expectedOptions, options, "options")
			assertEqual(t, tt.unknownOptions, unknownOptions, "unknown options")
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
		unknownOptions  map[string]any
	}{
		{
			name:            "positional after a value-taking flag still fills slot 0",
			args:            []string{"--cwd", "/x", "foo"},
			structure:       &onePositional{},
			expectedArgs:    []string{"foo"},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{"cwd": "/x"},
			unknownOptions:  map[string]any{},
		},
		{
			name:            "positional with key=value flag before it",
			args:            []string{"--cwd=/x", "foo"},
			structure:       &onePositional{},
			expectedArgs:    []string{"foo"},
			unknownArgs:     []string{},
			expectedOptions: map[string]any{"cwd": "/x"},
			unknownOptions:  map[string]any{},
		},
		{
			name:            "second positional beyond declared slots is unknown",
			args:            []string{"foo", "bar"},
			structure:       &onePositional{},
			expectedArgs:    []string{"foo"},
			unknownArgs:     []string{"bar"},
			expectedOptions: map[string]any{},
			unknownOptions:  map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := structs.GetStructFields(tt.structure, nil, structs.DefaultEncodingTags)
			assertNoError(t, err)

			args, unknownArgs, options, unknownOptions := getCommandArgs(tt.args, fields)

			assertEqual(t, tt.expectedArgs, args, "args")
			assertEqual(t, tt.unknownArgs, unknownArgs, "unknown args")
			assertEqual(t, tt.expectedOptions, options, "options")
			assertEqual(t, tt.unknownOptions, unknownOptions, "unknown options")
		})
	}
}
