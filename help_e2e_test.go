package cli_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runExample(t *testing.T, example string, args ...string) string {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "./examples/" + example}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "running example %s with args %v: %s", example, args, string(out))
	return string(out)
}

func Test_E2E_Help_Default(t *testing.T) {
	out := runExample(t, "basic", "help")

	assert.Contains(t, out, "Usage: basic")
	assert.Contains(t, out, "Commands:")
	assert.Contains(t, out, "help")
	assert.Contains(t, out, "version")
	assert.Contains(t, out, "info")
	assert.Contains(t, out, "Global Options:")
	assert.Contains(t, out, "--cwd")
}

func Test_E2E_Help_Flags(t *testing.T) {
	out := runExample(t, "deploy", "help", "--flags")

	assert.Contains(t, out, "Usage: deploy")
	assert.Contains(t, out, "Commands:")
	assert.Contains(t, out, "--force")
	assert.Contains(t, out, "--dry-run")
	assert.Contains(t, out, "deploy staging")
	assert.Contains(t, out, "deploy production")
}

func Test_E2E_Help_Flags_ShowsPerCommand(t *testing.T) {
	out := runExample(t, "deploy", "help", "--flags")

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

	assert.Greater(t, forceIdx, stagingIdx, "--force should appear after deploy staging")
}

func Test_E2E_Help_FlagsShowsEnv(t *testing.T) {
	out := runExample(t, "full", "help", "--flags")

	assert.Contains(t, out, "[env: BUILD_OUTPUT]")
	assert.Contains(t, out, "[env: SERVE_PORT]")
	assert.Contains(t, out, "[env: CWD]")
}

