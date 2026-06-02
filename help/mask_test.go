package help

import "testing"

func Test_maskValue(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"single rune shown as is", "5", "5"},
		{"short value reveals all but last", "ab", "a•"},
		{"reveals first three then masks", "8080", "808•"},
		{"longer value", "0.0.0.0", "0.0••••"},
		{"secret-like value mostly masked", "sk-live-abcdef", "sk-•••••••••••"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskValue(tt.in)
			if got != tt.want {
				t.Fatalf("maskValue(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func Test_maskValue_neverContainsFullSecret(t *testing.T) {
	secret := "super-secret-token-value"
	got := maskValue(secret)
	if got == secret {
		t.Fatalf("maskValue must not return the value unmasked: %q", got)
	}
	// the revealed prefix is short and the rest is masked.
	if len([]rune(got)) != len([]rune(secret)) {
		t.Fatalf("masked value should preserve length, got %d want %d", len([]rune(got)), len([]rune(secret)))
	}
}
