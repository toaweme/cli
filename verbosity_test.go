package cli

import (
	"testing"

	"github.com/toaweme/structs"
)

func Test_Verbosity_Level(t *testing.T) {
	tests := []struct {
		name string
		v    Verbosity
		want int
	}{
		{name: "none", v: Verbosity{}, want: 0},
		{name: "v", v: Verbosity{V1: true}, want: 1},
		{name: "vv", v: Verbosity{V2: true}, want: 2},
		{name: "vvv", v: Verbosity{V3: true}, want: 3},
		{name: "highest wins", v: Verbosity{V1: true, V3: true}, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqual(t, tt.want, tt.v.Level())
			assertEqual(t, tt.want > 0, tt.v.Verbose())
			assertEqual(t, tt.want >= 2, tt.v.AtLeast(2))
		})
	}
}

// Test_Verbosity_Parsing proves the embeddable flags resolve onto a command's
// input struct end-to-end: -v/-vv/-vvv each match their own short token and set
// the matching bool, which Level then reads back.
func Test_Verbosity_Parsing(t *testing.T) {
	type inputs struct {
		Verbosity
		Name string `arg:"name"`
	}

	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "no flag", args: []string{"--name", "x"}, want: 0},
		{name: "-v", args: []string{"-v"}, want: 1},
		{name: "-vv", args: []string{"-vv"}, want: 2},
		{name: "-vvv", args: []string{"-vvv"}, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &inputs{}
			fields, err := structs.GetStructFields(in, nil, structs.DefaultEncodingTags)
			if err != nil {
				t.Fatalf("failed to get struct fields: %v", err)
			}
			_, _, options, _ := getCommandArgs(tt.args, fields)
			if err := mapStructToOptions(in, options); err != nil {
				t.Fatalf("failed to map options: %v", err)
			}
			assertEqual(t, tt.want, in.Level())
		})
	}
}
