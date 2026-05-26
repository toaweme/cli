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

func TestE2E_Help_Default(t *testing.T) {
	out := runExample(t, "basic", "help")

	assert.Contains(t, out, "Usage: basic")
	assert.Contains(t, out, "Commands:")
	assert.Contains(t, out, "help")
	assert.Contains(t, out, "version")
	assert.Contains(t, out, "info")
	assert.Contains(t, out, "Global Options:")
	assert.Contains(t, out, "--cwd")
	assert.Contains(t, out, "--flags")
	assert.Contains(t, out, "--json")
	assert.Contains(t, out, "--jsonschema")
}

func TestE2E_Help_Flags(t *testing.T) {
	out := runExample(t, "deploy", "help", "--flags")

	assert.Contains(t, out, "Usage: deploy")
	assert.Contains(t, out, "Commands:")
	assert.Contains(t, out, "--force")
	assert.Contains(t, out, "--dry-run")
	assert.Contains(t, out, "deploy staging")
	assert.Contains(t, out, "deploy production")
}

func TestE2E_Help_Flags_ShowsPerCommand(t *testing.T) {
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

func TestE2E_Help_JSON(t *testing.T) {
	out := runExample(t, "deploy", "help", "--json")

	var commands []struct {
		Name        string `json:"name"`
		Help        string `json:"help"`
		Flags       []struct {
			Name     string `json:"name"`
			Short    string `json:"short"`
			Help     string `json:"help"`
			Type     string `json:"type"`
			Required bool   `json:"required"`
			Env      string `json:"env"`
		} `json:"flags"`
		SubCommands []struct {
			Name  string `json:"name"`
			Help  string `json:"help"`
			Flags []struct {
				Name     string `json:"name"`
				Short    string `json:"short"`
				Help     string `json:"help"`
				Type     string `json:"type"`
				Required bool   `json:"required"`
				Env      string `json:"env"`
			} `json:"flags"`
		} `json:"subcommands"`
	}

	err := json.Unmarshal([]byte(out), &commands)
	require.NoError(t, err, "should be valid JSON: %s", out)

	require.GreaterOrEqual(t, len(commands), 3)

	var deployCmd *struct {
		Name        string `json:"name"`
		Help        string `json:"help"`
		Flags       []struct {
			Name     string `json:"name"`
			Short    string `json:"short"`
			Help     string `json:"help"`
			Type     string `json:"type"`
			Required bool   `json:"required"`
			Env      string `json:"env"`
		} `json:"flags"`
		SubCommands []struct {
			Name  string `json:"name"`
			Help  string `json:"help"`
			Flags []struct {
				Name     string `json:"name"`
				Short    string `json:"short"`
				Help     string `json:"help"`
				Type     string `json:"type"`
				Required bool   `json:"required"`
				Env      string `json:"env"`
			} `json:"flags"`
		} `json:"subcommands"`
	}
	for i := range commands {
		if commands[i].Name == "deploy" {
			deployCmd = &commands[i]
			break
		}
	}
	require.NotNil(t, deployCmd, "should have deploy command")
	require.Len(t, deployCmd.SubCommands, 2)

	staging := deployCmd.SubCommands[0]
	assert.Equal(t, "staging", staging.Name)
	assert.Equal(t, "Deploy an image tag", staging.Help)
	require.Len(t, staging.Flags, 3)

	var tagFlag, forceFlag, dryRunFlag bool
	for _, f := range staging.Flags {
		switch f.Name {
		case "0":
			tagFlag = true
			assert.Equal(t, "string", f.Type)
			assert.True(t, f.Required)
			assert.Equal(t, "DEPLOY_TAG", f.Env)
		case "force":
			forceFlag = true
			assert.Equal(t, "bool", f.Type)
			assert.Equal(t, "f", f.Short)
		case "dry-run":
			dryRunFlag = true
			assert.Equal(t, "bool", f.Type)
		}
	}
	assert.True(t, tagFlag, "should have tag flag")
	assert.True(t, forceFlag, "should have force flag")
	assert.True(t, dryRunFlag, "should have dry-run flag")
}

func TestE2E_Help_JSONSchema(t *testing.T) {
	out := runExample(t, "deploy", "help", "--jsonschema")

	var schemas []struct {
		Name       string                 `json:"name"`
		Help       string                 `json:"help"`
		Properties map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"properties"`
		Required []string `json:"required"`
	}

	err := json.Unmarshal([]byte(out), &schemas)
	require.NoError(t, err, "should be valid JSON: %s", out)

	var stagingSchema *struct {
		Name       string                 `json:"name"`
		Help       string                 `json:"help"`
		Properties map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"properties"`
		Required []string `json:"required"`
	}
	for i := range schemas {
		if schemas[i].Name == "deploy staging" {
			stagingSchema = &schemas[i]
			break
		}
	}
	require.NotNil(t, stagingSchema, "should have deploy staging schema")

	assert.Contains(t, stagingSchema.Properties, "force")
	assert.Equal(t, "boolean", stagingSchema.Properties["force"].Type)
	assert.Equal(t, "Skip confirmation", stagingSchema.Properties["force"].Description)

	assert.Contains(t, stagingSchema.Properties, "dry-run")
	assert.Equal(t, "boolean", stagingSchema.Properties["dry-run"].Type)

	assert.Contains(t, stagingSchema.Required, "0")
}

func TestE2E_Help_BasicCommand(t *testing.T) {
	out := runExample(t, "basic", "help", "info")

	assert.Contains(t, out, "Print current global options")
	assert.Contains(t, out, "$ info")
}

func TestE2E_Help_JSON_BasicApp(t *testing.T) {
	out := runExample(t, "basic", "help", "--json")

	var commands []struct {
		Name string `json:"name"`
		Help string `json:"help"`
	}

	err := json.Unmarshal([]byte(out), &commands)
	require.NoError(t, err)

	names := make([]string, 0, len(commands))
	for _, cmd := range commands {
		names = append(names, cmd.Name)
	}
	assert.Contains(t, names, "help")
	assert.Contains(t, names, "version")
	assert.Contains(t, names, "info")
}

func TestE2E_Help_FlagsSingleCommand(t *testing.T) {
	out := runExample(t, "deploy", "help", "deploy", "--flags")

	assert.Contains(t, out, "staging")
	assert.Contains(t, out, "production")
	assert.Contains(t, out, "--force")
	assert.Contains(t, out, "--dry-run")
	assert.Contains(t, out, "Image tag to deploy")
}

func TestE2E_Help_FlagsViaGlobalHelp(t *testing.T) {
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
		{
			name:    "json with help flag",
			example: "deploy",
			args:    []string{"--json", "--help"},
			check: func(t *testing.T, out string) {
				var commands []json.RawMessage
				err := json.Unmarshal([]byte(out), &commands)
				require.NoError(t, err, "should output valid JSON")
			},
		},
		{
			name:    "jsonschema with help flag",
			example: "deploy",
			args:    []string{"--jsonschema", "--help"},
			check: func(t *testing.T, out string) {
				var schemas []json.RawMessage
				err := json.Unmarshal([]byte(out), &schemas)
				require.NoError(t, err, "should output valid JSON")
			},
		},
		{
			name:    "flags only without explicit help command",
			example: "basic",
			args:    []string{"--flags", "--help"},
			check: func(t *testing.T, out string) {
				assert.Contains(t, out, "--flags")
				assert.Contains(t, out, "--json")
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

func TestE2E_Help_Env(t *testing.T) {
	out := runExample(t, "full", "--env", "--help")

	assert.Contains(t, out, "[env: BUILD_OUTPUT]")
	assert.Contains(t, out, "[env: SERVE_PORT]")
	assert.Contains(t, out, "[env: CWD]")
	assert.Contains(t, out, "--output")
	assert.Contains(t, out, "--port")
}

func TestE2E_Help_EnvImpliesFlags(t *testing.T) {
	out := runExample(t, "full", "--env", "--help")

	assert.Contains(t, out, "--race")
	assert.Contains(t, out, "--tls")
	assert.Contains(t, out, "--dry-run")
}

func TestE2E_Help_Agent(t *testing.T) {
	out := runExample(t, "full", "--agent", "--help")

	assert.Contains(t, out, "\nbuild\n")
	assert.Contains(t, out, "Build the project")
	assert.Contains(t, out, "\nserve\n")
	assert.Contains(t, out, "Start the development server")
	assert.Contains(t, out, "\ndb migrate\n")
	assert.Contains(t, out, "Run database migrations")
	assert.Contains(t, out, "\ndb seed\n")
	assert.Contains(t, out, "\ndb reset\n")
	assert.Contains(t, out, "Global Options")

	assert.Contains(t, out, "--output, -o")
	assert.Contains(t, out, "BUILD_OUTPUT")
	assert.Contains(t, out, "string, required")
	assert.Contains(t, out, "--steps, -n")
}

func TestE2E_Help_AgentExamples(t *testing.T) {
	out := runExample(t, "full", "--agent", "--help")

	assert.Contains(t, out, "full build")
	assert.Contains(t, out, "full build -o ./dist --race")
	assert.Contains(t, out, "full serve -p 3000 --reload")
	assert.Contains(t, out, "full db migrate -n 3")
}

func TestE2E_Help_AgentFilter(t *testing.T) {
	out := runExample(t, "full", "--agent", "--filter=build,db seed", "--help")

	assert.Contains(t, out, "build\n")
	assert.Contains(t, out, "db seed\n")
	assert.NotContains(t, out, "serve\n  Start")
	assert.NotContains(t, out, "db migrate\n  Run")
	assert.NotContains(t, out, "db reset\n  Drop")
}

func TestE2E_Help_AgentFormatMd(t *testing.T) {
	out := runExample(t, "full", "--agent", "--format=md", "--help")

	assert.Contains(t, out, "## build")
	assert.Contains(t, out, "```")
	assert.Contains(t, out, "`--output`, `-o`")
}

func TestE2E_Help_AgentFormatPlain(t *testing.T) {
	out := runExample(t, "full", "--agent", "--format=plain", "--help")

	assert.NotContains(t, out, "```")
	assert.NotContains(t, out, "# full")
	assert.Contains(t, out, "full")
	assert.Contains(t, out, "build")
	assert.Contains(t, out, "--output")
}

func TestE2E_Help_AgentViaHelpCommand(t *testing.T) {
	out := runExample(t, "full", "help", "--agent")

	assert.Contains(t, out, "\nbuild\n")
	assert.Contains(t, out, "Build the project")
}
