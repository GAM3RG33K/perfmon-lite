package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/w1n/perfmon/internal/tui/styles"
)

// LogEntry represents a single log message.
type LogEntry struct {
	Timestamp time.Time
	Level     string // INFO, WARN, ERROR, DEBUG
	Message   string
}

// LogsView renders a scrollable log viewer.
type LogsView struct {
	Width    int
	Height   int
	Entries  []LogEntry
	ScrollPos int
	MaxEntries int
}

// NewLogsView creates a new logs view.
func NewLogsView(maxEntries int) *LogsView {
	return &LogsView{
		Width:      80,
		Height:     24,
		MaxEntries: maxEntries,
	}
}

// AddEntry adds a log entry to the view.
func (lv *LogsView) AddEntry(level, msg string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
	}

	lv.Entries = append(lv.Entries, entry)
	if len(lv.Entries) > lv.MaxEntries {
		lv.Entries = lv.Entries[len(lv.Entries)-lv.MaxEntries:]
	}

	// Auto-scroll to bottom
	lv.ScrollPos = len(lv.Entries)
}

// ScrollUp scrolls the log view up.
func (lv *LogsView) ScrollUp() {
	if lv.ScrollPos > 0 {
		lv.ScrollPos--
	}
}

// ScrollDown scrolls the log view down.
func (lv *LogsView) ScrollDown() {
	if lv.ScrollPos < len(lv.Entries) {
		lv.ScrollPos++
	}
}

// RenderRecent returns the last `n` log lines (for bottom console panel).
func (lv *LogsView) RenderRecent(n int) string {
	if len(lv.Entries) == 0 {
		return styles.LabelStyle.Render("  System ready. Waiting for telemetry...")
	}
	var b strings.Builder
	start := len(lv.Entries) - n
	if start < 0 {
		start = 0
	}
	for i := start; i < len(lv.Entries); i++ {
		entry := lv.Entries[i]
		var levelColor lipgloss.Color
		switch entry.Level {
		case "ERROR":
			levelColor = styles.Red
		case "WARN":
			levelColor = styles.Amber
		case "INFO":
			levelColor = styles.Green
		default:
			levelColor = styles.DimWhite
		}
		levelStr := lipgloss.NewStyle().Foreground(levelColor).Bold(true).Width(5).Render(entry.Level)
		b.WriteString(fmt.Sprintf(" %s %s\n", levelStr, entry.Message))
	}
	return b.String()
}

// Render draws the log viewer.
func (lv *LogsView) Render() string {
	if lv.Width < 40 {
		return "Terminal too narrow for logs"
	}

	var b strings.Builder

	b.WriteString(styles.HeaderStyle.Render("System Logs"))
	b.WriteString("\n\n")

	if len(lv.Entries) == 0 {
		b.WriteString(styles.LabelStyle.Render("No log entries yet.\n"))
		return b.String()
	}

	// Calculate visible range
	visibleLines := lv.Height - 5 // header + padding
	if visibleLines < 1 {
		visibleLines = 1
	}

	start := lv.ScrollPos - visibleLines
	if start < 0 {
		start = 0
	}
	end := lv.ScrollPos
	if end > len(lv.Entries) {
		end = len(lv.Entries)
	}

	for i := start; i < end; i++ {
		entry := lv.Entries[i]

		// Level color
		var levelColor lipgloss.Color
		switch entry.Level {
		case "ERROR":
			levelColor = styles.Red
		case "WARN":
			levelColor = styles.Amber
		case "INFO":
			levelColor = styles.Green
		case "DEBUG":
			levelColor = styles.DimWhite
		default:
			levelColor = styles.White
		}

		levelStr := lipgloss.NewStyle().
			Foreground(levelColor).
			Bold(true).
			Width(5).
			Render(entry.Level)

		timeStr := entry.Timestamp.Format("15:04:05")
		msg := entry.Message

		// Truncate message if too long
		maxMsgLen := lv.Width - 20
		if len(msg) > maxMsgLen {
			msg = msg[:maxMsgLen-3] + "..."
		}

		line := fmt.Sprintf(" %s [%s] %s", levelStr, timeStr, msg)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Scroll indicator
	if lv.ScrollPos < len(lv.Entries) {
		b.WriteString(styles.LabelStyle.Render(fmt.Sprintf("  ↓ %d more lines", len(lv.Entries)-lv.ScrollPos)))
	}

	return b.String()
}
