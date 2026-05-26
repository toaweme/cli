package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Load(t *testing.T) {
	tests := []struct {
		name    string
		content string
		pre     map[string]string
		want    map[string]string
	}{
		{
			name:    "basic key=value",
			content: "TEST_DOTENV_A=bar\nTEST_DOTENV_B=qux",
			want:    map[string]string{"TEST_DOTENV_A": "bar", "TEST_DOTENV_B": "qux"},
		},
		{
			name:    "quoted values",
			content: "TEST_DOTENV_C=\"hello world\"\nTEST_DOTENV_D='single'",
			want:    map[string]string{"TEST_DOTENV_C": "hello world", "TEST_DOTENV_D": "single"},
		},
		{
			name:    "skips comments and blanks",
			content: "# comment\n\nTEST_DOTENV_E=val\n  # indented comment",
			want:    map[string]string{"TEST_DOTENV_E": "val"},
		},
		{
			name:    "does not overwrite existing",
			content: "TEST_DOTENV_F=new",
			pre:     map[string]string{"TEST_DOTENV_F": "old"},
			want:    map[string]string{"TEST_DOTENV_F": "old"},
		},
		{
			name:    "trims whitespace",
			content: "  TEST_DOTENV_G  =  value  ",
			want:    map[string]string{"TEST_DOTENV_G": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k := range tt.want {
				t.Setenv(k, "")
				os.Unsetenv(k)
			}

			dir := t.TempDir()
			path := filepath.Join(dir, ".env")
			require.NoError(t, os.WriteFile(path, []byte(tt.content), 0o644))

			for k, v := range tt.pre {
				t.Setenv(k, v)
			}

			err := Load(path)
			require.NoError(t, err)

			for k, want := range tt.want {
				assert.Equal(t, want, os.Getenv(k))
			}
		})
	}
}

func Test_Load_MissingFile(t *testing.T) {
	err := Load("/nonexistent/.env")
	assert.NoError(t, err)
}
