package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var globalFormat string

func SetFormat(f string) {
	globalFormat = f
}

func GetFormat(override string) string {
	if override != "" {
		return override
	}
	return globalFormat
}

func GetOutputFormat(override string) string {
	return GetFormat(override)
}

func IsJSON(override string) bool {
	f := GetFormat(override)
	return f == "json" || f == "jsonl"
}

func IsJSONL(override string) bool {
	return GetFormat(override) == "jsonl"
}

func PrintJSON(data any) {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON encoding error: %v\n", err)
		return
	}
	fmt.Println(string(out))
}

func PrintJSONL(items any) {
	enc := json.NewEncoder(os.Stdout)
	switch v := items.(type) {
	case []map[string]any:
		for _, item := range v {
			enc.Encode(item)
		}
	default:
		data, _ := json.Marshal(v)
		fmt.Println(string(data))
	}
}

func PrintTable(headers []string, rows [][]string) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateRows = false

	colConfigs := make([]table.ColumnConfig, len(headers))
	for i := range headers {
		colConfigs[i] = table.ColumnConfig{Number: i + 1, WidthMax: 60}
	}
	t.SetColumnConfigs(colConfigs)

	headerRow := make(table.Row, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	t.AppendHeader(headerRow)

	for _, row := range rows {
		r := make(table.Row, len(row))
		for i, cell := range row {
			r[i] = cell
		}
		t.AppendRow(r)
	}

	fmt.Println(t.Render())
}

func PrintTableRight(headers []string, rows [][]string, rightCols ...int) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateRows = false

	colConfigs := make([]table.ColumnConfig, len(headers))
	for i := range headers {
		colConfigs[i] = table.ColumnConfig{Number: i + 1, WidthMax: 60}
	}
	for _, col := range rightCols {
		if col > 0 && col <= len(colConfigs) {
			colConfigs[col-1].Align = text.AlignRight
		}
	}
	t.SetColumnConfigs(colConfigs)

	headerRow := make(table.Row, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	t.AppendHeader(headerRow)

	for _, row := range rows {
		r := make(table.Row, len(row))
		for i, cell := range row {
			r[i] = cell
		}
		t.AppendRow(r)
	}

	fmt.Println(t.Render())
}

const (
	ansiReset      = "\x1b[0m"
	ansiBold       = "\x1b[1m"
	ansiDim        = "\x1b[2m"
	ansiRed        = "\x1b[31m"
	ansiGreen      = "\x1b[32m"
	ansiYellow     = "\x1b[33m"
	ansiBlue       = "\x1b[34m"
	ansiCyan       = "\x1b[36m"
	ansiGray       = "\x1b[90m"
	ansiHideCursor = "\x1b[?25l"
	ansiShowCursor = "\x1b[?25h"
	saveCursor     = "\x1b7"
	restoreCursor  = "\x1b8"
	clearDown      = "\x1b[0J"
)

func Bold(s string) string { return ansiBold + s + ansiReset }
func Dim(s string) string { return ansiDim + s + ansiReset }
func Green(s string) string { return ansiGreen + s + ansiReset }
func Red(s string) string { return ansiRed + s + ansiReset }
func Yellow(s string) string { return ansiYellow + s + ansiReset }
func Cyan(s string) string { return ansiCyan + s + ansiReset }
func Blue(s string) string { return ansiBlue + s + ansiReset }
func Gray(s string) string { return ansiGray + s + ansiReset }

func Success(msg string) { fmt.Println(Green("  ✓ ") + msg) }
func Failed(msg string) { fmt.Println(Red("  ✗ ") + msg) }
func Running(msg string) { fmt.Println(Dim("  ▸ ") + msg) }

func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

func GetTermSize() (int, int, error) {
	return getTermSize()
}

func RenderMarkdown(content string) string {
	w, _, _ := GetTermSize()
	return RenderMarkdownWidth(content, w)
}

func RenderMarkdownWidth(content string, width int) string {
	if width <= 0 {
		width = 80
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}
	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return out
}

func ReadPipeOrFile(filePath string) (string, error) {
	fi, err := os.Stdin.Stat()
	if err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	if filePath != "" && filePath != "-" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", fmt.Errorf("no input provided")
}
