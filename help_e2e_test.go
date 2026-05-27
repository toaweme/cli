package cli_test

import (
	"os/exec"
	"strings"
	"testing"
)

func runExample(t *testing.T, example string, args ...string) string {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "./examples/" + example}, args...)...)
	out, err := cmd.CombinedOutput()
	assertNoError(t, err, "running example %s with args %v: %s", example, args, string(out))
	return string(out)
}

func Test_E2E_Help_Default(t *testing.T) {
	out := runExample(t, "basic", "help")

	assertContains(t, out, "Usage: basic")
	assertContains(t, out, "Commands:")
	assertContains(t, out, "help")
	assertContains(t, out, "version")
	assertContains(t, out, "info")
	assertContains(t, out, "Global Options:")
	assertContains(t, out, "--cwd")
}

func Test_E2E_Help_Flags(t *testing.T) {
	out := runExample(t, "deploy", "--help", "--format=plain-flags")

	assertContains(t, out, "Usage: deploy")
	assertContains(t, out, "Commands:")
	assertContains(t, out, "--force")
	assertContains(t, out, "--dry-run")
	assertContains(t, out, "deploy staging")
	assertContains(t, out, "deploy production")
}

func Test_E2E_Help_Flags_ShowsPerCommand(t *testing.T) {
	out := runExample(t, "deploy", "--help", "--format=plain-flags")

	lines := strings.Split(out, "\n")

	var stagingIdx, forceIdx int
	for i, line := range lines {
		if strings.Contains(line, "deploy staging") {
			stagingIdx = i
		}
		if stagingIdx > 0 && strings.Contains(line, "--force") {
			forceIdx = i
			break
		}
	}

	assertGreater(t, forceIdx, stagingIdx, "--force should appear after deploy staging")
}

func Test_E2E_Help_FlagsShowsEnv(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=plain-flags")

	assertContains(t, out, "[env: BUILD_OUTPUT]")
	assertContains(t, out, "[env: SERVER_PORT]")
	assertContains(t, out, "[env: CWD]")
}

func Test_E2E_Help_FlagsViaGlobalHelp(t *testing.T) {
	tests := []struct {
		name    string
		example string
		args    []string
		check   func(t *testing.T, out string)
	}{
		{
			name:    "format before help flag",
			example: "deploy",
			args:    []string{"--format=plain-flags", "--help"},
			check: func(t *testing.T, out string) {
				assertContains(t, out, "--force")
				assertContains(t, out, "--dry-run")
			},
		},
		{
			name:    "help flag before format",
			example: "deploy",
			args:    []string{"--help", "--format=plain-flags"},
			check: func(t *testing.T, out string) {
				assertContains(t, out, "--force")
				assertContains(t, out, "--dry-run")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := runExample(t, tt.example, tt.args...)
			tt.check(t, out)
		})
	}
}

func Test_E2E_Help_FlagsSingleCommand(t *testing.T) {
	out := runExample(t, "deploy", "help", "deploy", "--format=plain-flags")

	assertContains(t, out, "staging")
	assertContains(t, out, "production")
	assertContains(t, out, "--force")
	assertContains(t, out, "--dry-run")
	assertNotContains(t, out, "--0")
}

func Test_E2E_Help_BasicCommand(t *testing.T) {
	out := runExample(t, "basic", "help", "info")

	assertContains(t, out, "Print current global options")
	assertContains(t, out, "$ info")
}

func Test_E2E_Help_FormatMd(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=md")

	assertContains(t, out, "## build")
	assertContains(t, out, "```")
	assertContains(t, out, "`--output`, `-o`")
}

func Test_E2E_Help_FormatPlain(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=plain")

	assertNotContains(t, out, "```")
	assertContains(t, out, "build")
	assertContains(t, out, "--output")
	assertContains(t, out, "BUILD_OUTPUT")
}

func Test_E2E_Help_FormatJSON(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=json")
	assertValidJSON(t, out)
}

func Test_E2E_Help_FormatJSONSchema(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=jsonschema")
	assertValidJSON(t, out)
}

