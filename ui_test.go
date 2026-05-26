package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toaweme/structs"
)

func Test_findCommandByArgs(t *testing.T) {
	parent := NewMockCommand(nil)
	parent.Name("parent")
	child := NewMockCommand(nil)
	parent.Add("child", child)

	sibling := NewMockCommand(nil)
	sibling.Name("sibling")

	commands := []Command[any]{parent, sibling}

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "empty args",
			args:     []string{},
			expected: "",
		},
		{
			name:     "top level match",
			args:     []string{"parent"},
			expected: "parent",
		},
		{
			name:     "nested match",
			args:     []string{"parent", "child"},
			expected: "child",
		},
		{
			name:     "no match",
			args:     []string{"unknown"},
			expected: "",
		},
		{
			name:     "sibling match",
			args:     []string{"sibling"},
			expected: "sibling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findCommandByArgs(commands, tt.args)
			if tt.expected == "" {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.Name(""))
			}
		})
	}
}

func Test_getLongestName(t *testing.T) {
	tests := []struct {
		name     string
		commands []Command[any]
		expected int
	}{
		{
			name:     "empty",
			commands: []Command[any]{},
			expected: 0,
		},
		{
			name: "single command",
			commands: func() []Command[any] {
				cmd := NewMockCommand(nil)
				cmd.Name("help")
				return []Command[any]{cmd}
			}(),
			expected: 4,
		},
		{
			name: "sub command longer than parent",
			commands: func() []Command[any] {
				parent := NewMockCommand(nil)
				parent.Name("a")
				child := NewMockCommand(nil)
				parent.Add("longchild", child)
				return []Command[any]{parent}
			}(),
			expected: 11, // "a longchild"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getLongestName(tt.commands))
		})
	}
}

func Test_newHelpOption(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		short    string
		help     string
		expected helpOption
	}{
		{
			name:  "both arg and short",
			arg:   "cwd",
			short: "c",
			help:  "Current working directory",
			expected: helpOption{
				Args: "-c, --cwd",
				Help: "Current working directory",
			},
		},
		{
			name:  "long only",
			arg:   "verbosity",
			short: "",
			help:  "Verbosity level",
			expected: helpOption{
				Args: "--verbosity",
				Help: "Verbosity level",
			},
		},
		{
			name:  "short only",
			arg:   "",
			short: "v",
			help:  "Version",
			expected: helpOption{
				Args: "-v",
				Help: "Version",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newHelpOption(tt.arg, tt.short, tt.help)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_pad(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		indent   int
		expected string
	}{
		{
			name:     "no padding needed",
			text:     "help",
			indent:   4,
			expected: "",
		},
		{
			name:     "some padding",
			text:     "help",
			indent:   10,
			expected: "      ",
		},
		{
			name:     "empty text",
			text:     "",
			indent:   3,
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pad(tt.text, tt.indent))
		})
	}
}

func Test_helpOptions(t *testing.T) {
	lines, err := helpOptions(&GlobalOptions{})
	assert.NoError(t, err)
	assert.NotEmpty(t, lines)

	found := false
	for _, line := range lines {
		if strings.Contains(line, "--cwd") {
			found = true
			break
		}
	}
	assert.True(t, found, "should contain --cwd option")
}

func Test_displayAllCommandsHelp(t *testing.T) {
	cmd := NewMockCommand(nil)
	cmd.Name("deploy")
	cmd.HelpText = "Deploy the app"

	lines := displayAllCommandsHelp("myapp", []Command[any]{cmd}, HelpDisplayOptions{})
	assert.NotEmpty(t, lines)

	joined := strings.Join(lines, "\n")
	assert.Contains(t, joined, "Usage: myapp")
	assert.Contains(t, joined, "deploy")
	assert.Contains(t, joined, "Deploy the app")
	assert.Contains(t, joined, "--[opt]=<arg>")
}

func Test_displaySingleCommandHelp(t *testing.T) {
	cmd := NewMockCommand(nil)
	cmd.Name("build")
	cmd.HelpText = "Build the project"

	tests := []struct {
		name     string
		commands []Command[any]
		command  []string
		contains []string
		empty    bool
	}{
		{
			name:     "existing command",
			commands: []Command[any]{cmd},
			command:  []string{"build"},
			contains: []string{"Build the project", "$ build"},
		},
		{
			name:     "unknown command returns empty",
			commands: []Command[any]{cmd},
			command:  []string{"unknown"},
			empty:    true,
		},
		{
			name: "command with subcommands",
			commands: func() []Command[any] {
				parent := NewMockCommand(nil)
				parent.Name("deploy")
				parent.HelpText = "Deploy"
				sub := NewMockCommand(nil)
				sub.HelpText = "To staging"
				parent.Add("staging", sub)
				return []Command[any]{parent}
			}(),
			command:  []string{"deploy"},
			contains: []string{"Deploy", "staging", "To staging"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := displaySingleCommandHelp("myapp", tt.commands, tt.command, HelpDisplayOptions{})
			if tt.empty {
				assert.Empty(t, lines)
				return
			}
			joined := strings.Join(lines, "\n")
			for _, s := range tt.contains {
				assert.Contains(t, joined, s)
			}
		})
	}
}

func Test_printableFields(t *testing.T) {
	fields := []structs.Field{
		{
			Tags:   map[string]string{"arg": "cwd", "short": "c", "help": "Working dir"},
			Fields: nil,
		},
		{
			Tags:   map[string]string{"arg": "", "short": "", "help": ""},
			Fields: nil,
		},
	}

	lines := printableFields(fields)
	assert.Len(t, lines, 1)
	assert.Contains(t, lines[0], "--cwd")
	assert.Contains(t, lines[0], "-c")
	assert.Contains(t, lines[0], "Working dir")
}

func Test_maxLen(t *testing.T) {
	tests := []struct {
		name     string
		fields   []structs.Field
		expected int
	}{
		{
			name:     "empty fields",
			fields:   []structs.Field{},
			expected: 2,
		},
		{
			name: "single field with both tags",
			fields: []structs.Field{
				{Tags: map[string]string{"arg": "cwd", "short": "c", "help": ""}},
			},
			expected: 11, // len("-c, --cwd") + 2
		},
		{
			name: "long only field",
			fields: []structs.Field{
				{Tags: map[string]string{"arg": "verbosity", "short": "", "help": ""}},
			},
			expected: 13, // len("--verbosity") + 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, maxLen(tt.fields))
		})
	}
}
