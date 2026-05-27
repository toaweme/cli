package help

import (
	"fmt"
	"os"
	"strings"
)

const (
	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiDim       = "\033[2m"
	ansiGreen     = "\033[32m"
	ansiYellow    = "\033[33m"
	ansiBoldCyan  = "\033[1;36m"
	ansiBoldWhite = "\033[1;37m"
	ansiDimWhite  = "\033[2;37m"
	ansiDivider   = "\033[38;5;238m"
)

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// renderMarkdown dispatches to the right renderer based on format.
// "md" returns raw markdown, "plain" strips formatting, "pretty" adds ANSI colors.
func renderMarkdown(text string, format string) string {
	if format == "md" {
		return text
	}
	if format == "plain" {
		return stripMarkdown(text)
	}
	if !isTTY() {
		return text
	}
	return prettyMarkdown(text)
}

func stripMarkdown(text string) string {
	var lines []string
	inCode := false

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}

		if inCode {
			lines = append(lines, line)
			continue
		}

		if strings.HasPrefix(trimmed, "# ") {
			lines = append(lines, strings.TrimPrefix(trimmed, "# "))
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			lines = append(lines, "")
			lines = append(lines, strings.TrimPrefix(trimmed, "## "))
			continue
		}
		if strings.HasPrefix(trimmed, "### ") {
			lines = append(lines, strings.TrimPrefix(trimmed, "### "))
			continue
		}

		if strings.HasPrefix(trimmed, "|") {
			if row := plainTableRow(line); row != "" {
				lines = append(lines, row)
			}
			continue
		}

		line = strings.ReplaceAll(line, "**", "")
		line = stripBackticks(line)
		line = strings.ReplaceAll(line, "*", "")
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func stripBackticks(s string) string {
	return strings.ReplaceAll(s, "`", "")
}

// prettyMarkdown converts markdown to ANSI-styled terminal output.
// Headings become bold/cyan, inline code green, italics yellow,
// table rows aligned with padding, code blocks dimmed.
func prettyMarkdown(text string) string {
	var lines []string
	inCode := false

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		indent := len(line) - len(strings.TrimLeft(line, " "))
		pad := strings.Repeat(" ", indent)

		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}

		if inCode {
			if strings.HasPrefix(trimmed, "❯ ") {
				cmd := strings.TrimPrefix(trimmed, "❯ ")
				lines = append(lines, pad+ansiDimWhite+"❯"+ansiDim+" "+cmd+ansiReset)
			} else {
				lines = append(lines, pad+ansiDim+trimmed+ansiReset)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "# ") {
			title := strings.TrimPrefix(trimmed, "# ")
			lines = append(lines, ansiBoldWhite+title+ansiReset)
			continue
		}

		if strings.HasPrefix(trimmed, "## ") {
			title := strings.TrimPrefix(trimmed, "## ")
			if len(lines) > 0 {
				lines = append(lines, ansiDivider+strings.Repeat("─", 40)+ansiReset)
			}
			lines = append(lines, ansiBoldCyan+title+ansiReset)
			continue
		}

		if strings.HasPrefix(trimmed, "### ") {
			title := strings.TrimPrefix(trimmed, "### ")
			lines = append(lines, pad+ansiBold+title+ansiReset)
			continue
		}

		if strings.HasPrefix(trimmed, "- ") {
			content := trimmed[2:]
			content = prettyInline(content)
			lines = append(lines, pad+content)
			continue
		}

		if strings.HasPrefix(trimmed, "|") {
			if row := prettyTableRow(trimmed, pad); row != "" {
				lines = append(lines, row)
			}
			continue
		}

		if trimmed == "" {
			lines = append(lines, "")
			continue
		}

		lines = append(lines, pad+prettyInline(trimmed))
	}

	return strings.Join(lines, "\n")
}

