package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadDotEnv(t *testing.T) {
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
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write .env file: %v", err)
			}

			for k, v := range tt.pre {
				t.Setenv(k, v)
			}

			err := LoadDotEnv(path)
			assertNoError(t, err)

			for k, want := range tt.want {
				assertEqual(t, want, os.Getenv(k))
			}
		})
	}
}

func Test_LoadDotEnv_MissingFile(t *testing.T) {
	err := LoadDotEnv("/nonexistent/.env")
	assertNoError(t, err)
}

func Test_GetDotEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "TEST_GET_A=one\nTEST_GET_B=\"two words\"\n# comment\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	got, err := GetDotEnv(path)
	assertNoError(t, err)
	assertEqual(t, "one", got["TEST_GET_A"])
	assertEqual(t, "two words", got["TEST_GET_B"])
	assertEqual(t, 2, len(got))
}

func Test_GetDotEnv_MissingFile(t *testing.T) {
	got, err := GetDotEnv("/nonexistent/.env")
	if !errors.Is(err, ErrDotenvNotFound) {
		t.Fatalf("expected ErrDotenvNotFound, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil map for missing file, got %v", got)
	}
}

func Test_GetDotEnvs_EarlierFileWins(t *testing.T) {
	dir := t.TempDir()

	first := filepath.Join(dir, "first.env")
	if err := os.WriteFile(first, []byte("SHARED=first\nONLY_FIRST=a\n"), 0o644); err != nil {
		t.Fatalf("failed to write first env file: %v", err)
	}

	second := filepath.Join(dir, "second.env")
	if err := os.WriteFile(second, []byte("SHARED=second\nONLY_SECOND=b\n"), 0o644); err != nil {
		t.Fatalf("failed to write second env file: %v", err)
	}

	got, err := GetDotEnvs(first, second)
	assertNoError(t, err)
	assertEqual(t, "first", got["SHARED"])
	assertEqual(t, "a", got["ONLY_FIRST"])
	assertEqual(t, "b", got["ONLY_SECOND"])
	assertEqual(t, 3, len(got))
}

func Test_GetDotEnvs_MissingFileErrors(t *testing.T) {
	got, err := GetDotEnvs("/nonexistent/a.env", "/nonexistent/b.env")
	if !errors.Is(err, ErrDotenvNotFound) {
		t.Fatalf("expected ErrDotenvNotFound, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil map on error, got %v", got)
	}
}

func Test_ReadDotEnv_NotFound(t *testing.T) {
	_, err := readDotEnv("/nonexistent/.env")
	if !errors.Is(err, ErrDotenvNotFound) {
		t.Fatalf("expected ErrDotenvNotFound, got %v", err)
	}
}
