package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
	"github.com/soda-data-inc/soda-cli/internal/ctx"
	"golang.org/x/term"
)

// Lipgloss styles.
var (
	Bold   = lipgloss.NewStyle().Bold(true)
	Dim    = lipgloss.NewStyle().Faint(true)
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00C850"))
	Red    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))
)

// IsTerminal returns true when stdout is a TTY.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// EffectiveFmt resolves "auto" to table or json based on TTY.
func EffectiveFmt(gctx *ctx.GlobalCtx) string {
	if gctx.Output != "auto" {
		return gctx.Output
	}
	if IsTerminal() {
		return "table"
	}
	return "json"
}

// FmtStatus applies color + symbol to known status values.
func FmtStatus(s string) string {
	switch s {
	case "passing":
		return Green.Render("✓") + " " + s
	case "failing":
		return Red.Render("✗") + " " + s
	case "error":
		return Yellow.Render("⚠") + " " + s
	case "alert":
		return Yellow.Render("⚠") + " " + s
	case "n/a":
		return Dim.Render("- n/a")
	case "running":
		return Dim.Render("⟳") + " " + s
	case "connected":
		return Green.Render("✓") + " " + s
	case "disconnected":
		return Red.Render("✗") + " " + s
	case "open":
		return Red.Render("●") + " " + s
	case "closed":
		return Dim.Render("●") + " " + s
	default:
		return s
	}
}

// FmtID returns a dimmed ID string.
func FmtID(id string) string {
	return Dim.Render(id)
}

// FmtTime returns a dimmed timestamp string.
func FmtTime(t string) string {
	return Dim.Render(t)
}

// Render outputs rows in the effective format.
// statusCols is a set of column names whose values should be colored via FmtStatus.
func Render(rows []map[string]string, cols []string, statusCols map[string]bool, gctx *ctx.GlobalCtx) {
	switch EffectiveFmt(gctx) {
	case "json":
		renderJSON(rows)
	case "csv":
		renderCSV(rows, cols)
	default:
		renderTable(rows, cols, statusCols, gctx)
	}
}

// RenderOne outputs a single item as key-value pairs or JSON.
func RenderOne(item map[string]string, keys []string, gctx *ctx.GlobalCtx) {
	if EffectiveFmt(gctx) == "json" {
		renderJSON([]map[string]string{item})
		return
	}
	for _, k := range keys {
		fmt.Printf("  %-20s %s\n", Bold.Render(k), item[k])
	}
}

// PrintSuccess prints a green success message to stdout.
func PrintSuccess(msg string, gctx *ctx.GlobalCtx) {
	if gctx.Quiet {
		return
	}
	fmt.Println(Green.Render("✓") + " " + msg)
}

// PrintInfo prints an informational message.
func PrintInfo(msg string, gctx *ctx.GlobalCtx) {
	if gctx.Quiet {
		return
	}
	fmt.Println(msg)
}

// ExitError is a typed error that carries an exit code.
type ExitError struct {
	Code int
	Msg  string
}

func (e *ExitError) Error() string { return e.Msg }

// Errorf returns an ExitError with the given code and formatted message.
func Errorf(code int, format string, args ...any) *ExitError {
	return &ExitError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// PrintError prints msg to stderr and exits with code.
func PrintError(msg string, code int) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Red.Render("[Error]"), msg)
	os.Exit(code)
}

// --- internal renderers ---

func renderJSON(rows []map[string]string) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(rows)
}

func renderCSV(rows []map[string]string, cols []string) {
	w := csv.NewWriter(os.Stdout)
	_ = w.Write(cols)
	for _, row := range rows {
		record := make([]string, len(cols))
		for i, c := range cols {
			record[i] = row[c]
		}
		_ = w.Write(record)
	}
	w.Flush()
}

func renderTable(rows []map[string]string, cols []string, statusCols map[string]bool, gctx *ctx.GlobalCtx) {
	if len(rows) == 0 {
		fmt.Println(Dim.Render("No results."))
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetColumnSeparator("  ")
	table.SetHeaderLine(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	// Bold uppercase headers.
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = Bold.Render(strings.ToUpper(c))
	}
	table.SetHeader(headers)

	for _, row := range rows {
		record := make([]string, len(cols))
		for i, c := range cols {
			val := row[c]
			if statusCols[c] {
				val = FmtStatus(val)
			}
			// Dim IDs and timestamps heuristically.
			if c == "id" || c == "created" || c == "date" || c == "updated" {
				val = Dim.Render(val)
			}
			record[i] = val
		}
		table.Append(record)
	}

	table.Render()
}