func Test_E2E_Help_FormatJSON_Detail(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=json")

	type flagInfo struct {
		Name     string `json:"name"`
		Short    string `json:"short"`
		Help     string `json:"help"`
		Type     string `json:"type"`
		Required bool   `json:"required"`
		Env      string `json:"env"`
	}
	type cmdInfo struct {
		Name        string     `json:"name"`
		Help        string     `json:"help"`
		Flags       []flagInfo `json:"flags"`
		SubCommands []struct {
			Name  string     `json:"name"`
			Help  string     `json:"help"`
			Flags []flagInfo `json:"flags"`
		} `json:"subcommands"`
	}

	var commands []cmdInfo
	unmarshalJSON(t, out, &commands)

	var dbCmd *cmdInfo
	for i := range commands {
		if commands[i].Name == "db" {
			dbCmd = &commands[i]
			break
		}
	}
	assertNotNil(t, dbCmd, "should have db command")
	assertLen(t, dbCmd.SubCommands, 3)

	seed := dbCmd.SubCommands[1]
	assertEqual(t, "seed", seed.Name)
	assertEqual(t, "Seed the database with test data", seed.Help)

	var fileFlag bool
	for _, f := range seed.Flags {
		if f.Name == "file" {
			fileFlag = true
			assertEqual(t, "string", f.Type)
			assertTrue(t, f.Required)
			assertEqual(t, "f", f.Short)
		}
	}
	assertTrue(t, fileFlag, "should have file flag")
}

func Test_E2E_Help_FormatJSONSchema_Detail(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=jsonschema")

	type schemaField struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	var schemas []struct {
		Name       string                 `json:"name"`
		Help       string                 `json:"help"`
		Properties map[string]schemaField `json:"properties"`
		Required   []string               `json:"required"`
	}

	unmarshalJSON(t, out, &schemas)

	var buildSchema *struct {
		Name       string                 `json:"name"`
		Help       string                 `json:"help"`
		Properties map[string]schemaField `json:"properties"`
		Required   []string               `json:"required"`
	}
	for i := range schemas {
		if schemas[i].Name == "build" {
			buildSchema = &schemas[i]
			break
		}
	}
	assertNotNil(t, buildSchema, "should have build schema")

	_, hasOutput := buildSchema.Properties["output"]
	assertTrue(t, hasOutput, "should have output property")
	assertEqual(t, "string", buildSchema.Properties["output"].Type)
	assertEqual(t, "Output directory", buildSchema.Properties["output"].Description)

	_, hasRace := buildSchema.Properties["race"]
	assertTrue(t, hasRace, "should have race property")
	assertEqual(t, "boolean", buildSchema.Properties["race"].Type)
}

func Test_E2E_Help_FormatPrettyDefault(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=pretty")

	assertContains(t, out, "build")
	assertContains(t, out, "Build the project")
	assertContains(t, out, "db migrate")
	assertContains(t, out, "Global Options")
	assertContains(t, out, "--output, -o")
	assertContains(t, out, "BUILD_OUTPUT")
	assertContains(t, out, "string, required")
}

func Test_E2E_Help_FormatPretty_Examples(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=pretty")

	assertContains(t, out, "full build")
	assertContains(t, out, "full build -o ./dist --race")
	assertContains(t, out, "full serve -p 3000 --reload")
	assertContains(t, out, "full db migrate -n 3")
}

func Test_E2E_Help_ScopedCommand(t *testing.T) {
	out := runExample(t, "full", "help", "build", "--format=md")

	assertContains(t, out, "## build")
	assertContains(t, out, "`--output`, `-o`")
	assertNotContains(t, out, "## serve")
	assertNotContains(t, out, "## db")
}

func Test_E2E_Help_ScopedJSON(t *testing.T) {
	type flagInfo struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	type cmdInfo struct {
		Name  string     `json:"name"`
		Help  string     `json:"help"`
		Flags []flagInfo `json:"flags"`
	}

	tests := []struct {
		name         string
		args         []string
		wantNames    []string
		excludeNames []string
		checkFlags   map[string][]string
	}{
		{
			name:         "scoped to build",
			args:         []string{"help", "build", "--format=json"},
			wantNames:    []string{"build"},
			excludeNames: []string{"serve", "help", "db", "config"},
			checkFlags:   map[string][]string{"build": {"output", "tags", "race", "verbose"}},
		},
		{
			name:         "scoped to serve",
			args:         []string{"help", "serve", "--format=json"},
			wantNames:    []string{"serve"},
			excludeNames: []string{"build", "help", "db"},
			checkFlags:   map[string][]string{"serve": {"port", "host", "tls", "reload"}},
		},
		{
			name:         "unscoped returns all",
			args:         []string{"--help", "--format=json"},
			wantNames:    []string{"help", "version", "completion", "dev", "build", "serve", "config", "db"},
			excludeNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := runExample(t, "full", tt.args...)
			assertValidJSON(t, out)

			var commands []cmdInfo
			unmarshalJSON(t, out, &commands)

			nameSet := make(map[string]bool)
			for _, cmd := range commands {
				nameSet[cmd.Name] = true
			}

			for _, want := range tt.wantNames {
				assertTrue(t, nameSet[want], "should contain command %q", want)
			}
			for _, exclude := range tt.excludeNames {
				assertTrue(t, !nameSet[exclude], "should not contain command %q", exclude)
			}

			for cmdName, wantFlags := range tt.checkFlags {
				for _, cmd := range commands {
					if cmd.Name != cmdName {
						continue
					}
					flagSet := make(map[string]bool)
					for _, f := range cmd.Flags {
						flagSet[f.Name] = true
					}
					for _, wf := range wantFlags {
						assertTrue(t, flagSet[wf], "command %q should have flag %q", cmdName, wf)
					}
				}
			}
		})
	}
}