func prettyInline(line string) string {
	var result strings.Builder
	runes := []rune(line)
	i := 0

	for i < len(runes) {
		if runes[i] == '`' {
			end := strings.IndexRune(string(runes[i+1:]), '`')
			if end >= 0 {
				code := string(runes[i+1 : i+1+end])
				result.WriteString(ansiGreen + code + ansiReset)
				i += end + 2
				continue
			}
		}

		if i < len(runes)-2 && runes[i] == '*' && runes[i+1] == '*' {
			end := strings.Index(string(runes[i+2:]), "**")
			if end >= 0 {
				bold := string(runes[i+2 : i+2+end])
				result.WriteString(ansiBold + bold + ansiReset)
				i += end + 4
				continue
			}
		}

		if runes[i] == '*' {
			end := strings.IndexRune(string(runes[i+1:]), '*')
			if end >= 0 {
				italic := string(runes[i+1 : i+1+end])
				result.WriteString(ansiDim + italic + ansiReset)
				i += end + 2
				continue
			}
		}

		result.WriteRune(runes[i])
		i++
	}

	return result.String()
}

// plainTableRow converts a markdown table row to plain text with aligned columns.
// Skips header and separator rows.
func plainTableRow(line string) string {
	stripped := strings.ReplaceAll(strings.TrimSpace(line), " ", "")
	stripped = strings.ReplaceAll(stripped, "|", "")
	stripped = strings.ReplaceAll(stripped, "-", "")
	if stripped == "" {
		return ""
	}

	indent := len(line) - len(strings.TrimLeft(line, " "))
	pad := strings.Repeat(" ", indent)
	cells, widths := splitTableCellsWithWidths(line)

	isHeader := true
	for _, c := range cells {
		if strings.ContainsAny(c, "`*") {
			isHeader = false
			break
		}
	}
	if isHeader {
		return ""
	}

	var clean []string
	for i, c := range cells {
		targetW := visibleWidth(c)
		if i < len(widths) {
			targetW = widths[i] - (len(c) - visibleWidth(c))
		}
		c = strings.ReplaceAll(c, "`", "")
		c = strings.ReplaceAll(c, "*", "")
		if targetW > len(c) {
			c += strings.Repeat(" ", targetW-len(c))
		}
		clean = append(clean, c)
	}
	return pad + strings.TrimRight(strings.Join(clean, "  "), " ")
}

// prettyTableRow converts a markdown table row to ANSI-styled text.
// Preserves column widths from the source, applies inline formatting.
// Skips header and separator rows.
func prettyTableRow(line, pad string) string {
	stripped := strings.ReplaceAll(line, " ", "")
	stripped = strings.ReplaceAll(stripped, "|", "")
	stripped = strings.ReplaceAll(stripped, "-", "")
	if stripped == "" {
		return ""
	}

	cells, widths := splitTableCellsWithWidths(line)
	isHeader := true
	for _, c := range cells {
		if strings.ContainsAny(c, "`*") {
			isHeader = false
			break
		}
	}

	if isHeader {
		envIdx := -1
		for i, c := range cells {
			if strings.TrimSpace(c) == "Env" {
				envIdx = i
				break
			}
		}
		if envIdx < 0 {
			return ""
		}
		offset := 0
		for i := 0; i < envIdx; i++ {
			if i < len(widths) {
				offset += widths[i]
			}
			offset += 2
		}
		return pad + strings.Repeat(" ", offset) + ansiDim + "ENV" + ansiReset
	}

	var styled []string
	for i, c := range cells {
		rendered := prettyInline(c)
		visibleLen := visibleWidth(c)
		if i < len(widths) && widths[i] > visibleLen {
			rendered += strings.Repeat(" ", widths[i]-visibleLen)
		}
		styled = append(styled, rendered)
	}
	return pad + strings.Join(styled, "  ")
}

func splitTableCellsWithWidths(line string) ([]string, []int) {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	var cells []string
	var widths []int
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		w := len(p)
		if len(p) >= 2 {
			w = len(p) - 2
		}
		cells = append(cells, trimmed)
		widths = append(widths, w)
	}
	return cells, widths
}

// visibleWidth returns the display width of s after stripping markdown formatting.
func visibleWidth(s string) int {
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "*", "")
	return len(s)
}

func resolveFormat(format string) string {
	switch format {
	case "plain", "md":
		return format
	default:
		return "pretty"
	}
}

func printRendered(md string, format string) {
	fmt.Print(renderMarkdown(md, resolveFormat(format)))
}
