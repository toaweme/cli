package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mapStructToOptions(t *testing.T) {
	type SimpleOpts struct {
		Name string `arg:"name"`
		Port int    `arg:"port"`
		Flag bool   `arg:"flag"`
	}

	tests := []struct {
		name     string
		vars     map[string]any
		expected SimpleOpts
		wantErr  bool
	}{
		{
			name:     "sets string field",
			vars:     map[string]any{"name": "test"},
			expected: SimpleOpts{Name: "test"},
		},
		{
			name:     "sets int field from string",
			vars:     map[string]any{"port": "8080"},
			expected: SimpleOpts{Port: 8080},
		},
		{
			name:     "sets bool field",
			vars:     map[string]any{"flag": true},
			expected: SimpleOpts{Flag: true},
		},
		{
			name:     "sets multiple fields",
			vars:     map[string]any{"name": "app", "port": "3000", "flag": true},
			expected: SimpleOpts{Name: "app", Port: 3000, Flag: true},
		},
		{
			name:     "ignores unknown keys",
			vars:     map[string]any{"name": "test", "unknown": "val"},
			expected: SimpleOpts{Name: "test"},
		},
		{
			name:     "empty vars",
			vars:     map[string]any{},
			expected: SimpleOpts{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &SimpleOpts{}
			err := mapStructToOptions(opts, tt.vars)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, *opts)
		})
	}
}