func Test_E2E_Help_ScopedJSONSchema(t *testing.T) {
	type schemaField struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	type schema struct {
		Name       string                 `json:"name"`
		Help       string                 `json:"help"`
		Properties map[string]schemaField `json:"properties"`
		Required   []string               `json:"required"`
	}

	tests := []struct {
		name         string
		args         []string
		wantNames    []string
		excludeNames []string
		checkProps   map[string][]string
	}{
		{
			name:         "scoped to build",
			args:         []string{"help", "build", "--format=jsonschema"},
			wantNames:    []string{"build"},
			excludeNames: []string{"serve", "help", "db migrate"},
			checkProps:   map[string][]string{"build": {"output", "race", "tags", "verbose"}},
		},
		{
			name:         "scoped to db shows parent and all subs",
			args:         []string{"help", "db", "--format=jsonschema"},
			wantNames:    []string{"db", "db migrate", "db seed", "db reset"},
			excludeNames: []string{"build", "serve"},
		},
		{
			name:         "unscoped returns all",
			args:         []string{"--help", "--format=jsonschema"},
			wantNames:    []string{"build", "serve", "db", "db migrate"},
			excludeNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := runExample(t, "full", tt.args...)
			assertValidJSON(t, out)

			var schemas []schema
			unmarshalJSON(t, out, &schemas)

			nameSet := make(map[string]bool)
			for _, s := range schemas {
				nameSet[s.Name] = true
			}

			for _, want := range tt.wantNames {
				assertTrue(t, nameSet[want], "should contain schema %q", want)
			}
			for _, exclude := range tt.excludeNames {
				assertTrue(t, !nameSet[exclude], "should not contain schema %q", exclude)
			}

			for schemaName, wantProps := range tt.checkProps {
				for _, s := range schemas {
					if s.Name != schemaName {
						continue
					}
					for _, wp := range wantProps {
						_, ok := s.Properties[wp]
						assertTrue(t, ok, "schema %q should have property %q", schemaName, wp)
					}
				}
			}
		})
	}
}

func Test_E2E_Complete_TopLevel(t *testing.T) {
	out := runExample(t, "full", "__complete", "")

	assertContains(t, out, "help\tDisplay help")
	assertContains(t, out, "build\tBuild the project")
	assertContains(t, out, "serve\t")
	assertContains(t, out, "db\t")
	assertContains(t, out, ":4")
}

func Test_E2E_Complete_Subcommand(t *testing.T) {
	out := runExample(t, "full", "__complete", "db", "")

	assertContains(t, out, "migrate\tRun database migrations")
	assertContains(t, out, "seed\t")
	assertContains(t, out, "reset\t")
	assertNotContains(t, out, "build")
}

func Test_E2E_Complete_Flags(t *testing.T) {
	out := runExample(t, "full", "__complete", "build", "--")

	assertContains(t, out, "--output\t")
	assertContains(t, out, "--race\t")
	assertContains(t, out, "--tags\t")
	assertContains(t, out, "--cwd\t")
}

func Test_E2E_Complete_Partial(t *testing.T) {
	out := runExample(t, "full", "__complete", "b")

	assertContains(t, out, "build\t")
	assertNotContains(t, out, "serve")
	assertNotContains(t, out, "help")
}

func Test_E2E_Completion_Bash(t *testing.T) {
	out := runExample(t, "full", "completion", "bash")

	assertContains(t, out, "complete")
	assertContains(t, out, "__complete")
	assertContains(t, out, "full")
}

func Test_E2E_Completion_Zsh(t *testing.T) {
	out := runExample(t, "full", "completion", "zsh")

	assertContains(t, out, "#compdef full")
	assertContains(t, out, "__complete")
}

func Test_E2E_Completion_Fish(t *testing.T) {
	out := runExample(t, "full", "completion", "fish")

	assertContains(t, out, "complete -c full")
	assertContains(t, out, "__complete")
}