func Test_E2E_Help_FlagsViaGlobalHelp(t *testing.T) {
	tests := []struct {
		name    string
		example string
		args    []string
		check   func(t *testing.T, out string)
	}{
		{
			name:    "flags with help flag",
			example: "deploy",
			args:    []string{"--flags", "--help"},
			check: func(t *testing.T, out string) {
				assert.Contains(t, out, "--force")
				assert.Contains(t, out, "--dry-run")
			},
		},
		{
			name:    "help flag before flags",
			example: "deploy",
			args:    []string{"--help", "--flags"},
			check: func(t *testing.T, out string) {
				assert.Contains(t, out, "--force")
				assert.Contains(t, out, "--dry-run")
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
	out := runExample(t, "deploy", "help", "deploy", "--flags")

	assert.Contains(t, out, "staging")
	assert.Contains(t, out, "production")
	assert.Contains(t, out, "--force")
	assert.Contains(t, out, "--dry-run")
	assert.Contains(t, out, "Image tag to deploy")
}

func Test_E2E_Help_BasicCommand(t *testing.T) {
	out := runExample(t, "basic", "help", "info")

	assert.Contains(t, out, "Print current global options")
	assert.Contains(t, out, "$ info")
}

func Test_E2E_Help_FormatMd(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=md")

	assert.Contains(t, out, "## build")
	assert.Contains(t, out, "```")
	assert.Contains(t, out, "`--output`, `-o`")
}

func Test_E2E_Help_FormatPlain(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=plain")

	assert.NotContains(t, out, "```")
	assert.Contains(t, out, "build")
	assert.Contains(t, out, "--output")
	assert.Contains(t, out, "BUILD_OUTPUT")
}

func Test_E2E_Help_FormatJSON(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=json")

	var commands []json.RawMessage
	err := json.Unmarshal([]byte(out), &commands)
	require.NoError(t, err, "should output valid JSON")
}

func Test_E2E_Help_FormatJSONSchema(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=jsonschema")

	var schemas []json.RawMessage
	err := json.Unmarshal([]byte(out), &schemas)
	require.NoError(t, err, "should output valid JSON")
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
	err := json.Unmarshal([]byte(out), &commands)
	require.NoError(t, err, "should be valid JSON: %s", out)

	var dbCmd *cmdInfo
	for i := range commands {
		if commands[i].Name == "db" {
			dbCmd = &commands[i]
			break
		}
	}
	require.NotNil(t, dbCmd, "should have db command")
	require.Len(t, dbCmd.SubCommands, 3)

	seed := dbCmd.SubCommands[1]
	assert.Equal(t, "seed", seed.Name)
	assert.Equal(t, "Seed the database with test data", seed.Help)

	var fileFlag bool
	for _, f := range seed.Flags {
		if f.Name == "file" {
			fileFlag = true
			assert.Equal(t, "string", f.Type)
			assert.True(t, f.Required)
			assert.Equal(t, "f", f.Short)
		}
	}
	assert.True(t, fileFlag, "should have file flag")
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

	err := json.Unmarshal([]byte(out), &schemas)
	require.NoError(t, err, "should be valid JSON: %s", out)

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
	require.NotNil(t, buildSchema, "should have build schema")

	assert.Contains(t, buildSchema.Properties, "output")
	assert.Equal(t, "string", buildSchema.Properties["output"].Type)
	assert.Equal(t, "Output directory", buildSchema.Properties["output"].Description)

	assert.Contains(t, buildSchema.Properties, "race")
	assert.Equal(t, "boolean", buildSchema.Properties["race"].Type)
}

func Test_E2E_Help_FormatPrettyDefault(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=pretty")

	assert.Contains(t, out, "build")
	assert.Contains(t, out, "Build the project")
	assert.Contains(t, out, "db migrate")
	assert.Contains(t, out, "Global Options")
	assert.Contains(t, out, "--output, -o")
	assert.Contains(t, out, "BUILD_OUTPUT")
	assert.Contains(t, out, "string, required")
}

func Test_E2E_Help_FormatPretty_Examples(t *testing.T) {
	out := runExample(t, "full", "--help", "--format=pretty")

	assert.Contains(t, out, "full build")
	assert.Contains(t, out, "full build -o ./dist --race")
	assert.Contains(t, out, "full serve -p 3000 --reload")
	assert.Contains(t, out, "full db migrate -n 3")
}

func Test_E2E_Help_ScopedCommand(t *testing.T) {
	out := runExample(t, "full", "help", "build", "--format=md")

	assert.Contains(t, out, "## build")
	assert.Contains(t, out, "`--output`, `-o`")
	assert.NotContains(t, out, "## serve")
	assert.NotContains(t, out, "## db")
}

func Test_E2E_Complete_TopLevel(t *testing.T) {
	out := runExample(t, "full", "__complete", "")

	assert.Contains(t, out, "help\tDisplay help")
	assert.Contains(t, out, "build\tBuild the project")
	assert.Contains(t, out, "serve\t")
	assert.Contains(t, out, "db\t")
	assert.Contains(t, out, ":4")
}

func Test_E2E_Complete_Subcommand(t *testing.T) {
	out := runExample(t, "full", "__complete", "db", "")

	assert.Contains(t, out, "migrate\tRun database migrations")
	assert.Contains(t, out, "seed\t")
	assert.Contains(t, out, "reset\t")
	assert.NotContains(t, out, "build")
}

func Test_E2E_Complete_Flags(t *testing.T) {
	out := runExample(t, "full", "__complete", "build", "--")

	assert.Contains(t, out, "--output\t")
	assert.Contains(t, out, "--race\t")
	assert.Contains(t, out, "--tags\t")
	assert.Contains(t, out, "--cwd\t")
}

func Test_E2E_Complete_Partial(t *testing.T) {
	out := runExample(t, "full", "__complete", "b")

	assert.Contains(t, out, "build\t")
	assert.NotContains(t, out, "serve")
	assert.NotContains(t, out, "help")
}

func Test_E2E_Completion_Bash(t *testing.T) {
	out := runExample(t, "full", "completion", "bash")

	assert.Contains(t, out, "complete")
	assert.Contains(t, out, "__complete")
	assert.Contains(t, out, "full")
}

func Test_E2E_Completion_Zsh(t *testing.T) {
	out := runExample(t, "full", "completion", "zsh")

	assert.Contains(t, out, "#compdef full")
	assert.Contains(t, out, "__complete")
}

func Test_E2E_Completion_Fish(t *testing.T) {
	out := runExample(t, "full", "completion", "fish")

	assert.Contains(t, out, "complete -c full")
	assert.Contains(t, out, "__complete")
}
