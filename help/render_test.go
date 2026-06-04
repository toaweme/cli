package help

import (
	"strings"
	"testing"
)

func Test_stripMarkdown(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "h1 heading loses marker",
			in:   "# Title",
			want: "Title",
		},
		{
			name: "bold and italics stripped",
			in:   "this is **bold** and *italic*",
			want: "this is bold and italic",
		},
		{
			name: "inline code backticks stripped",
			in:   "run `go test` now",
			want: "run go test now",
		},
		{
			name: "code fence markers removed but content kept",
			in:   "```\ncode line\n```",
			want: "code line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripMarkdown(tt.in); got != tt.want {
				t.Fatalf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func Test_stripMarkdown_NoAnsi(t *testing.T) {
	out := stripMarkdown("# Heading\n**bold** text")
	if strings.Contains(out, "\033[") {
		t.Fatalf("plain output should not contain ANSI escapes, got %q", out)
	}
}

func Test_prettyMarkdown_AddsAnsi(t *testing.T) {
	out := prettyMarkdown("# Heading")
	if !strings.Contains(out, ansiBoldWhite) {
		t.Fatalf("expected bold white heading, got %q", out)
	}
	if !strings.Contains(out, "Heading") {
		t.Fatalf("expected heading text, got %q", out)
	}
}

func Test_prettyInline(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		contains string
	}{
		{name: "inline code becomes green", in: "use `flag`", contains: ansiGreen},
		{name: "bold becomes bold", in: "**strong**", contains: ansiBold},
		{name: "italic becomes dim", in: "*soft*", contains: ansiDim},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := prettyInline(tt.in)
			if !strings.Contains(out, tt.contains) {
				t.Fatalf("expected %q to contain escape %q, got %q", tt.in, tt.contains, out)
			}
			if !strings.Contains(out, ansiReset) {
				t.Fatalf("expected reset escape, got %q", out)
			}
		})
	}
}

// Test_prettyInline_MultiByteRune guards against a byte-vs-rune offset bug: an
// emphasised span whose content starts with a multi-byte rune (e.g. the "…" a
// shortened path uses) must still consume its closing marker, not leak it into the
// styled text or pad the column past its measured visible width.
func Test_prettyInline_MultiByteRune(t *testing.T) {
	out := prettyInline("*…/toaweme/cli*")
	if strings.Contains(out, "*") {
		t.Fatalf("closing marker leaked into styled text, got %q", out)
	}
	if strings.ContainsRune(out, '\x00') {
		t.Fatalf("output should not contain a NUL byte, got %q", out)
	}
	want := ansiDim + "…/toaweme/cli" + ansiReset
	if out != want {
		t.Fatalf("want %q, got %q", want, out)
	}
}

func Test_resolveFormat(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "plain", want: "plain"},
		{in: "md", want: "md"},
		{in: "pretty", want: "pretty"},
		{in: "", want: "pretty"},
		{in: "json", want: "pretty"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := resolveFormat(tt.in); got != tt.want {
				t.Fatalf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func Test_renderMarkdown(t *testing.T) {
	md := "# Title\n**bold**"

	if got := renderMarkdown(md, "md"); got != md {
		t.Fatalf("md format should return raw markdown, got %q", got)
	}

	plain := renderMarkdown(md, "plain")
	if strings.Contains(plain, "**") || strings.Contains(plain, "#") {
		t.Fatalf("plain format should strip markers, got %q", plain)
	}
}

func Test_visibleWidth(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{in: "abc", want: 3},
		{in: "`abc`", want: 3},
		{in: "**abc**", want: 3},
		{in: "", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := visibleWidth(tt.in); got != tt.want {
				t.Fatalf("want %d, got %d", tt.want, got)
			}
		})
	}
}

func Test_splitTableCellsWithWidths(t *testing.T) {
	cells, widths := splitTableCellsWithWidths("| a | bb | ccc |")

	want := []string{"a", "bb", "ccc"}
	if len(cells) != len(want) {
		t.Fatalf("expected %d cells, got %d (%v)", len(want), len(cells), cells)
	}
	for i, c := range want {
		if cells[i] != c {
			t.Fatalf("cell %d: want %q, got %q", i, c, cells[i])
		}
	}
	if len(widths) != len(cells) {
		t.Fatalf("expected widths to match cells, got %d widths for %d cells", len(widths), len(cells))
	}
}

func Test_plainTableRow_SkipsHeaderAndSeparator(t *testing.T) {
	if got := plainTableRow("| Flag | Env |"); got != "" {
		t.Fatalf("expected header row to be skipped, got %q", got)
	}
	if got := plainTableRow("|------|-----|"); got != "" {
		t.Fatalf("expected separator row to be skipped, got %q", got)
	}
}

func Test_plainTableRow_RendersDataRow(t *testing.T) {
	got := plainTableRow("| `--flag` | `ENV` |")
	if !strings.Contains(got, "--flag") {
		t.Fatalf("expected flag in row, got %q", got)
	}
	if strings.Contains(got, "`") {
		t.Fatalf("expected backticks stripped, got %q", got)
	}
}
